package imports

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

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

// ImportVeterinarians импортирует врачей из CSV/Excel
func (i *CSVImporter) ImportVeterinarians(file io.Reader, filename string) (*models.ImportResult, error) {
	records, err := i.readFile(file, filename)
	if err != nil {
		return nil, err
	}

	result := &models.ImportResult{
		TotalRows: len(records) - 1, // минус заголовок
		Errors:    []models.ImportError{},
	}

	// Предзагружаем специализации для быстрого поиска
	specializations, err := i.db.GetAllSpecializations()
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки специализаций: %w", err)
	}

	specMap := make(map[string]int)
	for _, spec := range specializations {
		specMap[strings.ToLower(spec.Name)] = spec.ID
	}

	for idx, record := range records {
		if idx == 0 {
			continue // Пропускаем заголовок
		}

		if len(record) < 5 {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "all",
				Message:   "Недостаточно колонок (требуется минимум 5)",
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

		// Опыт работы (опционально)
		if len(record) > 4 && record[4] != "" {
			if exp, err := strconv.ParseInt(strings.TrimSpace(record[4]), 10, 64); err == nil {
				vet.ExperienceYears = sql.NullInt64{Int64: exp, Valid: true}
			}
		}

		// Описание (опционально)
		if len(record) > 5 {
			vet.Description = i.parseNullString(record[5])
		}

		// Добавляем врача в базу
		err := i.addVeterinarianWithSpecializations(vet, record, specMap, idx+1)
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

// addVeterinarianWithSpecializations добавляет врача со специализациями
func (i *CSVImporter) addVeterinarianWithSpecializations(vet *models.Veterinarian, record []string, specMap map[string]int, rowNum int) error {
	// Начинаем транзакцию
	tx, err := i.db.GetDB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Добавляем врача
	query := `INSERT INTO veterinarians (first_name, last_name, phone, email, experience_years, description, is_active) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err = tx.QueryRow(query, vet.FirstName, vet.LastName, vet.Phone, vet.Email,
		vet.ExperienceYears, vet.Description, vet.IsActive).Scan(&vet.ID)
	if err != nil {
		return err
	}

	// Обрабатываем специализации (колонка 6 и далее)
	if len(record) > 6 {
		specNames := strings.Split(record[6], ",")
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

			// Добавляем связь врач-специализация
			_, err = tx.Exec(
				"INSERT INTO vet_specializations (vet_id, specialization_id) VALUES ($1, $2)",
				vet.ID, specID,
			)
			if err != nil {
				log.Printf("Ошибка добавления специализации %d для врача %d: %v", specID, vet.ID, err)
			}
		}
	}

	return tx.Commit()
}

// Вспомогательные методы для чтения файлов
func (i *CSVImporter) readFile(file io.Reader, filename string) ([][]string, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		return i.readExcel(file)
	}
	return i.readCSV(file)
}

func (i *CSVImporter) readCSV(file io.Reader) ([][]string, error) {
	reader := csv.NewReader(file)
	reader.Comma = ';' // Используем точку с запятой как разделитель
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
