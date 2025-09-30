package imports

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
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

// ImportCities импортирует города из CSV/Excel
func (i *CSVImporter) ImportCities(file io.Reader, filename string) (*models.ImportResult, error) {
	records, err := i.readFile(file, filename)
	if err != nil {
		return nil, err
	}

	result := &models.ImportResult{
		TotalRows: len(records) - 1,
		Errors:    []models.ImportError{},
	}

	for idx, record := range records {
		if idx == 0 {
			continue
		}

		if len(record) < 2 {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "all",
				Message:   "Недостаточно колонок",
			})
			continue
		}

		city := &models.City{
			Name:   strings.TrimSpace(record[0]),
			Region: strings.TrimSpace(record[1]),
		}

		if err := i.db.CreateCity(city); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "name",
				Message:   fmt.Sprintf("Ошибка сохранения: %v", err),
			})
		} else {
			result.SuccessCount++
		}
	}

	return result, nil
}

// ImportClinics импортирует клиники из CSV/Excel
func (i *CSVImporter) ImportClinics(file io.Reader, filename string) (*models.ImportResult, error) {
	records, err := i.readFile(file, filename)
	if err != nil {
		return nil, err
	}

	result := &models.ImportResult{
		TotalRows: len(records) - 1,
		Errors:    []models.ImportError{},
	}

	// Предзагружаем города для быстрого поиска
	cities, err := i.db.GetAllCities()
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки городов: %w", err)
	}

	cityMap := make(map[string]int)
	for _, city := range cities {
		cityMap[strings.ToLower(city.Name)] = city.ID
	}

	for idx, record := range records {
		if idx == 0 {
			continue
		}

		if len(record) < 7 {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "all",
				Message:   "Недостаточно колонок (требуется 7)",
			})
			continue
		}

		// Парсим данные
		cityName := strings.TrimSpace(record[1])
		cityID, exists := cityMap[strings.ToLower(cityName)]
		if !exists {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "city",
				Message:   fmt.Sprintf("Город '%s' не найден в базе", cityName),
			})
			continue
		}

		clinic := &models.Clinic{
			Name:         strings.TrimSpace(record[0]),
			Address:      strings.TrimSpace(record[2]),
			Phone:        i.parseNullString(record[3]),
			WorkingHours: i.parseNullString(record[4]),
			District:     i.parseNullString(record[5]),
			MetroStation: i.parseNullString(record[6]),
			IsActive:     true,
			CityID:       i.parseNullInt64(cityID),
		}

		if err := i.db.CreateClinicWithCity(clinic); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, models.ImportError{
				RowNumber: idx + 1,
				Field:     "name",
				Message:   fmt.Sprintf("Ошибка сохранения: %v", err),
			})
		} else {
			result.SuccessCount++
		}
	}

	return result, nil
}

// Вспомогательные методы
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

func (i *CSVImporter) parseNullInt64(value int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(value), Valid: true}
}
