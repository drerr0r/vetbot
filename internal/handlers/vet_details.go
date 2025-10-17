package handlers

import (
	"fmt"
	"strings"

	"github.com/drerr0r/vetbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleVetDetails отображает детальную информацию о враче
func (h *VetHandlers) HandleVetDetails(chatID int64, vetID int, messageID int) error {
	InfoLog.Printf("HandleVetDetails called for vet ID: %d", vetID)

	// Получаем полную информацию о враче
	vet, err := h.db.GetVeterinarianWithDetails(vetID)
	if err != nil {
		ErrorLog.Printf("Error getting vet details: %v", err)
		return fmt.Errorf("failed to get vet details: %v", err)
	}

	// Формируем сообщение с полной информацией
	message := h.formatVeterinarianDetails(vet)

	// Создаем клавиатуру с закрепленными кнопками
	replyMarkup := h.createVetDetailsKeyboard(vetID)

	// Если есть предыдущее сообщение, редактируем его
	if messageID != 0 {
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, message)
		editMsg.ParseMode = "Markdown"
		editMsg.ReplyMarkup = &replyMarkup
		_, err = h.bot.Send(editMsg)
	} else {
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = replyMarkup
		_, err = h.bot.Send(msg)
	}

	return err
}

// formatVeterinarianDetails форматирует детальную информацию о враче
func (h *VetHandlers) formatVeterinarianDetails(vet *models.Veterinarian) string {
	var message strings.Builder

	message.WriteString("🐾 *Детальная информация о враче*\n\n")

	// Основная информация
	message.WriteString(fmt.Sprintf("👨‍⚕️ *%s %s*\n", vet.FirstName, vet.LastName))

	// Получаем статистику отзывов
	stats, err := h.db.GetReviewStats(models.GetVetIDAsIntOrZero(vet))
	if err == nil {
		if stats.ApprovedReviews > 0 {
			message.WriteString(fmt.Sprintf("⭐ *Рейтинг:* %.1f/5 (%d отзывов)\n", stats.AverageRating, stats.ApprovedReviews))
		} else {
			message.WriteString("⭐ *Рейтинг:* пока нет отзывов\n")
		}
	}

	if vet.Phone != "" {
		message.WriteString(fmt.Sprintf("📞 *Телефон:* `%s`\n", vet.Phone))
	}

	if vet.Email.Valid && vet.Email.String != "" {
		message.WriteString(fmt.Sprintf("📧 *Email:* %s\n", vet.Email.String))
	}

	if vet.ExperienceYears.Valid {
		message.WriteString(fmt.Sprintf("⏳ *Опыт работы:* %d лет\n", vet.ExperienceYears.Int64))
	}

	if vet.Description.Valid && vet.Description.String != "" {
		message.WriteString(fmt.Sprintf("📝 *Описание:* %s\n", vet.Description.String))
	}

	// Информация о городе
	if vet.City != nil {
		message.WriteString(fmt.Sprintf("🏙️ *Город:* %s", vet.City.Name))
		if vet.City.Region != "" {
			message.WriteString(fmt.Sprintf(" (%s)", vet.City.Region))
		}
		message.WriteString("\n")
	}

	// Специализации
	if len(vet.Specializations) > 0 {
		message.WriteString("🎯 *Специализации:* ")
		specs := make([]string, len(vet.Specializations))
		for i, spec := range vet.Specializations {
			specs[i] = spec.Name
		}
		message.WriteString(strings.Join(specs, ", "))
		message.WriteString("\n")
	}

	// Клиники и расписание
	if len(vet.Schedules) > 0 {
		message.WriteString("\n🏥 *Места приема и расписание:*\n")

		// Группируем по клиникам
		clinicSchedules := make(map[string][]*models.Schedule)
		for _, schedule := range vet.Schedules {
			if schedule.Clinic != nil {
				clinicName := schedule.Clinic.Name
				clinicSchedules[clinicName] = append(clinicSchedules[clinicName], schedule)
			}
		}

		for clinicName, schedules := range clinicSchedules {
			message.WriteString(fmt.Sprintf("\n*%s*\n", clinicName))

			// Адрес и контакты клиники
			if len(schedules) > 0 && schedules[0].Clinic != nil {
				clinic := schedules[0].Clinic
				message.WriteString(fmt.Sprintf("📍 *Адрес:* %s\n", clinic.Address))

				// Информация о метро и районе
				if clinic.MetroStation.Valid && clinic.MetroStation.String != "" {
					message.WriteString(fmt.Sprintf("🚇 *Метро:* %s\n", clinic.MetroStation.String))
				}
				if clinic.District.Valid && clinic.District.String != "" {
					message.WriteString(fmt.Sprintf("🏘️ *Район:* %s\n", clinic.District.String))
				}

				if clinic.Phone.Valid && clinic.Phone.String != "" {
					message.WriteString(fmt.Sprintf("📞 *Телефон клиники:* %s\n", clinic.Phone.String))
				}

				if clinic.WorkingHours.Valid && clinic.WorkingHours.String != "" {
					message.WriteString(fmt.Sprintf("🕐 *Часы работы:* %s\n", clinic.WorkingHours.String))
				}
			}

			// Расписание по дням
			daysMap := make(map[int][]string)
			for _, schedule := range schedules {
				_ = getDayName(schedule.DayOfWeek)
				timeSlot := fmt.Sprintf("%s-%s", schedule.StartTime, schedule.EndTime)
				daysMap[schedule.DayOfWeek] = append(daysMap[schedule.DayOfWeek], timeSlot)
			}

			if len(daysMap) > 0 {
				message.WriteString("📅 *Расписание приема:*\n")
				var scheduleParts []string
				for day := 1; day <= 7; day++ {
					if times, exists := daysMap[day]; exists {
						dayName := getDayName(day)
						scheduleParts = append(scheduleParts, fmt.Sprintf("   • %s: %s", dayName, strings.Join(times, ", ")))
					}
				}
				message.WriteString(strings.Join(scheduleParts, "\n"))
			}
			message.WriteString("\n")
		}
	} else {
		message.WriteString("\n📅 *Расписание:* не указано\n")
	}

	return message.String()
}

// createVetDetailsKeyboard создает клавиатуру для детального просмотра врача
func (h *VetHandlers) createVetDetailsKeyboard(vetID int) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Отзывы", fmt.Sprintf("show_reviews_%d", vetID)),
			tgbotapi.NewInlineKeyboardButtonData("💬 Оставить отзыв", fmt.Sprintf("add_review_%d", vetID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Записаться", fmt.Sprintf("appointment_%d", vetID)),
			tgbotapi.NewInlineKeyboardButtonData("⭐ В избранное", fmt.Sprintf("favorite_%d", vetID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 Поиск врачей", "main_menu"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}
