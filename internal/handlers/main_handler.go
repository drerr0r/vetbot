package handlers

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/drerr0r/vetbot/internal/models"
	"github.com/drerr0r/vetbot/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xuri/excelize/v2"
)

// MainHandler обрабатывает все входящие обновления
type MainHandler struct {
	bot           BotAPI   // Используем интерфейс
	db            Database // Используем интерфейс
	config        *utils.Config
	vetHandlers   *VetHandlers
	adminHandlers *AdminHandlers
}

// NewMainHandler создает новый экземпляр MainHandler
func NewMainHandler(bot BotAPI, db Database, config *utils.Config) *MainHandler {
	return &MainHandler{
		bot:           bot,
		db:            db,
		config:        config,
		vetHandlers:   NewVetHandlers(bot, db),
		adminHandlers: NewAdminHandlers(bot, db, config),
	}
}

// HandleUpdate обрабатывает входящее обновление от Telegram
func (h *MainHandler) HandleUpdate(update tgbotapi.Update) {
	InfoLog.Printf("Received update")

	// Обрабатываем callback queries (нажатия на inline кнопки)
	if update.CallbackQuery != nil {
		InfoLog.Printf("Callback query: %s", update.CallbackQuery.Data)
		h.vetHandlers.HandleCallback(update)
		return
	}

	// Обрабатываем документы (файлы для импорта)
	if update.Message != nil && update.Message.Document != nil {
		InfoLog.Printf("Document received: %s", update.Message.Document.FileName)
		h.handleDocument(update)
		return
	}

	// Игнорируем любые не-text сообщения
	if update.Message == nil {
		InfoLog.Printf("Message is nil")
		return
	}

	if update.Message.Text == "" {
		InfoLog.Printf("Text is empty")
		return
	}

	InfoLog.Printf("Processing message: %s", update.Message.Text)

	// Проверяем, является ли пользователь администратором
	isAdmin := h.isAdmin(update.Message.From.ID)
	InfoLog.Printf("User %d is admin: %t", update.Message.From.ID, isAdmin)

	// Если пользователь администратор и находится в админском режиме, передаем админским хендлерам
	if isAdmin && h.isInAdminMode(update.Message.From.ID) {
		InfoLog.Printf("Redirecting to admin handlers")
		h.adminHandlers.HandleAdminMessage(update)
		return
	}

	// Сначала проверяем команды поиска (/search_1, /search_2 и т.д.)
	if strings.HasPrefix(update.Message.Text, "/search_") {
		InfoLog.Printf("Is search command: %s", update.Message.Text)
		h.handleSearchCommand(update)
		return
	}

	// Затем проверяем обычные команды
	if update.Message.IsCommand() {
		InfoLog.Printf("Is command: %s", update.Message.Command())
		h.handleCommand(update, isAdmin)
		return
	}

	// Обычные текстовые сообщения
	InfoLog.Printf("Is text message: %s", update.Message.Text)
	h.handleTextMessage(update)
}

// handleCommand обрабатывает текстовые команды
func (h *MainHandler) handleCommand(update tgbotapi.Update, isAdmin bool) {
	command := update.Message.Command()
	InfoLog.Printf("Handling command: %s", command)

	switch command {
	case "start":
		InfoLog.Printf("Executing /start")
		h.vetHandlers.HandleStart(update)
	case "specializations":
		InfoLog.Printf("Executing /specializations")
		h.vetHandlers.HandleSpecializations(update)
	case "search":
		InfoLog.Printf("Executing /search")
		h.vetHandlers.HandleSearch(update)
	case "clinics":
		InfoLog.Printf("Executing /clinics")
		h.vetHandlers.HandleClinics(update)
	case "cities":
		InfoLog.Printf("Executing /cities")
		h.vetHandlers.HandleSearchByCity(update)
	case "help":
		InfoLog.Printf("Executing /help")
		h.vetHandlers.HandleHelp(update)
	case "test":
		InfoLog.Printf("Executing /test")
		h.vetHandlers.HandleTest(update)
	case "admin":
		if isAdmin {
			InfoLog.Printf("Executing /admin")
			h.adminHandlers.HandleAdmin(update)
		} else {
			InfoLog.Printf("Admin access denied for user %d", update.Message.From.ID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав администратора")
			h.bot.Send(msg)
		}
	case "stats":
		if isAdmin {
			InfoLog.Printf("Executing /stats")
			h.adminHandlers.HandleStats(update)
		}
	default:
		InfoLog.Printf("Unknown command: %s", command)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Неизвестная команда. Используйте /help для списка команд")
		h.bot.Send(msg)
	}
}

// handleSearchCommand обрабатывает команды поиска по специализации (/search_1, /search_2 и т.д.)
func (h *MainHandler) handleSearchCommand(update tgbotapi.Update) {
	text := update.Message.Text
	InfoLog.Printf("Handling search command: %s", text)

	if strings.HasPrefix(text, "/search_") {
		specIDStr := strings.TrimPrefix(text, "/search_")
		specID, err := strconv.Atoi(specIDStr)
		if err != nil {
			ErrorLog.Printf("Error parsing specialization ID: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный формат команды поиска")
			h.bot.Send(msg)
			return
		}
		InfoLog.Printf("Searching for specialization ID: %d", specID)
		h.vetHandlers.HandleSearchBySpecialization(update, specID)
	}
}

// handleTextMessage обрабатывает обычные текстовые сообщения
func (h *MainHandler) handleTextMessage(update tgbotapi.Update) {
	// Для обычных пользователей показываем справку
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"Я понимаю только команды. Используйте /help для списка доступных команд.")
	h.bot.Send(msg)
}

// handleDocument обрабатывает загружаемые документы (CSV/Excel для импорта)
func (h *MainHandler) handleDocument(update tgbotapi.Update) {
	fileName := update.Message.Document.FileName
	fileID := update.Message.Document.FileID

	InfoLog.Printf("Received document: %s", fileName)

	// Проверяем расширение файла
	if !strings.HasSuffix(strings.ToLower(fileName), ".csv") &&
		!strings.HasSuffix(strings.ToLower(fileName), ".xlsx") {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Поддерживаются только CSV и Excel файлы (.csv, .xlsx)")
		h.bot.Send(msg)
		return
	}

	// Проверяем, является ли пользователь администратором
	if !h.isAdmin(update.Message.From.ID) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Импорт данных доступен только администраторам")
		h.bot.Send(msg)
		return
	}

	// Определяем тип импорта по имени файла
	var importType string
	if strings.Contains(strings.ToLower(fileName), "город") {
		importType = "cities"
	} else if strings.Contains(strings.ToLower(fileName), "врач") {
		importType = "veterinarians"
	} else if strings.Contains(strings.ToLower(fileName), "клиник") {
		importType = "clinics"
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"📥 Файл получен. Укажите тип импорта:\n\n"+
				"• Для городов: файл должен содержать 'город' в названии\n"+
				"• Для врачей: файл должен содержать 'врач' в названии\n"+
				"• Для клиник: файл должен содержать 'клиник' в названии")
		h.bot.Send(msg)
		return
	}

	// Отправляем сообщение о начале обработки
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("📥 Файл '%s' получен. Начинаю обработку...", fileName))
	h.bot.Send(msg)

	// Скачиваем файл
	filePath, err := h.downloadFile(fileID, fileName)
	if err != nil {
		h.sendErrorMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка скачивания файла: %v", err))
		return
	}
	defer os.Remove(filePath) // Удаляем временный файл после обработки

	// Обрабатываем файл в зависимости от типа
	var result string
	switch importType {
	case "veterinarians":
		result, err = h.importVeterinarians(filePath, fileName)
	case "cities":
		result, err = h.importCities(filePath, fileName)
	case "clinics":
		result, err = h.importClinics(filePath, fileName)
	}

	if err != nil {
		h.sendErrorMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка импорта: %v", err))
		return
	}

	// Отправляем результат
	msg = tgbotapi.NewMessage(update.Message.Chat.ID, result)
	h.bot.Send(msg)
}

// Функция для скачивания файла
func (h *MainHandler) downloadFile(fileID string, fileName string) (string, error) {
	fileConfig := tgbotapi.FileConfig{FileID: fileID}
	file, err := h.bot.GetFile(fileConfig)
	if err != nil {
		return "", fmt.Errorf("не удалось получить файл: %v", err)
	}

	// Создаем временную директорию если нет
	tempDir := "temp"
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.Mkdir(tempDir, 0755)
	}

	// Скачиваем файл
	filePath := filepath.Join(tempDir, fileName)

	// ИСПРАВЛЕНО: Используем метод GetToken() вместо прямого доступа к полю
	token := h.bot.GetToken()
	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, file.FilePath)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("ошибка скачивания: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения файла: %v", err)
	}

	return filePath, nil
}

// Импорт врачей (обновленная версия)
func (h *MainHandler) importVeterinarians(filePath string, fileName string) (string, error) {
	var vets []models.Veterinarian
	var err error

	// Определяем тип файла и парсим
	if strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		vets, err = h.parseVeterinariansCSV(filePath)
	} else if strings.HasSuffix(strings.ToLower(fileName), ".xlsx") {
		vets, err = h.parseVeterinariansXLSX(filePath)
	} else {
		return "", fmt.Errorf("неподдерживаемый формат файла. Используйте CSV или XLSX")
	}

	if err != nil {
		return "", err
	}

	if len(vets) == 0 {
		return "⚠️ В файле не найдено данных для импорта", nil
	}

	// Сохраняем в базу
	successCount := 0
	for _, vet := range vets {
		// Сохраняем основную информацию о враче
		err := h.db.CreateVeterinarian(&vet)
		if err != nil {
			log.Printf("Ошибка сохранения врача %s %s: %v", vet.FirstName, vet.LastName, err)
			continue
		}

		// Сохраняем специализации
		for _, spec := range vet.Specializations {
			err := h.db.AddVeterinarianSpecialization(vet.ID, spec.ID)
			if err != nil {
				log.Printf("Ошибка добавления специализации для врача %d: %v", vet.ID, err)
			}
		}

		successCount++
	}

	return fmt.Sprintf("✅ Импорт завершен!\n\nОбработано записей: %d\nУспешно импортировано: %d\nОшибок: %d",
		len(vets), successCount, len(vets)-successCount), nil
}

// Парсинг CSV файла с врачами (адаптировано под ваши модели)
func (h *MainHandler) parseVeterinariansCSV(filePath string) ([]models.Veterinarian, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t' // Табуляция как разделитель
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения CSV: %v", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("файл не содержит данных")
	}

	var vets []models.Veterinarian

	for i, record := range records[1:] {
		if len(record) < 9 {
			log.Printf("Пропускаем строку %d: недостаточно данных (нужно 9 колонок, получили %d)", i+2, len(record))
			continue
		}

		// Парсим опыт работы (может быть в формате "5 лет")
		var experienceYears sql.NullInt64
		if expStr := strings.TrimSpace(record[4]); expStr != "" {
			if years, err := extractYearsFromExperience(expStr); err == nil {
				experienceYears = sql.NullInt64{Int64: int64(years), Valid: true}
			}
		}

		vet := models.Veterinarian{
			FirstName:       strings.TrimSpace(record[0]),
			LastName:        strings.TrimSpace(record[1]),
			Phone:           strings.TrimSpace(record[2]),
			Email:           sql.NullString{String: strings.TrimSpace(record[3]), Valid: record[3] != ""},
			ExperienceYears: experienceYears,
			Description:     sql.NullString{String: strings.TrimSpace(record[5]), Valid: record[5] != ""},
			IsActive:        true,
			CreatedAt:       time.Now(),
		}

		// Получаем CityID по имени города
		cityID, err := h.getOrCreateCityID(strings.TrimSpace(record[7]), strings.TrimSpace(record[8]))
		if err != nil {
			log.Printf("Ошибка получения CityID для города %s: %v", record[7], err)
			continue
		}
		vet.CityID = sql.NullInt64{Int64: int64(cityID), Valid: true}

		// Обрабатываем специализации
		if specStr := strings.TrimSpace(record[6]); specStr != "" {
			specializations, err := h.processSpecializations(specStr)
			if err != nil {
				log.Printf("Ошибка обработки специализаций для %s: %v", vet.FirstName, err)
			} else {
				vet.Specializations = specializations
			}
		}

		vets = append(vets, vet)
	}

	return vets, nil
}

// Парсинг XLSX файла с врачами (адаптировано под ваши модели)
func (h *MainHandler) parseVeterinariansXLSX(filePath string) ([]models.Veterinarian, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия XLSX файла: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("файл не содержит листов")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения листа: %v", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("файл не содержит данных")
	}

	var vets []models.Veterinarian

	for i, row := range rows[1:] {
		if len(row) < 9 {
			log.Printf("Пропускаем строку %d: недостаточно данных (нужно 9 колонок, получили %d)", i+2, len(row))
			continue
		}

		// Парсим опыт работы
		var experienceYears sql.NullInt64
		if expStr := strings.TrimSpace(row[4]); expStr != "" {
			if years, err := extractYearsFromExperience(expStr); err == nil {
				experienceYears = sql.NullInt64{Int64: int64(years), Valid: true}
			}
		}

		vet := models.Veterinarian{
			FirstName:       strings.TrimSpace(row[0]),
			LastName:        strings.TrimSpace(row[1]),
			Phone:           strings.TrimSpace(row[2]),
			Email:           sql.NullString{String: strings.TrimSpace(row[3]), Valid: row[3] != ""},
			ExperienceYears: experienceYears,
			Description:     sql.NullString{String: strings.TrimSpace(row[5]), Valid: row[5] != ""},
			IsActive:        true,
			CreatedAt:       time.Now(),
		}

		// Получаем CityID по имени города
		cityID, err := h.getOrCreateCityID(strings.TrimSpace(row[7]), strings.TrimSpace(row[8]))
		if err != nil {
			log.Printf("Ошибка получения CityID для города %s: %v", row[7], err)
			continue
		}
		vet.CityID = sql.NullInt64{Int64: int64(cityID), Valid: true}

		// Обрабатываем специализации
		if specStr := strings.TrimSpace(row[6]); specStr != "" {
			specializations, err := h.processSpecializations(specStr)
			if err != nil {
				log.Printf("Ошибка обработки специализаций для %s: %v", vet.FirstName, err)
			} else {
				vet.Specializations = specializations
			}
		}

		vets = append(vets, vet)
	}

	return vets, nil
}

// Вспомогательная функция для извлечения лет из строки опыта
func extractYearsFromExperience(expStr string) (int, error) {
	// Убираем все нецифровые символы и пытаемся извлечь число
	re := regexp.MustCompile(`\d+`)
	matches := re.FindStringSubmatch(expStr)
	if len(matches) > 0 {
		return strconv.Atoi(matches[0])
	}
	return 0, fmt.Errorf("не удалось извлечь количество лет из: %s", expStr)
}

// Обработка специализаций (разделение строки и создание объектов)
func (h *MainHandler) processSpecializations(specStr string) ([]*models.Specialization, error) {
	// Разделяем специализации по запятым, точкам с запятой или другим разделителям
	separators := []string{",", ";", "/", " и "}

	var specs []string
	for _, sep := range separators {
		if strings.Contains(specStr, sep) {
			specs = strings.Split(specStr, sep)
			break
		}
	}

	if len(specs) == 0 {
		specs = []string{specStr}
	}

	var specializations []*models.Specialization

	for _, specName := range specs {
		specName = strings.TrimSpace(specName)
		if specName == "" {
			continue
		}

		// Ищем существующую специализацию или создаем новую
		spec, err := h.getOrCreateSpecialization(specName)
		if err != nil {
			return nil, err
		}
		specializations = append(specializations, spec)
	}

	return specializations, nil
}

// Получить или создать специализацию
func (h *MainHandler) getOrCreateSpecialization(name string) (*models.Specialization, error) {
	// Сначала пытаемся найти существующую
	spec, err := h.db.GetSpecializationByName(name)
	if err == nil && spec != nil {
		return spec, nil
	}

	// Если не найдена, создаем новую
	newSpec := &models.Specialization{
		Name:      name,
		CreatedAt: time.Now(),
	}

	err = h.db.CreateSpecialization(newSpec)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания специализации %s: %v", name, err)
	}

	return newSpec, nil
}

// Получить или создать город (обновленная версия)
func (h *MainHandler) getOrCreateCityID(cityName string, region string) (int, error) {
	if cityName == "" {
		return 0, fmt.Errorf("название города не может быть пустым")
	}

	// Сначала пытаемся найти существующий город
	city, err := h.db.GetCityByName(cityName)
	if err == nil && city != nil {
		return city.ID, nil
	}

	// Если город не найден, создаем новый
	newCity := &models.City{
		Name:      cityName,
		Region:    region,
		CreatedAt: time.Now(),
	}

	err = h.db.CreateCity(newCity)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания города %s: %v", cityName, err)
	}

	return newCity.ID, nil
}

// Вспомогательная функция для отправки ошибок
func (h *MainHandler) sendErrorMessage(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, "❌ "+message)
	h.bot.Send(msg)
}

// isAdmin проверяет, является ли пользователь администратором
func (h *MainHandler) isAdmin(userID int64) bool {
	if h.config == nil || len(h.config.AdminIDs) == 0 {
		InfoLog.Printf("Config or AdminIDs is empty")
		return false
	}

	for _, adminID := range h.config.AdminIDs {
		if userID == adminID {
			InfoLog.Printf("User %d found in admin list", userID)
			return true
		}
	}

	InfoLog.Printf("User %d not found in admin list: %v", userID, h.config.AdminIDs)
	return false
}

// isInAdminMode проверяет, находится ли пользователь в админском режиме
func (h *MainHandler) isInAdminMode(userID int64) bool {
	// Защита от nil указателя
	if h.adminHandlers == nil {
		InfoLog.Printf("Admin handlers is nil for user %d", userID)
		return false
	}

	// Проверяем состояние админской сессии
	if state, exists := h.adminHandlers.adminState[userID]; exists {
		return state != ""
	}
	return false
}

// importCities и importClinics - временные заглушки
func (h *MainHandler) importCities(_ string, _ string) (string, error) {
	return "✅ Импорт городов завершен!\n\nФункция импорта городов в разработке", nil
}

func (h *MainHandler) importClinics(_ string, _ string) (string, error) {
	return "✅ Импорт клиник завершен!\n\nФункция импорта клиник в разработке", nil
}
