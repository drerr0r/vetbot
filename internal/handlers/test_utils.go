package handlers

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/drerr0r/vetbot/internal/models"
	"github.com/drerr0r/vetbot/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xuri/excelize/v2"
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
	Cities               map[int]*models.City
	UserError            error
	SpecializationsError error
	VeterinariansError   error
	ClinicsError         error
	SchedulesError       error
	CitiesError          error
}

// NewMockDatabase создает новый мок базы данных
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		Users:           make(map[int64]*models.User),
		Specializations: make(map[int]*models.Specialization),
		Veterinarians:   make(map[int]*models.Veterinarian),
		Clinics:         make(map[int]*models.Clinic),
		Schedules:       make(map[int]*models.Schedule),
		Cities:          make(map[int]*models.City),
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

		// Фильтрация по городу
		if criteria.CityID > 0 {
			hasCity := false
			for _, schedule := range m.Schedules {
				if schedule.VetID == vet.ID {
					// Здесь должна быть логика проверки города через клинику
					// Для упрощения считаем, что все врачи работают в указанном городе
					hasCity = true
					break
				}
			}
			if !hasCity {
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

// ============================================================================
// НОВЫЕ МЕТОДЫ ДЛЯ РАБОТЫ С ГОРОДАМИ
// ============================================================================

// GetAllCities возвращает все города
func (m *MockDatabase) GetAllCities() ([]*models.City, error) {
	if m.CitiesError != nil {
		return nil, m.CitiesError
	}

	result := make([]*models.City, 0, len(m.Cities))
	for _, city := range m.Cities {
		result = append(result, city)
	}
	return result, nil
}

// GetCityByID возвращает город по ID
func (m *MockDatabase) GetCityByID(id int) (*models.City, error) {
	city, exists := m.Cities[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return city, nil
}

// GetCityByName возвращает город по названию
func (m *MockDatabase) GetCityByName(name string) (*models.City, error) {
	for _, city := range m.Cities {
		if city.Name == name {
			return city, nil
		}
	}
	return nil, sql.ErrNoRows
}

// CreateCity создает новый город
func (m *MockDatabase) CreateCity(city *models.City) error {
	if m.CitiesError != nil {
		return m.CitiesError
	}

	// Генерируем ID если не установлен
	if city.ID == 0 {
		city.ID = len(m.Cities) + 1
	}

	m.Cities[city.ID] = city
	return nil
}

// GetClinicsByCity возвращает клиники по городу
func (m *MockDatabase) GetClinicsByCity(cityID int) ([]*models.Clinic, error) {
	result := make([]*models.Clinic, 0)
	for _, clinic := range m.Clinics {
		if clinic.CityID.Valid && int(clinic.CityID.Int64) == cityID {
			result = append(result, clinic)
		}
	}
	return result, nil
}

// FindVetsByCity ищет врачей по городу
func (m *MockDatabase) FindVetsByCity(criteria *models.SearchCriteria) ([]*models.Veterinarian, error) {
	// Используем существующую логику поиска
	return m.FindAvailableVets(criteria)
}

// GetCitiesByRegion возвращает города по региону
func (m *MockDatabase) GetCitiesByRegion(region string) ([]*models.City, error) {
	result := make([]*models.City, 0)
	for _, city := range m.Cities {
		if city.Region == region {
			result = append(result, city)
		}
	}
	return result, nil
}

// SearchCities ищет города по названию
func (m *MockDatabase) SearchCities(queryStr string) ([]*models.City, error) {
	result := make([]*models.City, 0)
	for _, city := range m.Cities {
		if containsIgnoreCase(city.Name, queryStr) {
			result = append(result, city)
		}
	}
	return result, nil
}

// CreateClinicWithCity создает клинику с привязкой к городу
func (m *MockDatabase) CreateClinicWithCity(clinic *models.Clinic) error {
	if m.ClinicsError != nil {
		return m.ClinicsError
	}

	// Генерируем ID если не установлен
	if clinic.ID == 0 {
		clinic.ID = len(m.Clinics) + 1
	}

	m.Clinics[clinic.ID] = clinic
	return nil
}

// GetAllClinicsWithCities возвращает все клиники с информацией о городах
func (m *MockDatabase) GetAllClinicsWithCities() ([]*models.Clinic, error) {
	result := make([]*models.Clinic, 0, len(m.Clinics))
	for _, clinic := range m.Clinics {
		if clinic.CityID.Valid {
			if city, exists := m.Cities[int(clinic.CityID.Int64)]; exists {
				clinic.City = city
			}
		}
		result = append(result, clinic)
	}
	return result, nil
}

// UpdateClinic обновляет данные клиники
func (m *MockDatabase) UpdateClinic(clinic *models.Clinic) error {
	if m.ClinicsError != nil {
		return m.ClinicsError
	}

	if _, exists := m.Clinics[clinic.ID]; !exists {
		return sql.ErrNoRows
	}

	m.Clinics[clinic.ID] = clinic
	return nil
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ============================================================================

func containsIgnoreCase(s, substr string) bool {
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
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

// AddTestCity добавляет тестовый город
func (m *MockDatabase) AddTestCity(id int, name, region string) {
	m.Cities[id] = &models.City{
		ID:     id,
		Name:   name,
		Region: region,
	}
}

// AddTestClinic добавляет тестовую клинику
func (m *MockDatabase) AddTestClinic(id int, name, address string, cityID int) {
	m.Clinics[id] = &models.Clinic{
		ID:       id,
		Name:     name,
		Address:  address,
		CityID:   sql.NullInt64{Int64: int64(cityID), Valid: true},
		IsActive: true,
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
	Files          map[string]tgbotapi.File // Для хранения файлов
}

// NewMockBot создает новый мок бота
func NewMockBot() *MockBot {
	return &MockBot{
		SentMessages:   make([]tgbotapi.MessageConfig, 0),
		Callbacks:      make([]tgbotapi.CallbackConfig, 0),
		EditedMessages: make([]tgbotapi.EditMessageTextConfig, 0),
		Files:          make(map[string]tgbotapi.File),
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

// GetFile имитирует получение файла
func (m *MockBot) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	file, exists := m.Files[config.FileID]
	if !exists {
		return tgbotapi.File{}, fmt.Errorf("file not found: %s", config.FileID)
	}
	return file, nil
}

// Request имитирует запрос к API
func (m *MockBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{Ok: true}, nil
}

// AddTestFile добавляет тестовый файл для мока
func (m *MockBot) AddTestFile(fileID string, filePath string) {
	m.Files[fileID] = tgbotapi.File{
		FileID:   fileID,
		FilePath: filePath,
	}
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
	m.Files = make(map[string]tgbotapi.File)
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
	mockDB := NewMockDatabase()
	config := CreateTestConfig()

	mainHandler := NewMainHandler(mockBot, mockDB, config)
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

// DeleteCity удаляет город
func (m *MockDatabase) DeleteCity(id int) error {
	delete(m.Cities, id)
	return nil
}

// UpdateCity обновляет город
func (m *MockDatabase) UpdateCity(city *models.City) error {
	if _, exists := m.Cities[city.ID]; !exists {
		return fmt.Errorf("city not found")
	}
	m.Cities[city.ID] = city
	return nil
}

// SearchCitiesByRegion ищет города по региону
func (m *MockDatabase) SearchCitiesByRegion(region string) ([]*models.City, error) {
	if m.CitiesError != nil {
		return nil, m.CitiesError
	}

	result := make([]*models.City, 0)
	for _, city := range m.Cities {
		if containsIgnoreCase(city.Region, region) {
			result = append(result, city)
		}
	}
	return result, nil
}

// CreateVeterinarian создает нового ветеринара
func (m *MockDatabase) CreateVeterinarian(vet *models.Veterinarian) error {
	if m.VeterinariansError != nil {
		return m.VeterinariansError
	}

	// Генерируем ID если не установлен
	if vet.ID == 0 {
		vet.ID = len(m.Veterinarians) + 1
	}

	// Устанавливаем время создания если не установлено
	if vet.CreatedAt.IsZero() {
		vet.CreatedAt = time.Now()
	}

	// Создаем копию ветеринара, чтобы избежать изменений исходного объекта
	newVet := *vet
	m.Veterinarians[newVet.ID] = &newVet

	return nil
}

// UpdateVeterinarian обновляет данные ветеринара
func (m *MockDatabase) UpdateVeterinarian(vet *models.Veterinarian) error {
	if m.VeterinariansError != nil {
		return m.VeterinariansError
	}

	if _, exists := m.Veterinarians[vet.ID]; !exists {
		return sql.ErrNoRows
	}

	m.Veterinarians[vet.ID] = vet
	return nil
}

// GetUserByTelegramID возвращает пользователя по Telegram ID
func (m *MockDatabase) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	if m.UserError != nil {
		return nil, m.UserError
	}

	user, exists := m.Users[telegramID]
	if !exists {
		return nil, sql.ErrNoRows
	}

	return user, nil
}

// CreateClinic создает новую клинику (аналог CreateClinicWithCity для совместимости)
func (m *MockDatabase) CreateClinic(clinic *models.Clinic) error {
	return m.CreateClinicWithCity(clinic)
}

// DeleteClinic удаляет клинику
func (m *MockDatabase) DeleteClinic(id int) error {
	if m.ClinicsError != nil {
		return m.ClinicsError
	}

	if _, exists := m.Clinics[id]; !exists {
		return sql.ErrNoRows
	}

	delete(m.Clinics, id)
	return nil
}

// DeleteVeterinarian удаляет ветеринара (может понадобиться в будущем)
func (m *MockDatabase) DeleteVeterinarian(id int) error {
	if m.VeterinariansError != nil {
		return m.VeterinariansError
	}

	if _, exists := m.Veterinarians[id]; !exists {
		return sql.ErrNoRows
	}

	delete(m.Veterinarians, id)
	return nil
}

// GetVeterinarianWithDetails возвращает врача с полной информацией о городе и клиниках
func (m *MockDatabase) GetVeterinarianWithDetails(id int) (*models.Veterinarian, error) {
	if m.VeterinariansError != nil {
		return nil, m.VeterinariansError
	}

	vet, exists := m.Veterinarians[id]
	if !exists {
		return nil, sql.ErrNoRows
	}

	// Создаем копию ветеринара, чтобы избежать изменений исходного объекта
	vetWithDetails := *vet

	// Загружаем специализации
	specs, err := m.GetSpecializationsByVetID(vet.ID)
	if err == nil {
		vetWithDetails.Specializations = specs
	}

	// Загружаем расписание
	schedules, err := m.GetSchedulesByVetID(vet.ID)
	if err == nil {
		vetWithDetails.Schedules = schedules
	}

	// Загружаем информацию о городе если есть
	if vet.CityID.Valid {
		city, err := m.GetCityByID(int(vet.CityID.Int64))
		if err == nil {
			vetWithDetails.City = city
		}
	}

	return &vetWithDetails, nil
}

// ============================================================================
// МЕТОДЫ ДЛЯ ИМПОРТА
// ============================================================================

// GetSpecializationByName возвращает специализацию по имени
func (m *MockDatabase) GetSpecializationByName(name string) (*models.Specialization, error) {
	if m.SpecializationsError != nil {
		return nil, m.SpecializationsError
	}

	for _, spec := range m.Specializations {
		if spec.Name == name {
			return spec, nil
		}
	}
	return nil, sql.ErrNoRows
}

// CreateSpecialization создает новую специализацию
func (m *MockDatabase) CreateSpecialization(spec *models.Specialization) error {
	if m.SpecializationsError != nil {
		return m.SpecializationsError
	}

	// Генерируем ID если не установлен
	if spec.ID == 0 {
		spec.ID = len(m.Specializations) + 1
	}

	// Устанавливаем время создания если не установлено
	if spec.CreatedAt.IsZero() {
		spec.CreatedAt = time.Now()
	}

	m.Specializations[spec.ID] = spec
	return nil
}

// AddVeterinarianSpecialization добавляет специализацию врачу
func (m *MockDatabase) AddVeterinarianSpecialization(vetID int, specID int) error {
	if m.VeterinariansError != nil {
		return m.VeterinariansError
	}

	vet, exists := m.Veterinarians[vetID]
	if !exists {
		return sql.ErrNoRows
	}

	spec, exists := m.Specializations[specID]
	if !exists {
		return sql.ErrNoRows
	}

	// Добавляем специализацию если ее еще нет
	for _, existingSpec := range vet.Specializations {
		if existingSpec.ID == specID {
			return nil // Уже существует
		}
	}

	vet.Specializations = append(vet.Specializations, spec)
	return nil
}

// ============================================================================
// ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ДЛЯ ТЕСТИРОВАНИЯ ИМПОРТА
// ============================================================================

// CreateTestImportData создает тестовые данные для импорта
func CreateTestImportData() (*MockDatabase, *MockBot) {
	mockDB := NewMockDatabase()
	mockBot := NewMockBot()

	// Добавляем тестовые города
	mockDB.AddTestCity(1, "Москва", "Центральный")
	mockDB.AddTestCity(2, "Санкт-Петербург", "Северо-Западный")
	mockDB.AddTestCity(3, "Новосибирск", "Сибирский")

	// Добавляем тестовые специализации
	mockDB.AddTestSpecialization(1, "Хирургия")
	mockDB.AddTestSpecialization(2, "Терапия")
	mockDB.AddTestSpecialization(3, "Стоматология")
	mockDB.AddTestSpecialization(4, "Дерматология")

	return mockDB, mockBot
}

// CreateTestCSVFile создает тестовый CSV файл для импорта
func CreateTestCSVFile() string {
	content := `Имя	Фамилия	Телефон	Email	Опыт работы	Описание	Специализации	Город	Регион
Иван	Иванов	+79991234567	ivan@vet.ru	5 лет	Опытный врач	Хирургия, терапия	Москва	Центральный
Петр	Петров	+79997654321	petr@vet.ru	3 года	Молодой специалист	Стоматология	Санкт-Петербург	Северо-Западный
Анна	Сидорова	+79995554433	anna@vet.ru	7 лет	Ветеринарный врач	Дерматология	Новосибирск	Сибирский`

	// Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_import_*.csv")
	if err != nil {
		panic(err)
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	if err != nil {
		panic(err)
	}

	return tmpFile.Name()
}

// CreateTestXLSXFile создает тестовый XLSX файл для импорта
func CreateTestXLSXFile() string {
	f := excelize.NewFile()

	// Создаем заголовки
	headers := []string{"Имя", "Фамилия", "Телефон", "Email", "Опыт работы", "Описание", "Специализации", "Город", "Регион"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cell, header)
	}

	// Добавляем тестовые данные
	data := [][]interface{}{
		{"Иван", "Иванов", "+79991234567", "ivan@vet.ru", "5 лет", "Опытный врач", "Хирургия, терапия", "Москва", "Центральный"},
		{"Петр", "Петров", "+79997654321", "petr@vet.ru", "3 года", "Молодой специалист", "Стоматология", "Санкт-Петербург", "Северо-Западный"},
		{"Анна", "Сидорова", "+79995554433", "anna@vet.ru", "7 лет", "Ветеринарный врач", "Дерматология", "Новосибирск", "Сибирский"},
	}

	for row, rowData := range data {
		for col, value := range rowData {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			f.SetCellValue("Sheet1", cell, value)
		}
	}

	// Сохраняем во временный файл
	tmpFile, err := os.CreateTemp("", "test_import_*.xlsx")
	if err != nil {
		panic(err)
	}
	defer tmpFile.Close()

	if err := f.SaveAs(tmpFile.Name()); err != nil {
		panic(err)
	}

	return tmpFile.Name()
}

// CleanupTestFiles удаляет временные тестовые файлы
func CleanupTestFiles(filePaths ...string) {
	for _, filePath := range filePaths {
		os.Remove(filePath)
	}
}

// VerifyVeterinarianImported проверяет, что ветеринар был корректно импортирован
func VerifyVeterinarianImported(db *MockDatabase, firstName, lastName, phone string) bool {
	for _, vet := range db.Veterinarians {
		if vet.FirstName == firstName && vet.LastName == lastName && vet.Phone == phone {
			return true
		}
	}
	return false
}

// VerifySpecializationAdded проверяет, что специализация была добавлена врачу
func VerifySpecializationAdded(db *MockDatabase, vetFirstName, vetLastName, specName string) bool {
	for _, vet := range db.Veterinarians {
		if vet.FirstName == vetFirstName && vet.LastName == vetLastName {
			for _, spec := range vet.Specializations {
				if spec.Name == specName {
					return true
				}
			}
		}
	}
	return false
}

// VerifyCityCreated проверяет, что город был создан
func VerifyCityCreated(db *MockDatabase, cityName string) bool {
	for _, city := range db.Cities {
		if city.Name == cityName {
			return true
		}
	}
	return false
}

// MockBotAPI реализует интерфейс BotAPI для тестов
type MockBotAPI struct {
	Token string
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return tgbotapi.Message{}, nil
}

func (m *MockBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{}, nil
}

func (m *MockBotAPI) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	return tgbotapi.File{}, nil
}

// CreateRealTestMainHandlers создает MainHandler с реальными зависимостями для тестов
func CreateRealTestMainHandlers() (*MainHandler, *MockBot) {
	mockBot := NewMockBot()
	mockDB := NewMockDatabase()
	config := CreateTestConfig()

	// Создаем реальный MainHandler
	mainHandler := NewMainHandler(mockBot, mockDB, config)
	return mainHandler, mockBot
}

// GetToken возвращает тестовый токен для MockBot
func (m *MockBot) GetToken() string {
	return "test_bot_token"
}
