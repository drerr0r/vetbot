// internal/imports/csv_importer.go
package imports

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/drerr0r/vetbot/internal/database"
	"github.com/drerr0r/vetbot/internal/models"
	"github.com/xuri/excelize/v2"
)

type CSVImporter struct {
	db *database.Database
}

func NewCSVImporter(db *database.Database) *CSVImporter {
	return &CSVImporter{db: db}
}

// ImportVeterinarians импортирует врачей из CSV/Excel с поддержкой городов, клиник и расписания
func (i *CSVImporter) ImportVeterinarians(file io.Reader, filename string) (*models.ImportResult, error) {
	records, err := i.readFile(file, filename)
	if err != nil {
		return nil, err
	}

	result := &models.ImportResult{
		TotalRows: len(records) - 1, // минус заголовок
		Errors:    []models.ImportError{},
	}

	// Предзагружаем справочники для быстрого поиска
	cities, err := i.db.GetAllCities()
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки городов: %w", err)
	}

	specializations, err := i.db.GetAllSpecializations()
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки специализаций: %w", err)
	}

	clinics, err := i.db.GetAllClinics()
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки клиник: %w", err)
	}

	cityMap := make(map[string]int)
	for _, city := range cities {
		cityMap[strings.ToLower(city.Name)] = city.ID
	}

	specMap := make(map[string]int)
	for _, spec := range specializations {
		specMap[strings.ToLower(spec.Name)] = spec.ID
	}

	clinicMap := make(map[string]int)
	for _, clinic := range clinics {
		clinicMap[strings.ToLower(clinic.Name)] = clinic.ID
	}

	for idx, record := range records {
		if idx == 0 {
			continue // Пропускаем заголовок
		}

		if len(record) < 7 {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "all",
				Message:   fmt.Sprintf("Недостаточно колонок (требуется минимум 7, получено %d)", len(record)),
			})
			continue
		}

		// Парсим данные врача
		vet := &models.Veterinarian{
			FirstName: strings.TrimSpace(record[0]),
			LastName:  strings.TrimSpace(record[1]),
			Phone:     strings.TrimSpace(record[2]),
			Email:     i.parseNullString(record[3]),
			IsActive:  true,
		}

		// Опыт работы
		if record[4] != "" {
			if exp, err := strconv.ParseInt(strings.TrimSpace(record[4]), 10, 64); err == nil {
				vet.ExperienceYears = sql.NullInt64{Int64: exp, Valid: true}
			}
		}

		// Описание
		if len(record) > 5 {
			vet.Description = i.parseNullString(record[5])
		}

		// Город
		if len(record) > 6 && record[6] != "" {
			cityName := strings.TrimSpace(record[6])
			if id, exists := cityMap[strings.ToLower(cityName)]; exists {
				vet.CityID = sql.NullInt64{Int64: int64(id), Valid: true}
			} else {
				result.ErrorCount++
				result.Errors = append(result.Errors, models.ImportError{
					RowNumber: idx + 1,
					Field:     "city",
					Message:   fmt.Sprintf("Город '%s' не найден в базе", cityName),
				})
				continue
			}
		}

		// Добавляем врача в базу со всеми связями
		err := i.addVeterinarianWithRelations(vet, record, specMap, clinicMap, idx+1)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "database",
				Message:   fmt.Sprintf("Ошибка сохранения: %v", err),
			})
		} else {
			result.SuccessCount++
		}
	}

	return result, nil
}

// addVeterinarianWithRelations добавляет врача со специализациями, клиниками и расписанием
func (i *CSVImporter) addVeterinarianWithRelations(vet *models.Veterinarian, record []string, specMap, clinicMap map[string]int, rowNum int) error {
	tx, err := i.db.GetDB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Добавляем врача
	query := `INSERT INTO veterinarians (first_name, last_name, phone, email, experience_years, description, city_id, is_active) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	err = tx.QueryRow(query, vet.FirstName, vet.LastName, vet.Phone, vet.Email,
		vet.ExperienceYears, vet.Description, vet.CityID, vet.IsActive).Scan(&vet.ID)
	if err != nil {
		return fmt.Errorf("ошибка добавления врача: %w", err)
	}

	// Обрабатываем специализации (колонка 7)
	if len(record) > 7 && record[7] != "" {
		specNames := strings.Split(record[7], ",")
		for _, specName := range specNames {
			specName = strings.TrimSpace(specName)
			if specName == "" {
				continue
			}

			specID, exists := specMap[strings.ToLower(specName)]
			if !exists {
				log.Printf("Специализация '%s' не найдена в строке %d", specName, rowNum)
				continue
			}

			_, err = tx.Exec(
				"INSERT INTO vet_specializations (vet_id, specialization_id) VALUES ($1, $2)",
				vet.ID, specID,
			)
			if err != nil {
				log.Printf("Ошибка добавления специализации %s для врача %d: %v", specName, vet.ID, err)
			}
		}
	}

	// Обрабатываем клиники и расписание (колонка 8)
	if len(record) > 8 && record[8] != "" {
		clinicSchedules := strings.Split(record[8], ";")
		for _, clinicSchedule := range clinicSchedules {
			parts := strings.Split(clinicSchedule, ":")
			if len(parts) < 2 {
				continue
			}

			clinicName := strings.TrimSpace(parts[0])
			scheduleStr := strings.TrimSpace(parts[1])

			clinicID, exists := clinicMap[strings.ToLower(clinicName)]
			if !exists {
				log.Printf("Клиника '%s' не найдена в строке %d", clinicName, rowNum)
				continue
			}

			// Парсим расписание
			schedules := i.parseSchedule(scheduleStr, vet.ID, clinicID)
			for _, schedule := range schedules {
				_, err = tx.Exec(
					`INSERT INTO schedules (vet_id, clinic_id, day_of_week, start_time, end_time, is_available) 
					 VALUES ($1, $2, $3, $4, $5, $6)`,
					schedule.VetID, schedule.ClinicID, schedule.DayOfWeek, schedule.StartTime, schedule.EndTime, schedule.IsAvailable,
				)
				if err != nil {
					log.Printf("Ошибка добавления расписания для врача %d в клинике %d: %v", vet.ID, clinicID, err)
				}
			}
		}
	}

	return tx.Commit()
}

// parseSchedule парсит строку расписания формата "Пн:9-18,Ср:9-18,Пт:14-20"
func (i *CSVImporter) parseSchedule(scheduleStr string, vetID, clinicID int) []models.Schedule {
	var schedules []models.Schedule

	dayMap := map[string]int{
		"пн": 1, "пон": 1, "понедельник": 1,
		"вт": 2, "вто": 2, "вторник": 2,
		"ср": 3, "сре": 3, "среда": 3,
		"чт": 4, "чет": 4, "четверг": 4,
		"пт": 5, "пят": 5, "пятница": 5,
		"сб": 6, "суб": 6, "суббота": 6,
		"вс": 7, "вос": 7, "воскресенье": 7,
	}

	days := strings.Split(scheduleStr, ",")
	for _, day := range days {
		parts := strings.Split(day, ":")
		if len(parts) != 2 {
			continue
		}

		dayName := strings.ToLower(strings.TrimSpace(parts[0]))
		timeRange := strings.TrimSpace(parts[1])

		dayOfWeek, exists := dayMap[dayName]
		if !exists {
			continue
		}

		timeParts := strings.Split(timeRange, "-")
		if len(timeParts) != 2 {
			continue
		}

		startTime, endTime := strings.TrimSpace(timeParts[0]), strings.TrimSpace(timeParts[1])

		// Добавляем :00 если время указано без минут
		if !strings.Contains(startTime, ":") {
			startTime += ":00"
		}
		if !strings.Contains(endTime, ":") {
			endTime += ":00"
		}

		schedule := models.Schedule{
			VetID:       vetID,
			ClinicID:    clinicID,
			DayOfWeek:   dayOfWeek,
			StartTime:   startTime,
			EndTime:     endTime,
			IsAvailable: true,
			CreatedAt:   time.Now(),
		}
		schedules = append(schedules, schedule)
	}

	return schedules
}

// Вспомогательные методы (остаются без изменений)
func (i *CSVImporter) readFile(file io.Reader, filename string) ([][]string, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		return i.readExcel(file)
	}
	return i.readCSV(file)
}

func (i *CSVImporter) readCSV(file io.Reader) ([][]string, error) {
	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.LazyQuotes = true
	return reader.ReadAll()
}

func (i *CSVImporter) readExcel(file io.Reader) ([][]string, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("нет листов в файле")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (i *CSVImporter) parseNullString(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" || s == "NULL" || s == "null" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
