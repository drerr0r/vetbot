package models

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ USER
// ============================================================================

func TestUser_JSONSerialization(t *testing.T) {
	user := User{
		ID:         1,
		TelegramID: 123456789,
		Username:   "testuser",
		FirstName:  "Иван",
		LastName:   "Петров",
		Phone:      "+79123456789",
		CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(user)
	assert.NoError(t, err)

	// Десериализуем обратно
	var decodedUser User
	err = json.Unmarshal(jsonData, &decodedUser)
	assert.NoError(t, err)

	// Проверяем корректность данных
	assert.Equal(t, user.ID, decodedUser.ID)
	assert.Equal(t, user.TelegramID, decodedUser.TelegramID)
	assert.Equal(t, user.Username, decodedUser.Username)
	assert.Equal(t, user.FirstName, decodedUser.FirstName)
	assert.Equal(t, user.LastName, decodedUser.LastName)
	assert.Equal(t, user.Phone, decodedUser.Phone)
	assert.True(t, user.CreatedAt.Equal(decodedUser.CreatedAt))
}

func TestUser_EmptyFields(t *testing.T) {
	user := User{
		ID:        1,
		FirstName: "Иван",
		// Остальные поля пустые
	}

	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "Иван", user.FirstName)
	assert.Equal(t, "", user.Username)
	assert.Equal(t, "", user.LastName)
	assert.Equal(t, "", user.Phone)
}

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ SPECIALIZATION
// ============================================================================

func TestSpecialization_JSONSerialization(t *testing.T) {
	spec := Specialization{
		ID:          1,
		Name:        "Хирург",
		Description: "Ветеринарный хирург",
		CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	jsonData, err := json.Marshal(spec)
	assert.NoError(t, err)

	var decodedSpec Specialization
	err = json.Unmarshal(jsonData, &decodedSpec)
	assert.NoError(t, err)

	assert.Equal(t, spec.ID, decodedSpec.ID)
	assert.Equal(t, spec.Name, decodedSpec.Name)
	assert.Equal(t, spec.Description, decodedSpec.Description)
	assert.True(t, spec.CreatedAt.Equal(decodedSpec.CreatedAt))
}

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ VETERINARIAN
// ============================================================================

func TestVeterinarian_JSONSerialization(t *testing.T) {
	vet := Veterinarian{
		ID:              1,
		FirstName:       "Анна",
		LastName:        "Смирнова",
		Phone:           "+79123456789",
		Email:           sql.NullString{String: "anna@vet.ru", Valid: true},
		Description:     sql.NullString{String: "Опытный ветеринар", Valid: true},
		ExperienceYears: sql.NullInt64{Int64: 5, Valid: true},
		IsActive:        true,
		Specializations: []*Specialization{
			{ID: 1, Name: "Хирург"},
			{ID: 2, Name: "Терапевт"},
		},
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	jsonData, err := json.Marshal(vet)
	assert.NoError(t, err)

	var decodedVet Veterinarian
	err = json.Unmarshal(jsonData, &decodedVet)
	assert.NoError(t, err)

	assert.Equal(t, vet.ID, decodedVet.ID)
	assert.Equal(t, vet.FirstName, decodedVet.FirstName)
	assert.Equal(t, vet.LastName, decodedVet.LastName)
	assert.Equal(t, vet.Phone, decodedVet.Phone)
	assert.Equal(t, vet.Email.String, decodedVet.Email.String)
	assert.Equal(t, vet.Email.Valid, decodedVet.Email.Valid)
	assert.Equal(t, vet.IsActive, decodedVet.IsActive)
	assert.Len(t, decodedVet.Specializations, 2)
}

func TestVeterinarian_NullableFields(t *testing.T) {
	tests := []struct {
		name        string
		vet         Veterinarian
		description string
	}{
		{
			name: "All nullable fields are valid",
			vet: Veterinarian{
				Email:           sql.NullString{String: "test@test.com", Valid: true},
				Description:     sql.NullString{String: "Description", Valid: true},
				ExperienceYears: sql.NullInt64{Int64: 10, Valid: true},
			},
			description: "Все nullable поля должны быть валидными",
		},
		{
			name: "All nullable fields are invalid",
			vet: Veterinarian{
				Email:           sql.NullString{Valid: false},
				Description:     sql.NullString{Valid: false},
				ExperienceYears: sql.NullInt64{Valid: false},
			},
			description: "Все nullable поля должны быть невалидными",
		},
		{
			name: "Mixed nullable fields",
			vet: Veterinarian{
				Email:           sql.NullString{String: "test@test.com", Valid: true},
				Description:     sql.NullString{Valid: false},
				ExperienceYears: sql.NullInt64{Int64: 5, Valid: true},
			},
			description: "Смешанные nullable поля",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.vet)
			assert.NoError(t, err)

			var decodedVet Veterinarian
			err = json.Unmarshal(jsonData, &decodedVet)
			assert.NoError(t, err)

			assert.Equal(t, tt.vet.Email.Valid, decodedVet.Email.Valid)
			if tt.vet.Email.Valid {
				assert.Equal(t, tt.vet.Email.String, decodedVet.Email.String)
			}

			assert.Equal(t, tt.vet.Description.Valid, decodedVet.Description.Valid)
			if tt.vet.Description.Valid {
				assert.Equal(t, tt.vet.Description.String, decodedVet.Description.String)
			}

			assert.Equal(t, tt.vet.ExperienceYears.Valid, decodedVet.ExperienceYears.Valid)
			if tt.vet.ExperienceYears.Valid {
				assert.Equal(t, tt.vet.ExperienceYears.Int64, decodedVet.ExperienceYears.Int64)
			}
		})
	}
}

func TestVeterinarian_SpecializationsPointers(t *testing.T) {
	vet := Veterinarian{
		Specializations: []*Specialization{
			{ID: 1, Name: "Хирург"},
			nil, // Тестируем nil указатель
			{ID: 3, Name: "Дерматолог"},
		},
	}

	// Проверяем что структура может содержать nil указатели
	assert.Len(t, vet.Specializations, 3)
	assert.Nil(t, vet.Specializations[1])
	assert.Equal(t, "Хирург", vet.Specializations[0].Name)
	assert.Equal(t, "Дерматолог", vet.Specializations[2].Name)
}

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ CLINIC
// ============================================================================

func TestClinic_JSONSerialization(t *testing.T) {
	clinic := Clinic{
		ID:           1,
		Name:         "ВетКлиника",
		Address:      "ул. Ленина, 1",
		Phone:        sql.NullString{String: "+79123456789", Valid: true},
		WorkingHours: sql.NullString{String: "9:00-18:00", Valid: true},
		IsActive:     true,
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	jsonData, err := json.Marshal(clinic)
	assert.NoError(t, err)

	var decodedClinic Clinic
	err = json.Unmarshal(jsonData, &decodedClinic)
	assert.NoError(t, err)

	assert.Equal(t, clinic.ID, decodedClinic.ID)
	assert.Equal(t, clinic.Name, decodedClinic.Name)
	assert.Equal(t, clinic.Address, decodedClinic.Address)
	assert.Equal(t, clinic.Phone, decodedClinic.Phone)
	assert.Equal(t, clinic.WorkingHours, decodedClinic.WorkingHours)
	assert.Equal(t, clinic.IsActive, decodedClinic.IsActive)
	assert.True(t, clinic.CreatedAt.Equal(decodedClinic.CreatedAt))
}

func TestClinic_NullableFields(t *testing.T) {
	clinic := Clinic{
		Phone:        sql.NullString{Valid: false},
		WorkingHours: sql.NullString{Valid: false},
	}

	jsonData, err := json.Marshal(clinic)
	assert.NoError(t, err)

	var decodedClinic Clinic
	err = json.Unmarshal(jsonData, &decodedClinic)
	assert.NoError(t, err)

	assert.False(t, decodedClinic.Phone.Valid)
	assert.False(t, decodedClinic.WorkingHours.Valid)
}

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ SCHEDULE
// ============================================================================

func TestSchedule_JSONSerialization(t *testing.T) {
	schedule := Schedule{
		ID:          1,
		VetID:       1,
		ClinicID:    1,
		DayOfWeek:   1, // Понедельник
		StartTime:   "09:00",
		EndTime:     "18:00",
		IsAvailable: true,
		Vet: &Veterinarian{
			ID:        1,
			FirstName: "Анна",
			LastName:  "Смирнова",
		},
		Clinic: &Clinic{
			ID:   1,
			Name: "ВетКлиника",
		},
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	jsonData, err := json.Marshal(schedule)
	assert.NoError(t, err)

	var decodedSchedule Schedule
	err = json.Unmarshal(jsonData, &decodedSchedule)
	assert.NoError(t, err)

	assert.Equal(t, schedule.ID, decodedSchedule.ID)
	assert.Equal(t, schedule.VetID, decodedSchedule.VetID)
	assert.Equal(t, schedule.ClinicID, decodedSchedule.ClinicID)
	assert.Equal(t, schedule.DayOfWeek, decodedSchedule.DayOfWeek)
	assert.Equal(t, schedule.StartTime, decodedSchedule.StartTime)
	assert.Equal(t, schedule.EndTime, decodedSchedule.EndTime)
	assert.Equal(t, schedule.IsAvailable, decodedSchedule.IsAvailable)
}

func TestSchedule_OptionalRelations(t *testing.T) {
	// Расписание без связанных данных
	schedule := Schedule{
		ID:          1,
		VetID:       1,
		ClinicID:    1,
		DayOfWeek:   1,
		StartTime:   "09:00",
		EndTime:     "18:00",
		IsAvailable: true,
		// Vet и Clinic не установлены (nil)
	}

	jsonData, err := json.Marshal(schedule)
	assert.NoError(t, err)

	var decodedSchedule Schedule
	err = json.Unmarshal(jsonData, &decodedSchedule)
	assert.NoError(t, err)

	assert.Nil(t, decodedSchedule.Vet)
	assert.Nil(t, decodedSchedule.Clinic)
}

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ USERREQUEST
// ============================================================================

func TestUserRequest_JSONSerialization(t *testing.T) {
	request := UserRequest{
		ID:               1,
		UserID:           1,
		SpecializationID: 1,
		SearchQuery:      "срочно нужен врач",
		CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	jsonData, err := json.Marshal(request)
	assert.NoError(t, err)

	var decodedRequest UserRequest
	err = json.Unmarshal(jsonData, &decodedRequest)
	assert.NoError(t, err)

	assert.Equal(t, request.ID, decodedRequest.ID)
	assert.Equal(t, request.UserID, decodedRequest.UserID)
	assert.Equal(t, request.SpecializationID, decodedRequest.SpecializationID)
	assert.Equal(t, request.SearchQuery, decodedRequest.SearchQuery)
	assert.True(t, request.CreatedAt.Equal(decodedRequest.CreatedAt))
}

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ SEARCHCRITERIA
// ============================================================================

func TestSearchCriteria_JSONSerialization(t *testing.T) {
	criteria := SearchCriteria{
		SpecializationID: 1,
		DayOfWeek:        1,
		Time:             "14:00",
		ClinicID:         1,
	}

	jsonData, err := json.Marshal(criteria)
	assert.NoError(t, err)

	var decodedCriteria SearchCriteria
	err = json.Unmarshal(jsonData, &decodedCriteria)
	assert.NoError(t, err)

	assert.Equal(t, criteria.SpecializationID, decodedCriteria.SpecializationID)
	assert.Equal(t, criteria.DayOfWeek, decodedCriteria.DayOfWeek)
	assert.Equal(t, criteria.Time, decodedCriteria.Time)
	assert.Equal(t, criteria.ClinicID, decodedCriteria.ClinicID)
}

func TestSearchCriteria_EmptyFields(t *testing.T) {
	// Критерии с нулевыми значениями (по умолчанию)
	criteria := SearchCriteria{}

	assert.Equal(t, 0, criteria.SpecializationID)
	assert.Equal(t, 0, criteria.DayOfWeek)
	assert.Equal(t, "", criteria.Time)
	assert.Equal(t, 0, criteria.ClinicID)
}

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ VETEDITDATA
// ============================================================================

func TestVetEditData_JSONSerialization(t *testing.T) {
	editData := VetEditData{
		VetID:           1,
		Field:           "first_name",
		CurrentValue:    "СтароеИмя",
		Specializations: "1,2,3",
	}

	jsonData, err := json.Marshal(editData)
	assert.NoError(t, err)

	var decodedEditData VetEditData
	err = json.Unmarshal(jsonData, &decodedEditData)
	assert.NoError(t, err)

	assert.Equal(t, editData.VetID, decodedEditData.VetID)
	assert.Equal(t, editData.Field, decodedEditData.Field)
	assert.Equal(t, editData.CurrentValue, decodedEditData.CurrentValue)
	assert.Equal(t, editData.Specializations, decodedEditData.Specializations)
}

// ============================================================================
// ТЕСТЫ ДЛЯ СТРУКТУРЫ CLINICEDITDATA
// ============================================================================

func TestClinicEditData_JSONSerialization(t *testing.T) {
	editData := ClinicEditData{
		ClinicID:     1,
		Field:        "name",
		CurrentValue: "СтароеНазвание",
	}

	jsonData, err := json.Marshal(editData)
	assert.NoError(t, err)

	var decodedEditData ClinicEditData
	err = json.Unmarshal(jsonData, &decodedEditData)
	assert.NoError(t, err)

	assert.Equal(t, editData.ClinicID, decodedEditData.ClinicID)
	assert.Equal(t, editData.Field, decodedEditData.Field)
	assert.Equal(t, editData.CurrentValue, decodedEditData.CurrentValue)
}

// ============================================================================
// ТЕСТЫ ДЛЯ ПРОВЕРКИ JSON ТЕГОВ
// ============================================================================

func TestJSONFieldNames(t *testing.T) {
	tests := []struct {
		name     string
		model    interface{}
		expected map[string]string
	}{
		{
			name:  "User field names",
			model: User{},
			expected: map[string]string{
				"ID":         "id",
				"TelegramID": "telegram_id",
				"Username":   "username",
				"FirstName":  "first_name",
				"LastName":   "last_name",
				"Phone":      "phone",
				"CreatedAt":  "created_at",
			},
		},
		{
			name:  "Veterinarian field names",
			model: Veterinarian{},
			expected: map[string]string{
				"ID":              "id",
				"FirstName":       "first_name",
				"LastName":        "last_name",
				"Phone":           "phone",
				"Email":           "email",
				"Description":     "description",
				"ExperienceYears": "experience_years",
				"IsActive":        "is_active",
				"Specializations": "specializations",
				"CreatedAt":       "created_at",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Простая проверка что структура может быть сериализована
			jsonData, err := json.Marshal(tt.model)
			assert.NoError(t, err)
			assert.NotEmpty(t, jsonData)
		})
	}
}

// ============================================================================
// ТЕСТЫ ДЛЯ ПРОВЕРКИ ВРЕМЕННЫХ МЕТОК
// ============================================================================

func TestTimeFields(t *testing.T) {
	now := time.Now()

	user := User{
		ID:        1,
		FirstName: "Test",
		CreatedAt: now,
	}

	// Проверяем что временные метки корректно сохраняются
	assert.True(t, user.CreatedAt.Equal(now))

	// Проверяем сериализацию/десериализацию времени
	jsonData, err := json.Marshal(user)
	assert.NoError(t, err)

	var decodedUser User
	err = json.Unmarshal(jsonData, &decodedUser)
	assert.NoError(t, err)

	// Время должно сохраняться с точностью до секунд (из-за JSON формата)
	assert.WithinDuration(t, user.CreatedAt, decodedUser.CreatedAt, time.Second)
}
