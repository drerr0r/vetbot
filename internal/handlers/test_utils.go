package handlers

import (
	"database/sql"
	"fmt"

	"github.com/drerr0r/vetbot/internal/models"
	"github.com/drerr0r/vetbot/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ============================================================================
// MOCK DATABASE
// ============================================================================

// MockDatabase представляет мок для базы данных
type MockDatabase struct {
	Users                map[int64]*models.User
	Specializations      map[int]*models.Specialization
	Veterinarians        map[int]*models.Veterinarian
	Clinics              map[int]*models.Clinic
	Schedules            map[int]*models.Schedule
	UserError            error
	SpecializationsError error
	VeterinariansError   error
	ClinicsError         error
	SchedulesError       error
}

// NewMockDatabase создает новый мок базы данных
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		Users:           make(map[int64]*models.User),
		Specializations: make(map[int]*models.Specialization),
		Veterinarians:   make(map[int]*models.Veterinarian),
		Clinics:         make(map[int]*models.Clinic),
		Schedules:       make(map[int]*models.Schedule),
	}
}

// GetDB возвращает объект базы данных (для совместимости)
func (m *MockDatabase) GetDB() *sql.DB {
	return nil // Возвращаем nil для тестов
}

// Close закрывает подключение (для совместимости)
func (m *MockDatabase) Close() error {
	return nil
}

// CreateUser создает пользователя
func (m *MockDatabase) CreateUser(user *models.User) error {
	if m.UserError != nil {
		return m.UserError
	}
	m.Users[user.TelegramID] = user
	return nil
}

// GetAllSpecializations возвращает все специализации
func (m *MockDatabase) GetAllSpecializations() ([]*models.Specialization, error) {
	if m.SpecializationsError != nil {
		return nil, m.SpecializationsError
	}

	result := make([]*models.Specialization, 0, len(m.Specializations))
	for _, spec := range m.Specializations {
		result = append(result, spec)
	}
	return result, nil
}

// GetSpecializationByID возвращает специализацию по ID
func (m *MockDatabase) GetSpecializationByID(id int) (*models.Specialization, error) {
	if m.SpecializationsError != nil {
		return nil, m.SpecializationsError
	}

	spec, exists := m.Specializations[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return spec, nil
}

// SpecializationExists проверяет существование специализации
func (m *MockDatabase) SpecializationExists(id int) (bool, error) {
	_, exists := m.Specializations[id]
	return exists, nil
}

// GetVeterinariansBySpecialization возвращает врачей по специализации
func (m *MockDatabase) GetVeterinariansBySpecialization(specializationID int) ([]*models.Veterinarian, error) {
	if m.VeterinariansError != nil {
		return nil, m.VeterinariansError
	}

	result := make([]*models.Veterinarian, 0)
	for _, vet := range m.Veterinarians {
		for _, spec := range vet.Specializations {
			if spec.ID == specializationID {
				result = append(result, vet)
				break
			}
		}
	}
	return result, nil
}

// GetAllClinics возвращает все клиники
func (m *MockDatabase) GetAllClinics() ([]*models.Clinic, error) {
	if m.ClinicsError != nil {
		return nil, m.ClinicsError
	}

	result := make([]*models.Clinic, 0, len(m.Clinics))
	for _, clinic := range m.Clinics {
		result = append(result, clinic)
	}
	return result, nil
}

// GetSchedulesByVetID возвращает расписание врача
func (m *MockDatabase) GetSchedulesByVetID(vetID int) ([]*models.Schedule, error) {
	if m.SchedulesError != nil {
		return nil, m.SchedulesError
	}

	result := make([]*models.Schedule, 0)
	for _, schedule := range m.Schedules {
		if schedule.VetID == vetID {
			result = append(result, schedule)
		}
	}
	return result, nil
}

// GetSpecializationsByVetID возвращает специализации врача
func (m *MockDatabase) GetSpecializationsByVetID(vetID int) ([]*models.Specialization, error) {
	vet, exists := m.Veterinarians[vetID]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return vet.Specializations, nil
}

// FindAvailableVets ищет доступных врачей
func (m *MockDatabase) FindAvailableVets(criteria *models.SearchCriteria) ([]*models.Veterinarian, error) {
	if m.VeterinariansError != nil {
		return nil, m.VeterinariansError
	}

	result := make([]*models.Veterinarian, 0)

	for _, vet := range m.Veterinarians {
		// Фильтрация по специализации
		if criteria.SpecializationID > 0 {
			found := false
			for _, spec := range vet.Specializations {
				if spec.ID == criteria.SpecializationID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Фильтрация по дню недели
		if criteria.DayOfWeek > 0 {
			hasSchedule := false
			for _, schedule := range m.Schedules {
				if schedule.VetID == vet.ID && schedule.DayOfWeek == criteria.DayOfWeek {
					hasSchedule = true
					break
				}
			}
			if !hasSchedule {
				continue
			}
		}

		// Фильтрация по клинике
		if criteria.ClinicID > 0 {
			hasClinic := false
			for _, schedule := range m.Schedules {
				if schedule.VetID == vet.ID && schedule.ClinicID == criteria.ClinicID {
					hasClinic = true
					break
				}
			}
			if !hasClinic {
				continue
			}
		}

		result = append(result, vet)
	}

	return result, nil
}

// GetAllVeterinarians возвращает всех врачей
func (m *MockDatabase) GetAllVeterinarians() ([]*models.Veterinarian, error) {
	result := make([]*models.Veterinarian, 0, len(m.Veterinarians))
	for _, vet := range m.Veterinarians {
		result = append(result, vet)
	}
	return result, nil
}

// GetVeterinarianByID возвращает врача по ID
func (m *MockDatabase) GetVeterinarianByID(id int) (*models.Veterinarian, error) {
	vet, exists := m.Veterinarians[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return vet, nil
}

// GetClinicByID возвращает клинику по ID
func (m *MockDatabase) GetClinicByID(id int) (*models.Clinic, error) {
	clinic, exists := m.Clinics[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return clinic, nil
}

// AddMissingColumns добавляет отсутствующие колонки
func (m *MockDatabase) AddMissingColumns() error {
	return nil
}

// AddTestVeterinarian добавляет тестового ветеринара
func (m *MockDatabase) AddTestVeterinarian(id int, firstName, lastName, phone string) {
	m.Veterinarians[id] = &models.Veterinarian{
		ID:              id,
		FirstName:       firstName,
		LastName:        lastName,
		Phone:           phone,
		Specializations: []*models.Specialization{},
	}
}

// AddTestSpecialization добавляет тестовую специализацию
func (m *MockDatabase) AddTestSpecialization(id int, name string) {
	m.Specializations[id] = &models.Specialization{
		ID:   id,
		Name: name,
	}
}

// ============================================================================
// MOCK BOT
// ============================================================================

// MockBot представляет мок для Telegram бота
type MockBot struct {
	SentMessages   []tgbotapi.MessageConfig
	Callbacks      []tgbotapi.CallbackConfig
	EditedMessages []tgbotapi.EditMessageTextConfig
}

// NewMockBot создает новый мок бота
func NewMockBot() *MockBot {
	return &MockBot{
		SentMessages:   make([]tgbotapi.MessageConfig, 0),
		Callbacks:      make([]tgbotapi.CallbackConfig, 0),
		EditedMessages: make([]tgbotapi.EditMessageTextConfig, 0),
	}
}

// Send имитирует отправку сообщения
func (m *MockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	switch msg := c.(type) {
	case tgbotapi.MessageConfig:
		m.SentMessages = append(m.SentMessages, msg)
		return tgbotapi.Message{MessageID: len(m.SentMessages)}, nil
	case tgbotapi.CallbackConfig:
		m.Callbacks = append(m.Callbacks, msg)
		return tgbotapi.Message{}, nil
	case tgbotapi.EditMessageTextConfig:
		m.EditedMessages = append(m.EditedMessages, msg)
		return tgbotapi.Message{MessageID: len(m.EditedMessages)}, nil
	default:
		return tgbotapi.Message{}, fmt.Errorf("unsupported message type: %T", c)
	}
}

// Request имитирует запрос к API
func (m *MockBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{Ok: true}, nil
}

// GetSentMessages возвращает отправленные сообщения
func (m *MockBot) GetSentMessages() []tgbotapi.MessageConfig {
	return m.SentMessages
}

// GetLastMessage возвращает последнее отправленное сообщение
func (m *MockBot) GetLastMessage() *tgbotapi.MessageConfig {
	if len(m.SentMessages) == 0 {
		return nil
	}
	return &m.SentMessages[len(m.SentMessages)-1]
}

// GetLastEditedMessage возвращает последнее отредактированное сообщение
func (m *MockBot) GetLastEditedMessage() *tgbotapi.EditMessageTextConfig {
	if len(m.EditedMessages) == 0 {
		return nil
	}
	return &m.EditedMessages[len(m.EditedMessages)-1]
}

// Clear очищает историю сообщений
func (m *MockBot) Clear() {
	m.SentMessages = make([]tgbotapi.MessageConfig, 0)
	m.Callbacks = make([]tgbotapi.CallbackConfig, 0)
	m.EditedMessages = make([]tgbotapi.EditMessageTextConfig, 0)
}

// ============================================================================
// TEST UTILITIES
// ============================================================================

// TestUpdateBuilder помогает создавать тестовые обновления
type TestUpdateBuilder struct {
	update tgbotapi.Update
}

// NewTestUpdate создает новый билдер обновлений
func NewTestUpdate() *TestUpdateBuilder {
	return &TestUpdateBuilder{
		update: tgbotapi.Update{},
	}
}

// WithMessage добавляет сообщение
func (b *TestUpdateBuilder) WithMessage(text string, chatID int64, userID int64) *TestUpdateBuilder {
	b.update.Message = &tgbotapi.Message{
		Text: text,
		Chat: &tgbotapi.Chat{ID: chatID},
		From: &tgbotapi.User{ID: userID},
	}
	return b
}

// WithCallback добавляет callback query
func (b *TestUpdateBuilder) WithCallback(data string, chatID int64, messageID int) *TestUpdateBuilder {
	b.update.CallbackQuery = &tgbotapi.CallbackQuery{
		ID:      "test_callback",
		Data:    data,
		Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, MessageID: messageID},
	}
	return b
}

// Build возвращает собранное обновление
func (b *TestUpdateBuilder) Build() tgbotapi.Update {
	return b.update
}

// CreateTestConfig создает тестовую конфигурацию
func CreateTestConfig() *utils.Config {
	return &utils.Config{
		TelegramToken: "test_token",
		DatabaseURL:   "test_url",
		Debug:         true,
		AdminIDs:      []int64{12345},
	}
}

// CreateTestMainHandlers создает тестовые MainHandler
func CreateTestMainHandlers() (*MainHandler, *MockBot) {
	mockBot := NewMockBot()

	// Создаем nil базу данных для тестов
	var db Database = nil
	config := CreateTestConfig()

	mainHandler := NewMainHandler(mockBot, db, config)
	return mainHandler, mockBot
}

// CreateTestVetHandlers создает тестовые VetHandlers
func CreateTestVetHandlers() (*VetHandlers, *MockBot, *MockDatabase) {
	mockBot := NewMockBot()
	mockDB := NewMockDatabase()
	handlers := NewVetHandlers(mockBot, mockDB)
	return handlers, mockBot, mockDB
}

// CreateTestAdminHandlers создает тестовые AdminHandlers
func CreateTestAdminHandlers() (*AdminHandlers, *MockBot, *MockDatabase) {
	mockBot := NewMockBot()
	mockDB := NewMockDatabase()
	config := CreateTestConfig()
	handlers := NewAdminHandlers(mockBot, mockDB, config)
	return handlers, mockBot, mockDB
}
