-- Добавляем поле отчества к врачам
ALTER TABLE veterinarians ADD COLUMN patronymic VARCHAR(100);

-- Обновляем индексы если нужно