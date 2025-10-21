package handlers

import (
	"database/sql"

	"github.com/drerr0r/vetbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotAPI интерфейс для Telegram бота
type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetToken() string
}

// Database интерфейс для работы с базой данных
type Database interface {
	CreateUser(user *models.User) error
	GetAllSpecializations() ([]*models.Specialization, error)
	GetSpecializationByID(id int) (*models.Specialization, error)
	GetVeterinariansBySpecialization(specializationID int) ([]*models.Veterinarian, error)
	GetAllClinics() ([]*models.Clinic, error)
	GetSchedulesByVetID(vetID int) ([]*models.Schedule, error)
	GetSpecializationsByVetID(vetID int) ([]*models.Specialization, error)
	FindAvailableVets(criteria *models.SearchCriteria) ([]*models.Veterinarian, error)
	GetAllVeterinarians() ([]*models.Veterinarian, error)
	GetVeterinarianByID(id int) (*models.Veterinarian, error)
	GetClinicByID(id int) (*models.Clinic, error)
	SpecializationExists(id int) (bool, error)
	AddMissingColumns() error
	CreateReview(review *models.Review) error
	GetReviewByID(reviewID int) (*models.Review, error)
	GetApprovedReviewsByVet(vetID int) ([]*models.Review, error)
	GetPendingReviews() ([]*models.Review, error)
	UpdateReviewStatus(reviewID int, status string, moderatorID int) error
	HasUserReviewForVet(userID int, vetID int) (bool, error)
	GetReviewStats(vetID int) (*models.ReviewStats, error)
	GetUserByTelegramID(telegramID int64) (*models.User, error)
	Close() error
	GetDB() *sql.DB

	// Новые методы для расширенного поиска
	GetClinicsByCity(cityID int) ([]*models.Clinic, error)
	FindVetsByCity(criteria *models.SearchCriteria) ([]*models.Veterinarian, error)
	GetCitiesByRegion(region string) ([]*models.City, error)
	SearchCities(queryStr string) ([]*models.City, error)

	// Новые методы для работы с клиниками
	CreateClinicWithCity(clinic *models.Clinic) error
	GetAllClinicsWithCities() ([]*models.Clinic, error)
	UpdateClinic(clinic *models.Clinic) error

	// Методы для городов
	CreateCity(city *models.City) error
	GetCityByID(id int) (*models.City, error)
	GetCityByName(name string) (*models.City, error)
	GetAllCities() ([]*models.City, error)
	SearchCitiesByRegion(region string) ([]*models.City, error)
	UpdateCity(city *models.City) error
	DeleteCity(id int) error

	// Методы для врачей
	CreateVeterinarian(vet *models.Veterinarian) error
	UpdateVeterinarian(vet *models.Veterinarian) error

	// Новый метод для получения полной информации о враче
	GetVeterinarianWithDetails(id int) (*models.Veterinarian, error)

	GetSpecializationByName(name string) (*models.Specialization, error)
	CreateSpecialization(spec *models.Specialization) error
	AddVeterinarianSpecialization(vetID int, specID int) error

	// Дебаг
	DebugSpecializationVetsCount() (map[int]int, error)

	// НОВЫЕ МЕТОДЫ ДЛЯ АДМИН-ПАНЕЛИ И СТАТИСТИКИ
	GetActiveClinicCount() (int, error)
	GetTotalClinicCount() (int, error)
	GetActiveVetCount() (int, error)
	GetTotalVetCount() (int, error)
	GetUserCount() (int, error)
	GetRequestCount() (int, error)

	// Методы для удаления
	DeleteClinic(clinicID int) error
	DeleteVeterinarian(vetID int) error

	// Методы для статистики городов
	GetCitiesCount() (int, error)
	GetVetsCountByCity(cityID int) (int, error)
	GetClinicsCountByCity(cityID int) (int, error)

	GetUserByID(userID int) (*models.User, error)

	GetAllActiveVeterinarians() ([]*models.Veterinarian, error)
}
