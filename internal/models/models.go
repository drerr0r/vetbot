package models

import (
	"database/sql"
	"time"
)

// User представляет пользователя бота
type User struct {
	ID         int       `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Phone      string    `json:"phone"`
	CreatedAt  time.Time `json:"created_at"`
}

// Specialization представляет специализацию врача
type Specialization struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Veterinarian представляет ветеринарного врача
type Veterinarian struct {
	ID              int               `json:"id"`
	FirstName       string            `json:"first_name"`
	LastName        string            `json:"last_name"`
	Phone           string            `json:"phone"`
	Email           sql.NullString    `json:"email"`            // Может быть NULL
	Description     sql.NullString    `json:"description"`      // Может быть NULL
	ExperienceYears sql.NullInt64     `json:"experience_years"` // Может быть NULL
	IsActive        bool              `json:"is_active"`
	Specializations []*Specialization `json:"specializations"` // Исправлено на указатели
	CreatedAt       time.Time         `json:"created_at"`
}

// Clinic представляет клинику/место приема
type Clinic struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Address      string         `json:"address"`
	Phone        sql.NullString `json:"phone"`         // Может быть NULL
	WorkingHours sql.NullString `json:"working_hours"` // Может быть NULL
	CreatedAt    time.Time      `json:"created_at"`
}

// Schedule представляет расписание врача
type Schedule struct {
	ID          int           `json:"id"`
	VetID       int           `json:"vet_id"`
	ClinicID    int           `json:"clinic_id"`
	DayOfWeek   int           `json:"day_of_week"`
	StartTime   string        `json:"start_time"`
	EndTime     string        `json:"end_time"`
	IsAvailable bool          `json:"is_available"`
	Vet         *Veterinarian `json:"vet,omitempty"`
	Clinic      *Clinic       `json:"clinic,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// UserRequest представляет запрос пользователя
type UserRequest struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	SpecializationID int       `json:"specialization_id"`
	SearchQuery      string    `json:"search_query"`
	CreatedAt        time.Time `json:"created_at"`
}

// SearchCriteria представляет критерии поиска врачей
type SearchCriteria struct {
	SpecializationID int    `json:"specialization_id"`
	DayOfWeek        int    `json:"day_of_week"`
	Time             string `json:"time"`
	ClinicID         int    `json:"clinic_id"`
}
