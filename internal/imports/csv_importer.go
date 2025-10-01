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
func (i *CSVImporter) ImportVeterinarians(file io.Reader, filename string, InfoLog, ErrorLog *log.Logger) (*models.ImportResult, error) {
	InfoLog.Printf("🚀 Начало импорта файла: %s", filename)

	records, err := i.readFile(file, filename)
	if err != nil {
		ErrorLog.Printf("❌ Ошибка чтения файла: %v", err)
		return nil, err
	}

	InfoLog.Printf("📊 Прочитано строк: %d", len(records))

	if len(records) > 0 {
		InfoLog.Printf("📋 Заголовки: %v", records[0])
	} else {
		ErrorLog.Printf("❌ Файл %s пустой", filename)
		return nil, fmt.Errorf("файл пустой")
	}

	result := &models.ImportResult{
		TotalRows: len(records) - 1, // минус заголовок
		Errors:    []models.ImportError{},
	}

	// Предзагружаем справочники для быстрого поиска
	InfoLog.Printf("🔍 Загрузка справочников...")
	cities, err := i.db.GetAllCities()
	if err != nil {
		ErrorLog.Printf("❌ Ошибка загрузки городов: %v", err)
		return nil, fmt.Errorf("ошибка загрузки городов: %w", err)
	}

	specializations, err := i.db.GetAllSpecializations()
	if err != nil {
		ErrorLog.Printf("❌ Ошибка загрузки специализаций: %v", err)
		return nil, fmt.Errorf("ошибка загрузки специализаций: %w", err)
	}

	clinics, err := i.db.GetAllClinics()
	if err != nil {
		ErrorLog.Printf("❌ Ошибка загрузки клиник: %v", err)
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

	InfoLog.Printf("✅ Справочники загружены: %d городов, %d специализаций, %d клиник", len(cities), len(specializations), len(clinics))

	InfoLog.Printf("🔍 Отладка - загруженные клиники:")
	for _, clinic := range clinics {
		district := "не указан"
		if clinic.District.Valid {
			district = clinic.District.String
		}

		metro := "не указана"
		if clinic.MetroStation.Valid {
			metro = clinic.MetroStation.String
		}

		InfoLog.Printf("   Клиника: %s (ID: %d, CityID: %v, District: %s, Metro: %s)",
			clinic.Name, clinic.ID, clinic.CityID, district, metro)
	}

	InfoLog.Printf("🔍 Отладка - загруженные города:")
	for _, city := range cities {
		InfoLog.Printf("   Город: %s (ID: %d, Region: %s)", city.Name, city.ID, city.Region)
	}

	for idx, record := range records {
		if idx == 0 {
			InfoLog.Printf("🔤 Пропускаем заголовок: %v", record)
			continue // Пропускаем заголовок
		}

		if len(record) < 7 {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "all",
				Message:   fmt.Sprintf("Недостаточно колонок (требуется минимум 7, получено %d)", len(record)),
			})
			ErrorLog.Printf("❌ Строка %d: недостаточно колонок (%d вместо 7)", idx+1, len(record))
			continue
		}

		InfoLog.Printf("📝 Обрабатываем строку %d: %v", idx+1, record)

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
				InfoLog.Printf("💼 Строка %d: опыт работы %d лет", idx+1, exp)
			} else {
				ErrorLog.Printf("⚠️ Строка %d: неверный формат опыта работы '%s'", idx+1, record[4])
			}
		}

		// Описание
		if len(record) > 5 {
			vet.Description = i.parseNullString(record[5])
			if vet.Description.Valid {
				InfoLog.Printf("📄 Строка %d: описание добавлено", idx+1)
			}
		}

		// Город
		if len(record) > 6 && record[6] != "" {
			cityName := strings.TrimSpace(record[6])
			if id, exists := cityMap[strings.ToLower(cityName)]; exists {
				vet.CityID = sql.NullInt64{Int64: int64(id), Valid: true}
				InfoLog.Printf("🏙️ Строка %d: город '%s' найден (ID: %d)", idx+1, cityName, id)
			} else {
				result.ErrorCount++
				result.Errors = append(result.Errors, models.ImportError{
					RowNumber: idx + 1,
					Field:     "city",
					Message:   fmt.Sprintf("Город '%s' не найден в базе", cityName),
				})
				ErrorLog.Printf("❌ Строка %d: город '%s' не найден в базе", idx+1, cityName)
				continue
			}
		} else {
			ErrorLog.Printf("⚠️ Строка %d: город не указан", idx+1)
		}

		// Добавляем врача в базу со всеми связями
		err := i.addVeterinarianWithRelations(vet, record, specMap, clinicMap, idx+1, InfoLog, ErrorLog)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "database",
				Message:   fmt.Sprintf("Ошибка сохранения: %v", err),
			})
			ErrorLog.Printf("❌ Строка %d: ошибка сохранения врача: %v", idx+1, err)
		} else {
			result.SuccessCount++
			InfoLog.Printf("✅ Строка %d: врач %s %s успешно добавлен (ID: %d)", idx+1, vet.FirstName, vet.LastName, vet.ID)
		}
	}

	InfoLog.Printf("🎯 Импорт завершен. Успешно: %d, Ошибок: %d, Всего строк: %d",
		result.SuccessCount, result.ErrorCount, result.TotalRows)
	return result, nil
}

// addVeterinarianWithRelations добавляет врача со специализациями, клиниками и расписанием
func (i *CSVImporter) addVeterinarianWithRelations(vet *models.Veterinarian, record []string, specMap, clinicMap map[string]int, rowNum int, InfoLog, ErrorLog *log.Logger) error {
	tx, err := i.db.GetDB().Begin()
	if err != nil {
		ErrorLog.Printf("❌ Строка %d: ошибка начала транзакции: %v", rowNum, err)
		return err
	}
	defer tx.Rollback()

	// Добавляем врача
	query := `INSERT INTO veterinarians (first_name, last_name, phone, email, experience_years, description, city_id, is_active) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	err = tx.QueryRow(query, vet.FirstName, vet.LastName, vet.Phone, vet.Email,
		vet.ExperienceYears, vet.Description, vet.CityID, vet.IsActive).Scan(&vet.ID)
	if err != nil {
		ErrorLog.Printf("❌ Строка %d: ошибка добавления врача в БД: %v", rowNum, err)
		return fmt.Errorf("ошибка добавления врача: %w", err)
	}

	InfoLog.Printf("👨‍⚕️ Строка %d: врач добавлен в БД с ID: %d", rowNum, vet.ID)

	// Обрабатываем специализации (колонка 7)
	if len(record) > 7 && record[7] != "" {
		specNames := strings.Split(record[7], ",")
		specCount := 0
		for _, specName := range specNames {
			specName = strings.TrimSpace(specName)
			if specName == "" {
				continue
			}

			specID, exists := specMap[strings.ToLower(specName)]
			if !exists {
				ErrorLog.Printf("⚠️ Строка %d: специализация '%s' не найдена", rowNum, specName)
				continue
			}

			_, err = tx.Exec(
				"INSERT INTO vet_specializations (vet_id, specialization_id) VALUES ($1, $2)",
				vet.ID, specID,
			)
			if err != nil {
				ErrorLog.Printf("⚠️ Строка %d: ошибка добавления специализации '%s': %v", rowNum, specName, err)
			} else {
				specCount++
				InfoLog.Printf("🎯 Строка %d: добавлена специализация '%s'", rowNum, specName)
			}
		}
		InfoLog.Printf("✅ Строка %d: добавлено %d специализаций", rowNum, specCount)
	}

	// Обрабатываем клиники и расписание (колонка 8) - С ОБРАБОТКОЙ ОШИБОК
	if len(record) > 8 && record[8] != "" {
		clinicSchedules := strings.Split(record[8], ";")
		scheduleCount := 0
		scheduleErrors := 0

		for _, clinicSchedule := range clinicSchedules {
			parts := strings.Split(clinicSchedule, ":")
			if len(parts) < 2 {
				ErrorLog.Printf("⚠️ Строка %d: неверный формат расписания '%s'", rowNum, clinicSchedule)
				scheduleErrors++
				continue
			}

			clinicName := strings.TrimSpace(parts[0])
			scheduleStr := strings.TrimSpace(parts[1])

			clinicID, exists := clinicMap[strings.ToLower(clinicName)]
			if !exists {
				ErrorLog.Printf("⚠️ Строка %d: клиника '%s' не найдена", rowNum, clinicName)
				scheduleErrors++
				continue
			}

			InfoLog.Printf("🏥 Строка %d: обработка клиники '%s' (ID: %d)", rowNum, clinicName, clinicID)

			// Парсим расписание
			schedules := i.parseSchedule(scheduleStr, vet.ID, clinicID)
			for _, schedule := range schedules {
				_, err = tx.Exec(
					`INSERT INTO schedules (vet_id, clinic_id, day_of_week, start_time, end_time, is_available) 
                     VALUES ($1, $2, $3, $4, $5, $6)`,
					schedule.VetID, schedule.ClinicID, schedule.DayOfWeek, schedule.StartTime, schedule.EndTime, schedule.IsAvailable,
				)
				if err != nil {
					ErrorLog.Printf("⚠️ Строка %d: ошибка добавления расписания: %v", rowNum, err)
					scheduleErrors++
				} else {
					scheduleCount++
				}
			}
		}
		InfoLog.Printf("📅 Строка %d: добавлено %d записей расписания, ошибок: %d", rowNum, scheduleCount, scheduleErrors)

		// НЕ ПРЕРЫВАЕМ ИМПОРТ ИЗ-ЗА ОШИБОК РАСПИСАНИЯ
		if scheduleErrors > 0 {
			InfoLog.Printf("⚠️ Строка %d: есть ошибки расписания, но врач добавлен", rowNum)
		}
	}

	err = tx.Commit()
	if err != nil {
		ErrorLog.Printf("❌ Строка %d: ошибка коммита транзакции: %v", rowNum, err)
		return err
	}

	InfoLog.Printf("💾 Строка %d: транзакция успешно завершена", rowNum)
	return nil
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
