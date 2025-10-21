package handlers

import (
	"log"

	"github.com/drerr0r/vetbot/internal/models"
	"github.com/xuri/excelize/v2"
)

type TemplateGenerator struct {
	db Database
}

func NewTemplateGenerator(db Database) *TemplateGenerator {
	return &TemplateGenerator{db: db}
}

// GenerateTemplate создает шаблон Excel для импорта врачей
func (tg *TemplateGenerator) GenerateTemplate(filepath string) error {
	f := excelize.NewFile()

	// Создаем основной лист с данными
	sheetName := "Врачи"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)

	// Удаляем дефолтный лист
	if err := f.DeleteSheet("Sheet1"); err != nil {
		log.Printf("Warning: failed to delete default sheet: %v", err)
	}

	// Заголовки колонок
	headers := []string{
		"Имя",
		"Фамилия",
		"Телефон",
		"Email",
		"ОпытРаботы",
		"Описание",
		"Город",
		"Специализации",
		"КлиникиИРасписание",
		"ПримерыЗаполнения",
	}

	// Устанавливаем заголовки
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return err
		}
	}

	// Получаем справочные данные
	cities, _ := tg.db.GetAllCities()
	specializations, _ := tg.db.GetAllSpecializations()
	clinics, _ := tg.db.GetAllClinics()

	// Добавляем примеры данных
	examples := []map[string]string{
		{
			"Имя":                "Иван",
			"Фамилия":            "Петров",
			"Телефон":            "+79161234567",
			"Email":              "ivan.petrov@vetclinic.ru",
			"ОпытРаботы":         "10",
			"Описание":           "Опытный терапевт, специалист по мелким животным",
			"Город":              "Москва",
			"Специализации":      "Терапевт, Хирург",
			"КлиникиИРасписание": "ВетКлиника Центр:Пн:9-18,Ср:9-18,Пт:9-18;ВетКлиника Север:Вт:10-19,Чт:10-19",
			"ПримерыЗаполнения":  "✅ Корректный пример",
		},
		{
			"Имя":                "Мария",
			"Фамилия":            "Сидорова",
			"Телефон":            "+79167654321",
			"Email":              "maria.sidorova@vetclinic.ru",
			"ОпытРаботы":         "8",
			"Описание":           "Дерматолог, аллерголог",
			"Город":              "Москва",
			"Специализации":      "Дерматолог",
			"КлиникиИРасписание": "ВетКлиника Центр:Пн:12-20,Ср:12-20",
			"ПримерыЗаполнения":  "✅ Корректный пример",
		},
	}

	// Заполняем примеры
	for row, example := range examples {
		for col, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			if err := f.SetCellValue(sheetName, cell, example[header]); err != nil {
				return err
			}
		}
	}

	// Создаем лист со справочниками
	if err := tg.createReferenceSheet(f, "Справочники", cities, specializations, clinics); err != nil {
		return err
	}

	// Создаем лист с инструкцией
	if err := tg.createInstructionSheet(f, "Инструкция"); err != nil {
		return err
	}

	// Настраиваем ширину колонок
	tg.setColumnWidths(f, sheetName)

	// Сохраняем файл
	return f.SaveAs(filepath)
}

// Исправляем сигнатуру метода для работы с указателями
func (tg *TemplateGenerator) createReferenceSheet(f *excelize.File, sheetName string, cities []*models.City, specs []*models.Specialization, clinics []*models.Clinic) error {
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Города
	if err := f.SetCellValue(sheetName, "A1", "Доступные города:"); err != nil {
		return err
	}
	for i, city := range cities {
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		if err := f.SetCellValue(sheetName, cell, city.Name); err != nil {
			return err
		}
	}

	// Специализации
	if err := f.SetCellValue(sheetName, "C1", "Доступные специализации:"); err != nil {
		return err
	}
	for i, spec := range specs {
		cell, _ := excelize.CoordinatesToCellName(3, i+2)
		if err := f.SetCellValue(sheetName, cell, spec.Name); err != nil {
			return err
		}
	}

	// Клиники
	if err := f.SetCellValue(sheetName, "E1", "Доступные клиники:"); err != nil {
		return err
	}
	for i, clinic := range clinics {
		cell, _ := excelize.CoordinatesToCellName(5, i+2)
		if err := f.SetCellValue(sheetName, cell, clinic.Name); err != nil {
			return err
		}
	}

	f.SetActiveSheet(index)
	return nil
}

func (tg *TemplateGenerator) createInstructionSheet(f *excelize.File, sheetName string) error {
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	instructions := []string{
		"ИНСТРУКЦИЯ ПО ЗАПОЛНЕНИЮ",
		"",
		"1. Заполняйте данные только на листе 'Врачи'",
		"2. Используйте данные из листа 'Справочники' для корректного заполнения",
		"3. Обязательные поля: Имя, Фамилия, Телефон, Город",
		"4. Формат телефона: +79161234567",
		"5. Опыт работы: указывается в годах (только цифры)",
		"",
		"ФОРМАТЫ ДАННЫХ:",
		"- Специализации: перечисляются через запятую - 'Терапевт, Хирург'",
		"- Клиники и расписание: формат 'НазваниеКлиники:День:Часы;ДругаяКлиника:День:Часы'",
		"- Пример: 'ВетКлиника Центр:Пн:9-18,Ср:9-18;ВетКлиника Север:Вт:10-19'",
		"",
		"ОБОЗНАЧЕНИЯ ДНЕЙ:",
		"- Пн, Вт, Ср, Чт, Пт, Сб, Вс",
		"- или: Понедельник, Вторник, Среда, Четверг, Пятница, Суббота, Воскресенье",
		"",
		"После заполнения сохраните файл и импортируйте в систему.",
	}

	for i, instruction := range instructions {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetCellValue(sheetName, cell, instruction); err != nil {
			return err
		}
	}

	f.SetActiveSheet(index)
	return nil
}

func (tg *TemplateGenerator) setColumnWidths(f *excelize.File, sheetName string) {
	widths := map[string]float64{
		"A": 15, // Имя
		"B": 15, // Фамилия
		"C": 20, // Телефон
		"D": 25, // Email
		"E": 12, // Опыт
		"F": 30, // Описание
		"G": 15, // Город
		"H": 20, // Специализации
		"I": 40, // Клиники и расписание
		"J": 25, // Примеры
	}

	for col, width := range widths {
		if err := f.SetColWidth(sheetName, col, col, width); err != nil {
			log.Printf("Warning: failed to set column width for %s: %v", col, err)
		}
	}
}
