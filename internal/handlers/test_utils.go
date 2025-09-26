package handlers

import (
	"strconv"
	"strings"
)

// TestUtils содержит вспомогательные функции для тестов
type TestUtils struct{}

// ParseTestCallback имитирует парсинг callback данных для тестов
func (tu *TestUtils) ParseTestCallback(data string) (string, int, error) {
	switch {
	case strings.HasPrefix(data, "search_spec_"):
		idStr := strings.TrimPrefix(data, "search_spec_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return "", 0, err
		}
		return "specialization", id, nil

	case strings.HasPrefix(data, "search_clinic_"):
		idStr := strings.TrimPrefix(data, "search_clinic_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return "", 0, err
		}
		return "clinic", id, nil

	case strings.HasPrefix(data, "search_day_"):
		idStr := strings.TrimPrefix(data, "search_day_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return "", 0, err
		}
		return "day", id, nil

	default:
		return "", 0, strconv.ErrSyntax
	}
}
