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
	ID              sql.NullInt64     `json:"id"`
	FirstName       string            `json:"first_name"`
	LastName        string            `json:"last_name"`
	Patronymic      sql.NullString    `json:"patronymic"`
	Phone           string            `json:"phone"`
	Email           sql.NullString    `json:"email"`
	Description     sql.NullString    `json:"description"`
	ExperienceYears sql.NullInt64     `json:"experience_years"`
	IsActive        bool              `json:"is_active"`
	CityID          sql.NullInt64     `json:"city_id"`
	CreatedAt       time.Time         `json:"created_at"`
	Specializations []*Specialization `json:"specializations,omitempty"`

	// Для удобства - связанные данные
	City      *City       `json:"city,omitempty"`
	Schedules []*Schedule `json:"schedules,omitempty"`
}

// City представляет населенный пункт
type City struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
}

// Clinic представляет клинику/место приема
type Clinic struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Address      string         `json:"address"`
	Phone        sql.NullString `json:"phone"`         // Может быть NULL
	WorkingHours sql.NullString `json:"working_hours"` // Может быть NULL
	IsActive     bool           `json:"is_active"`     // Добавлено поле для деактивации
	CityID       sql.NullInt64  `json:"city_id"`       // Ссылка на город
	District     sql.NullString `json:"district"`      // Район города
	MetroStation sql.NullString `json:"metro_station"` // Станция метро
	CreatedAt    time.Time      `json:"created_at"`

	// Для удобства - связанные данные
	City *City `json:"city,omitempty"`
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

// VetEditData временные данные для редактирования врача
type VetEditData struct {
	VetID           int    `json:"vet_id"`
	Field           string `json:"field"`
	CurrentValue    string `json:"current_value"`
	Specializations string `json:"specializations"`
}

// ClinicEditData временные данные для редактирования клиники
type ClinicEditData struct {
	ClinicID     int    `json:"clinic_id"`
	Field        string `json:"field"`
	CurrentValue string `json:"current_value"`
}

// ImportResult представляет результат импорта
type ImportResult struct {
	TotalRows    int           `json:"total_rows"`
	SuccessCount int           `json:"success_count"`
	ErrorCount   int           `json:"error_count"`
	Errors       []ImportError `json:"errors"`
}

type ImportError struct {
	RowNumber int    `json:"row_number"`
	Field     string `json:"field"`
	Message   string `json:"message"`
}

// Добавляем новые структуры для импорта
type ImportRequest struct {
	Type      string        `json:"type"` // "cities", "clinics", "veterinarians"
	FilePath  string        `json:"file_path"`
	UserID    int64         `json:"user_id"`
	Status    string        `json:"status"` // "pending", "processing", "completed", "failed"
	Result    *ImportResult `json:"result,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// Расширяем SearchCriteria для поиска по городам
type SearchCriteria struct {
	SpecializationID int    `json:"specialization_id"`
	DayOfWeek        int    `json:"day_of_week"`
	Time             string `json:"time"`
	ClinicID         int    `json:"clinic_id"`
	CityID           int    `json:"city_id"`       // Добавляем поиск по городу
	CityName         string `json:"city_name"`     // Поиск по названию города
	District         string `json:"district"`      // Поиск по району
	MetroStation     string `json:"metro_station"` // Поиск по станции метро
}

// CityEditData временные данные для редактирования города
type CityEditData struct {
	CityID       int
	Field        string
	CurrentValue string
}

// GetVetIDAsInt безопасно конвертирует vet.ID в int
func GetVetIDAsInt(vet *Veterinarian) (int, bool) {
	if !vet.ID.Valid {
		return 0, false
	}
	return int(vet.ID.Int64), true
}

// GetVetIDAsIntOrZero возвращает ID как int или 0 если невалиден
func GetVetIDAsIntOrZero(vet *Veterinarian) int {
	if !vet.ID.Valid {
		return 0
	}
	return int(vet.ID.Int64)
}

// Review представляет отзыв о враче
type Review struct {
	ID             int           `json:"id"`
	VeterinarianID int           `json:"veterinarian_id"`
	UserID         int           `json:"user_id"`
	Rating         int           `json:"rating"` // 1-5 звезд
	Comment        string        `json:"comment"`
	Status         string        `json:"status"` // pending/approved/rejected
	CreatedAt      time.Time     `json:"created_at"`
	ModeratedAt    sql.NullTime  `json:"moderated_at"`
	ModeratedBy    sql.NullInt64 `json:"moderated_by"`
	ModeratorID    int           `json:"moderator_id"`

	// Для удобства - связанные данные
	Veterinarian *Veterinarian `json:"veterinarian,omitempty"`
	User         *User         `json:"user,omitempty"`
	Moderator    *User         `json:"moderator,omitempty"`
}

// ReviewStats представляет статистику отзывов по врачу
type ReviewStats struct {
	VeterinarianID  int     `json:"veterinarian_id"`
	AverageRating   float64 `json:"average_rating"`
	TotalReviews    int     `json:"total_reviews"`
	ApprovedReviews int     `json:"approved_reviews"`
}

// ReviewEditData временные данные для модерации отзывов
type ReviewEditData struct {
	ReviewID int    `json:"review_id"`
	Action   string `json:"action"` // approve/reject
}
