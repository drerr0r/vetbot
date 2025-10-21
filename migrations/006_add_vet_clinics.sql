-- Таблица связи многие-ко-многим между врачами и клиниками
CREATE TABLE IF NOT EXISTS vet_clinics (
    vet_id INTEGER REFERENCES veterinarians(id) ON DELETE CASCADE,
    clinic_id INTEGER REFERENCES clinics(id) ON DELETE CASCADE,
    PRIMARY KEY (vet_id, clinic_id)
);

-- Индексы для улучшения производительности
CREATE INDEX IF NOT EXISTS idx_vet_clinics_vet_id ON vet_clinics(vet_id);
CREATE INDEX IF NOT EXISTS idx_vet_clinics_clinic_id ON vet_clinics(clinic_id);

-- Добавляем существующих связей из расписания в новую таблицу
INSERT INTO vet_clinics (vet_id, clinic_id)
SELECT DISTINCT vet_id, clinic_id 
FROM schedules 
WHERE vet_id IS NOT NULL AND clinic_id IS NOT NULL
ON CONFLICT (vet_id, clinic_id) DO NOTHING;