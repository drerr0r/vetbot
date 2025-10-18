package handlers

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/drerr0r/vetbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ReviewHandlers содержит обработчики для системы отзывов
type ReviewHandlers struct {
	bot          BotAPI
	db           Database
	adminIDs     []int64
	stateManager *StateManager
}

// NewReviewHandlers создает новый экземпляр ReviewHandlers
func NewReviewHandlers(bot BotAPI, db Database, adminIDs []int64, stateManager *StateManager) *ReviewHandlers {
	return &ReviewHandlers{
		bot:          bot,
		db:           db,
		adminIDs:     adminIDs,
		stateManager: stateManager,
	}
}

// HandleReviewCancel обрабатывает отмену добавления отзыва
func (h *ReviewHandlers) HandleReviewCancel(update tgbotapi.Update) {
	userID := update.CallbackQuery.From.ID

	// Очищаем состояние и данные
	h.stateManager.ClearUserState(userID)
	h.stateManager.ClearUserData(userID)

	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "❌ Добавление отзыва отменено.")
	h.bot.Send(msg)

	// Отвечаем на callback
	callbackConfig := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	h.bot.Request(callbackConfig)
}

// HandleAddReview начинает процесс добавления отзыва
func (h *ReviewHandlers) HandleAddReview(update tgbotapi.Update, vetID int) {
	if update.CallbackQuery == nil || update.CallbackQuery.Message == nil {
		ErrorLog.Printf("HandleAddReview: CallbackQuery or Message is nil")
		return
	}

	chatID := update.CallbackQuery.Message.Chat.ID
	userID := update.CallbackQuery.From.ID

	// Проверяем, оставлял ли пользователь уже отзыв этому врачу
	hasReview, err := h.db.HasUserReviewForVet(int(userID), vetID)
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка проверки отзывов")
		return
	}

	if hasReview {
		msg := tgbotapi.NewMessage(chatID,
			"❌ Вы уже оставляли отзыв этому врачу. Вы можете отредактировать существующий отзыв.")
		h.bot.Send(msg)
		return
	}

	// Сохраняем данные для процесса добавления отзыва
	h.stateManager.SetUserData(userID, "review_vet_id", vetID)
	h.stateManager.SetUserState(userID, "review_rating")

	// Показываем выбор рейтинга
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐", "review_rate_1"),
			tgbotapi.NewInlineKeyboardButtonData("⭐⭐", "review_rate_2"),
			tgbotapi.NewInlineKeyboardButtonData("⭐⭐⭐", "review_rate_3"),
			tgbotapi.NewInlineKeyboardButtonData("⭐⭐⭐⭐", "review_rate_4"),
			tgbotapi.NewInlineKeyboardButtonData("⭐⭐⭐⭐⭐", "review_rate_5"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "review_cancel"),
		),
	)

	msg := tgbotapi.NewMessage(chatID,
		"📝 *Добавление отзыва*\n\nВыберите оценку врачу (1-5 звезд):")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

// В HandleReviewRating добавьте логирование:
func (h *ReviewHandlers) HandleReviewRating(update tgbotapi.Update, rating int) {
	callback := update.CallbackQuery
	chatID := callback.Message.Chat.ID
	userID := callback.From.ID

	InfoLog.Printf("HandleReviewRating: user %d selected rating %d", userID, rating)

	// Сохраняем рейтинг и переходим к следующему шагу
	h.stateManager.SetUserData(userID, "review_rating", rating)
	h.stateManager.SetUserState(userID, "review_comment")

	InfoLog.Printf("HandleReviewRating: user %d state set to 'review_comment'", userID)

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID,
		fmt.Sprintf("📝 *Добавление отзыва*\n\n✅ Оценка: %d/5 ⭐\n\nТеперь напишите ваш отзыв (максимум 500 символов):", rating))
	editMsg.ParseMode = "Markdown"

	// Убираем клавиатуру - передаем nil вместо указателя
	editMsg.ReplyMarkup = nil

	_, err := h.bot.Send(editMsg)
	if err != nil {
		ErrorLog.Printf("Error editing message in HandleReviewRating: %v", err)
	}

	// Отвечаем на callback
	callbackConfig := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("✅ Выбрано %d звезд", rating))
	h.bot.Request(callbackConfig)
}

// В HandleReviewComment добавьте дополнительное логирование перед сохранением:
func (h *ReviewHandlers) HandleReviewComment(update tgbotapi.Update, comment string) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	InfoLog.Printf("HandleReviewComment: user %d submitted comment: %s", userID, comment)

	// Проверяем длину комментария
	if len(comment) > 500 {
		msg := tgbotapi.NewMessage(chatID, "❌ Отзыв слишком длинный (максимум 500 символов). Сократите текст и отправьте снова.")
		h.bot.Send(msg)
		return
	}

	// Получаем сохраненные данные
	vetID, ok := h.stateManager.GetUserDataInt(userID, "review_vet_id")
	if !ok {
		ErrorLog.Printf("HandleReviewComment: review_vet_id not found for user %d", userID)
		h.sendErrorMessage(chatID, "Ошибка: данные о враче не найдены")
		h.stateManager.ClearUserState(userID)
		return
	}

	rating, ok := h.stateManager.GetUserDataInt(userID, "review_rating")
	if !ok {
		ErrorLog.Printf("HandleReviewComment: review_rating not found for user %d", userID)
		h.sendErrorMessage(chatID, "Ошибка: данные о рейтинге не найдены")
		h.stateManager.ClearUserState(userID)
		return
	}

	InfoLog.Printf("HandleReviewComment: user %d, vetID %d, rating %d, comment length %d",
		userID, vetID, rating, len(comment))

	// Получаем ID пользователя из базы
	user, err := h.db.GetUserByTelegramID(userID)
	if err != nil {
		ErrorLog.Printf("HandleReviewComment: user not found in database: %v", err)
		h.sendErrorMessage(chatID, "Ошибка: пользователь не найден")
		h.stateManager.ClearUserState(userID)
		return
	}

	// Проверяем, не оставлял ли пользователь уже отзыв этому врачу
	hasReview, err := h.db.HasUserReviewForVet(user.ID, vetID)
	if err != nil {
		ErrorLog.Printf("HandleReviewComment: error checking existing review: %v", err)
		h.sendErrorMessage(chatID, "Ошибка проверки существующих отзывов")
		h.stateManager.ClearUserState(userID)
		return
	}

	if hasReview {
		ErrorLog.Printf("HandleReviewComment: user %d already has review for vet %d", user.ID, vetID)
		h.sendErrorMessage(chatID, "❌ Вы уже оставляли отзыв этому врачу.")
		h.stateManager.ClearUserState(userID)
		return
	}

	// Создаем отзыв
	review := &models.Review{
		VeterinarianID: vetID,
		UserID:         user.ID,
		Rating:         rating,
		Comment:        strings.TrimSpace(comment),
		Status:         "pending", // На модерации
		CreatedAt:      time.Now(),
	}

	// Сохраняем в базу
	err = h.db.CreateReview(review)
	if err != nil {
		ErrorLog.Printf("HandleReviewComment: error saving review: %v", err)
		h.sendErrorMessage(chatID, "❌ Ошибка при сохранении отзыва")
		h.stateManager.ClearUserState(userID)
		return
	}

	// Очищаем состояние и данные
	h.stateManager.ClearUserState(userID)
	h.stateManager.ClearUserData(userID)

	InfoLog.Printf("HandleReviewComment: review saved successfully for user %d, review ID: %d", userID, review.ID)

	// Отправляем подтверждение
	msg := tgbotapi.NewMessage(chatID,
		"✅ *Отзыв успешно отправлен!*\n\nВаш отзыв будет опубликован после проверки модератором. Спасибо за ваш вклад!")
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)

	// Уведомляем администраторов о новом отзыве
	h.notifyAdminsAboutNewReview(review)
}

// HandleShowReviews показывает отзывы о враче
func (h *ReviewHandlers) HandleShowReviews(update tgbotapi.Update, vetID int) {
	chatID := update.CallbackQuery.Message.Chat.ID

	// Получаем одобренные отзывы
	reviews, err := h.db.GetApprovedReviewsByVet(vetID)
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка при загрузке отзывов")
		return
	}

	// Получаем статистику
	stats, err := h.db.GetReviewStats(vetID)
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка при загрузке статистики")
		return
	}

	var message strings.Builder

	if len(reviews) == 0 {
		message.WriteString("📝 *Отзывы о враче*\n\n")
		message.WriteString("Пока нет одобренных отзывов.\n\n")
	} else {
		message.WriteString(fmt.Sprintf("📝 *Отзывы о враче*\n\n⭐ Средняя оценка: %.1f/5\n📊 Всего отзывов: %d\n\n",
			stats.AverageRating, stats.ApprovedReviews))

		for i, review := range reviews {
			if i >= 10 { // Ограничиваем показ 10 отзывами
				message.WriteString(fmt.Sprintf("\n... и еще %d отзывов", len(reviews)-10))
				break
			}

			message.WriteString(fmt.Sprintf("**%d. %s** ⭐\n", i+1, strings.Repeat("⭐", review.Rating)))
			message.WriteString(fmt.Sprintf("💬 %s\n", html.EscapeString(review.Comment)))
			if review.User != nil {
				message.WriteString(fmt.Sprintf("👤 %s\n", html.EscapeString(review.User.FirstName)))
			}
			message.WriteString(fmt.Sprintf("📅 %s\n\n", review.CreatedAt.Format("02.01.2006")))
		}
	}

	// Добавляем кнопки
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Добавить отзыв", fmt.Sprintf("add_review_%d", vetID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад к врачу", fmt.Sprintf("vet_details_%d", vetID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)

	// Отвечаем на callback
	if update.CallbackQuery != nil {
		callbackConfig := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		h.bot.Request(callbackConfig)
	}
}

// HandleReviewModeration показывает меню модерации отзывов для администраторов
func (h *ReviewHandlers) HandleReviewModeration(update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	// Проверяем права администратора
	if !h.isAdmin(userID) {
		msg := tgbotapi.NewMessage(chatID, "❌ Эта функция доступна только администраторам")
		h.bot.Send(msg)
		return
	}

	// Получаем отзывы на модерации
	pendingReviews, err := h.db.GetPendingReviews()
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка при загрузке отзывов")
		return
	}

	if len(pendingReviews) == 0 {
		msg := tgbotapi.NewMessage(chatID, "✅ Нет отзывов, ожидающих модерации.")
		h.bot.Send(msg)
		return
	}

	// Сохраняем список отзывов для модерации
	h.stateManager.SetUserData(userID, "pending_reviews", pendingReviews)
	h.stateManager.SetUserState(userID, "review_moderation")

	var message strings.Builder
	message.WriteString("⚡ *Модерация отзывов*\n\n")
	message.WriteString(fmt.Sprintf("Отзывов на модерации: %d\n\n", len(pendingReviews)))

	for i, review := range pendingReviews {
		if i >= 5 { // Показываем первые 5
			message.WriteString(fmt.Sprintf("\n... и еще %d отзывов", len(pendingReviews)-5))
			break
		}

		message.WriteString(fmt.Sprintf("**%d. %s %s**\n", i+1,
			html.EscapeString(review.Veterinarian.FirstName),
			html.EscapeString(review.Veterinarian.LastName)))
		message.WriteString(fmt.Sprintf("⭐ Оценка: %d/5\n", review.Rating))
		message.WriteString(fmt.Sprintf("💬 Отзыв: %s\n", html.EscapeString(review.Comment)))
		if review.User != nil {
			message.WriteString(fmt.Sprintf("👤 Пользователь: %s\n", html.EscapeString(review.User.FirstName)))
		}
		message.WriteString(fmt.Sprintf("📅 Дата: %s\n", review.CreatedAt.Format("02.01.2006")))
		message.WriteString(fmt.Sprintf("🆔 ID отзыва: %d\n\n", review.ID))
	}

	message.WriteString("Введите ID отзыва для модерации:")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад в админку"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

// HandleReviewModerationAction обрабатывает действия модерации
func (h *ReviewHandlers) HandleReviewModerationAction(update tgbotapi.Update, reviewID int) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	// Получаем отзыв
	review, err := h.db.GetReviewByID(reviewID)
	if err != nil {
		h.sendErrorMessage(chatID, "Отзыв не найден")
		return
	}

	// Сохраняем данные для подтверждения
	h.stateManager.SetUserData(userID, "moderation_review", review)
	h.stateManager.SetUserState(userID, "review_moderation_confirm")

	var message strings.Builder
	message.WriteString("⚡ *Модерация отзыва*\n\n")
	message.WriteString(fmt.Sprintf("**Врач:** %s %s\n",
		html.EscapeString(review.Veterinarian.FirstName),
		html.EscapeString(review.Veterinarian.LastName)))
	message.WriteString(fmt.Sprintf("**Оценка:** %d/5 ⭐\n", review.Rating))
	message.WriteString(fmt.Sprintf("**Отзыв:** %s\n", html.EscapeString(review.Comment)))
	if review.User != nil {
		message.WriteString(fmt.Sprintf("**Пользователь:** %s\n", html.EscapeString(review.User.FirstName)))
	}
	message.WriteString(fmt.Sprintf("**Дата:** %s\n\n", review.CreatedAt.Format("02.01.2006")))

	message.WriteString("Выберите действие:")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Одобрить отзыв"),
			tgbotapi.NewKeyboardButton("❌ Отклонить отзыв"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 Назад к списку"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

// HandleReviewModerationConfirm обрабатывает подтверждение модерации
func (h *ReviewHandlers) HandleReviewModerationConfirm(update tgbotapi.Update, action string) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	// Получаем данные отзыва
	reviewInterface := h.stateManager.GetUserData(userID, "moderation_review")
	review, ok := reviewInterface.(*models.Review)
	if !ok {
		h.sendErrorMessage(chatID, "Ошибка: данные отзыва не найдены")
		return
	}

	// Получаем ID модератора из базы
	moderator, err := h.db.GetUserByTelegramID(userID)
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка: модератор не найден")
		return
	}

	var status string
	var message string

	switch action {
	case "✅ Одобрить отзыв":
		status = "approved"
		message = "✅ Отзыв одобрен и опубликован!"
	case "❌ Отклонить отзыв":
		status = "rejected"
		message = "❌ Отзыв отклонен."
	default:
		h.sendErrorMessage(chatID, "Неизвестное действие")
		return
	}

	// Обновляем статус отзыва
	err = h.db.UpdateReviewStatus(review.ID, status, moderator.ID)
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка при обновлении статуса отзыва")
		return
	}

	// Очищаем временные данные
	h.stateManager.ClearUserData(userID)
	h.stateManager.SetUserState(userID, "review_moderation")

	// Отправляем подтверждение
	msg := tgbotapi.NewMessage(chatID, message)
	h.bot.Send(msg)

	// Показываем следующий отзыв или возвращаем в меню
	h.HandleReviewModeration(update)
}

// ========== ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ==========

func (h *ReviewHandlers) notifyAdminsAboutNewReview(review *models.Review) {

	if review == nil || review.Veterinarian == nil {
		ErrorLog.Printf("notifyAdminsAboutNewReview: review or veterinarian is nil")
		return
	}

	// Реализация уведомления администраторов
	for _, adminID := range h.adminIDs {
		msg := tgbotapi.NewMessage(adminID,
			fmt.Sprintf("⚡ *Новый отзыв на модерацию!*\n\nВрач: %s %s\nОценка: %d/5 ⭐\nОтзыв: %s",
				html.EscapeString(review.Veterinarian.FirstName),
				html.EscapeString(review.Veterinarian.LastName),
				review.Rating,
				html.EscapeString(review.Comment)))
		msg.ParseMode = "Markdown"
		h.bot.Send(msg)
	}
}
func (h *ReviewHandlers) isAdmin(userID int64) bool {
	for _, adminID := range h.adminIDs {
		if userID == adminID {
			return true
		}
	}
	return false
}

func (h *ReviewHandlers) sendErrorMessage(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, "❌ "+message)
	h.bot.Send(msg)
}

// HandleReviewModerationInput обрабатывает текстовый ввод при модерации отзывов
func (h *ReviewHandlers) HandleReviewModerationInput(update tgbotapi.Update) {
	userID := update.Message.From.ID
	text := strings.TrimSpace(update.Message.Text)

	InfoLog.Printf("ReviewModerationInput: user %d, text: '%s'", userID, text)

	// Проверяем кнопки действий
	switch text {
	case "✅ Одобрить отзыв":
		h.HandleReviewModerationConfirm(update, "✅ Одобрить отзыв")
		return
	case "❌ Отклонить отзыв":
		h.HandleReviewModerationConfirm(update, "❌ Отклонить отзыв")
		return
	case "🔙 Назад к списку":
		h.HandleReviewModeration(update)
		return
	}

	// Если не кнопка, пытаемся распарсить ID отзыва
	reviewID, err := strconv.Atoi(text)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"❌ Неверный формат ID отзыва. Введите числовой ID отзыва.")
		h.bot.Send(msg)
		return
	}

	// Получаем отзыв по ID (проверяем существование)
	_, err = h.db.GetReviewByID(reviewID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			fmt.Sprintf("❌ Отзыв с ID %d не найден.", reviewID))
		h.bot.Send(msg)
		h.HandleReviewModeration(update) // Показываем список снова
		return
	}

	// Показываем детали отзыва и кнопки действий
	h.HandleReviewModerationAction(update, reviewID)
}

// // approveReview одобряет отзыв
// func (h *ReviewHandlers) approveReview(update tgbotapi.Update) {
// 	userID := update.Message.From.ID
// 	userIDStr := strconv.FormatInt(userID, 10)

// 	reviewID, ok := h.tempData[userIDStr+"_review_action"].(int)
// 	if !ok {
// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Не найден активный отзыв для модерации")
// 		h.bot.Send(msg)
// 		return
// 	}

// 	// Получаем ID модератора из базы
// 	moderator, err := h.db.GetUserByTelegramID(userID)
// 	if err != nil {
// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Ошибка: модератор не найден")
// 		h.bot.Send(msg)
// 		return
// 	}

// 	err = h.db.UpdateReviewStatus(reviewID, "approved", moderator.ID) // Добавлен moderator.ID
// 	if err != nil {
// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
// 			fmt.Sprintf("❌ Ошибка при одобрении отзыва: %v", err))
// 		h.bot.Send(msg)
// 		return
// 	}

// 	// Очищаем временные данные
// 	delete(h.tempData, userIDStr+"_review_action")

// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Отзыв успешно одобрен!")
// 	h.bot.Send(msg)

// 	// Возвращаем к списку отзывов
// 	h.HandleReviewModeration(update)
// }

// // rejectReview отклоняет отзыв
// func (h *ReviewHandlers) rejectReview(update tgbotapi.Update) {
// 	userID := update.Message.From.ID
// 	userIDStr := strconv.FormatInt(userID, 10)

// 	reviewID, ok := h.tempData[userIDStr+"_review_action"].(int)
// 	if !ok {
// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Не найден активный отзыв для модерации")
// 		h.bot.Send(msg)
// 		return
// 	}

// 	// Получаем ID модератора из базы
// 	moderator, err := h.db.GetUserByTelegramID(userID)
// 	if err != nil {
// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Ошибка: модератор не найден")
// 		h.bot.Send(msg)
// 		return
// 	}

// 	err = h.db.UpdateReviewStatus(reviewID, "rejected", moderator.ID) // Добавлен moderator.ID
// 	if err != nil {
// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
// 			fmt.Sprintf("❌ Ошибка при отклонении отзыва: %v", err))
// 		h.bot.Send(msg)
// 		return
// 	}

// 	// Очищаем временные данные
// 	delete(h.tempData, userIDStr+"_review_action")

// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Отзыв отклонен!")
// 	h.bot.Send(msg)

// 	// Возвращаем к списку отзывов
// 	h.HandleReviewModeration(update)
// }

// showReviewForModeration показывает отзыв с кнопками одобрить/отклонить
// func (h *ReviewHandlers) showReviewForModeration(update tgbotapi.Update, review *models.Review) {
// 	var message strings.Builder

// 	message.WriteString("📝 *Отзыв для модерации*\n\n")
// 	message.WriteString(fmt.Sprintf("🆔 ID: %d\n", review.ID))

// 	// Получаем информацию о пользователе
// 	user, err := h.db.GetUserByID(review.UserID)
// 	if err == nil && user != nil {
// 		message.WriteString(fmt.Sprintf("👤 Пользователь: %s\n", user.FirstName))
// 	} else {
// 		message.WriteString(fmt.Sprintf("👤 Пользователь ID: %d\n", review.UserID))
// 	}

// 	// Получаем информацию о враче
// 	vet, err := h.db.GetVeterinarianByID(review.VeterinarianID)
// 	if err == nil && vet != nil {
// 		message.WriteString(fmt.Sprintf("👨‍⚕️ Врач: %s %s\n", vet.FirstName, vet.LastName))
// 	} else {
// 		message.WriteString(fmt.Sprintf("👨‍⚕️ Врач ID: %d\n", review.VeterinarianID))
// 	}

// 	message.WriteString(fmt.Sprintf("⭐ Оценка: %d/5\n", review.Rating))
// 	message.WriteString(fmt.Sprintf("💬 Текст: %s\n", review.Comment))
// 	message.WriteString(fmt.Sprintf("📅 Дата: %s\n", review.CreatedAt.Format("02.01.2006 15:04")))

// 	keyboard := tgbotapi.NewReplyKeyboard(
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton("✅ Одобрить"),
// 			tgbotapi.NewKeyboardButton("❌ Отклонить"),
// 		),
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton("🔙 Назад к списку"),
// 		),
// 	)

// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message.String())
// 	msg.ParseMode = "Markdown"
// 	msg.ReplyMarkup = keyboard

// 	// Сохраняем ID отзыва для дальнейших действий (используем временные данные вместо adminState)
// 	userID := update.Message.From.ID
// 	userIDStr := strconv.FormatInt(userID, 10)
// 	h.tempData[userIDStr+"_review_action"] = review.ID

// 	h.bot.Send(msg)
// }
