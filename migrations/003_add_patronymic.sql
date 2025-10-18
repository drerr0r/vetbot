-- Добавляем поле отчества к врачам
ALTER TABLE veterinarians ADD COLUMN IF NOT EXISTS patronymic VARCHAR(100);

-- Обновляем индексы если нужно