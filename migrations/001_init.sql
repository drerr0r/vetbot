-- Таблица специализаций врачей
CREATE TABLE IF NOT EXISTS specializations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица населенных пунктов
CREATE TABLE IF NOT EXISTS cities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    region VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица ветеринарных врачей
CREATE TABLE IF NOT EXISTS veterinarians (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    description TEXT,
    experience_years INTEGER,
    city_id INTEGER REFERENCES cities(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Связь многие-ко-многим между врачами и специализациями
CREATE TABLE IF NOT EXISTS vet_specializations (
    vet_id INTEGER REFERENCES veterinarians(id) ON DELETE CASCADE,
    specialization_id INTEGER REFERENCES specializations(id) ON DELETE CASCADE,
    PRIMARY KEY (vet_id, specialization_id)
);

-- Таблица клиник/мест приема
CREATE TABLE IF NOT EXISTS clinics (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    address TEXT NOT NULL,
    phone VARCHAR(20),
    working_hours TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    city_id INTEGER REFERENCES cities(id),
    district VARCHAR(100),
    metro_station VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица расписания врачей
CREATE TABLE IF NOT EXISTS schedules (
    id SERIAL PRIMARY KEY,
    vet_id INTEGER REFERENCES veterinarians(id) ON DELETE CASCADE,
    clinic_id INTEGER REFERENCES clinics(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 1 AND 7), -- 1=Monday, 7=Sunday
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица пользователей (оставляем существующую)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(100),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица запросов пользователей 
CREATE TABLE IF NOT EXISTS user_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    specialization_id INTEGER REFERENCES specializations(id) ON DELETE SET NULL,
    search_query TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);



-- Уникальное ограничение для предотвращения дублей врачей
ALTER TABLE veterinarians ADD CONSTRAINT unique_vet_identity UNIQUE (first_name, last_name, phone);

-- Для городов 
ALTER TABLE cities ADD CONSTRAINT unique_city_name UNIQUE (name);

-- Для клиник
ALTER TABLE clinics ADD CONSTRAINT unique_clinic_address UNIQUE (name, address);

-- Для специализаций  
ALTER TABLE specializations ADD CONSTRAINT unique_specialization_name UNIQUE (name);

-- Вставляем основные города
INSERT INTO cities (name, region) VALUES 
('Москва', 'Центральный федеральный округ'),
('Санкт-Петербург', 'Северо-Западный федеральный округ'),
('Казань', 'Приволжский федеральный округ'),
('Новосибирск', 'Сибирский федеральный округ'),
('Екатеринбург', 'Уральский федеральный округ'),
('Нижний Новгород', 'Приволжский федеральный округ'),
('Краснодар', 'Южный федеральный округ'),
('Воронеж', 'Центральный федеральный округ'),
('Самара', 'Приволжский федеральный округ'),
('Ростов-на-Дону', 'Южный федеральный округ')
ON CONFLICT (name) DO NOTHING;

-- Вставляем специализации
INSERT INTO specializations (name, description) VALUES 
('Терапевт', 'Общее лечение животных'),
('Хирург', 'Проведение операций'),
('Стоматолог', 'Лечение зубов и полости рта'),
('Дерматолог', 'Лечение кожных заболеваний'),
('Офтальмолог', 'Лечение заболеваний глаз'),
('Кардиолог', 'Лечение сердечных заболеваний'),
('Ортопед', 'Лечение опорно-двигательного аппарата')
ON CONFLICT (name) DO NOTHING;

-- Вставляем тестовых врачей (привязываем к Москве)
INSERT INTO veterinarians (first_name, last_name, phone, email, experience_years, city_id) VALUES 
('Иван', 'Петров', '+79161234567', 'ivan.petrov@vetclinic.ru', 10, 1),
('Мария', 'Сидорова', '+79167654321', 'maria.sidorova@vetclinic.ru', 8, 1),
('Алексей', 'Кузнецов', '+79169998877', 'alexey.kuznetsov@vetclinic.ru', 12, 1)
ON CONFLICT (first_name, last_name, phone) DO NOTHING;

-- Вставляем клиники
INSERT INTO clinics (name, address, phone, working_hours, city_id, district, metro_station) VALUES 
('ВетКлиника Центр', 'ул. Центральная, д. 1', '+74950000001', 'Пн-Пт: 9:00-21:00, Сб-Вс: 10:00-18:00', 1, 'Центральный', 'Охотный ряд'),
('ВетКлиника Север', 'ул. Северная, д. 25', '+74950000002', 'Пн-Вс: 8:00-20:00', 1, 'Северный', 'Речной вокзал'),
('ВетКлиника Петербург', 'Невский пр-т, д. 100', '+78120000001', 'Пн-Пт: 8:00-20:00, Сб-Вс: 9:00-18:00', 2, 'Центральный', 'Невский проспект'),
('ВетКлиника Запад', 'ул. Западная, д. 50', '+74950000003', 'Пн-Вс: 9:00-19:00', 1, 'Западный', 'Кунцевская')
ON CONFLICT DO NOTHING;

-- Связываем врачей со специализациями
INSERT INTO vet_specializations (vet_id, specialization_id) VALUES 
(1, 1), (1, 2), -- Иван Петров: Терапевт, Хирург
(2, 1), (2, 4), -- Мария Сидорова: Терапевт, Дерматолог
(3, 2), (3, 6) -- Алексей Кузнецов: Хирург, Кардиолог
ON CONFLICT (vet_id, specialization_id) DO NOTHING;

-- Очищаем некорректные связи в расписаниях (если есть)
DELETE FROM schedules WHERE clinic_id NOT IN (SELECT id FROM clinics);
DELETE FROM schedules WHERE vet_id NOT IN (SELECT id FROM veterinarians);

-- Создаем расписания (только для существующих врачей и клиник)
INSERT INTO schedules (vet_id, clinic_id, day_of_week, start_time, end_time) VALUES 
(1, 1, 1, '09:00', '15:00'),
(1, 1, 3, '09:00', '15:00'),
(1, 1, 5, '09:00', '15:00'),
(2, 1, 2, '12:00', '18:00'),
(2, 1, 4, '12:00', '18:00'),
(3, 2, 1, '10:00', '16:00'),
(3, 2, 3, '10:00', '16:00'),
(3, 2, 5, '10:00', '16:00'),
(1, 3, 2, '14:00', '20:00'),
(2, 3, 5, '10:00', '16:00')
ON CONFLICT DO NOTHING;


-- Добавляем индексы для улучшения производительности поиска
CREATE INDEX IF NOT EXISTS idx_clinics_city_id ON clinics(city_id);
CREATE INDEX IF NOT EXISTS idx_clinics_district ON clinics(district);
CREATE INDEX IF NOT EXISTS idx_clinics_metro ON clinics(metro_station);
CREATE INDEX IF NOT EXISTS idx_clinics_is_active ON clinics(is_active);
CREATE INDEX IF NOT EXISTS idx_veterinarians_city_id ON veterinarians(city_id);
CREATE INDEX IF NOT EXISTS idx_veterinarians_is_active ON veterinarians(is_active);
CREATE INDEX IF NOT EXISTS idx_schedules_vet_id ON schedules(vet_id);
CREATE INDEX IF NOT EXISTS idx_schedules_clinic_id ON schedules(clinic_id);
CREATE INDEX IF NOT EXISTS idx_schedules_day ON schedules(day_of_week);
CREATE INDEX IF NOT EXISTS idx_schedules_is_available ON schedules(is_available);
CREATE INDEX IF NOT EXISTS idx_vet_specializations_vet_id ON vet_specializations(vet_id);
CREATE INDEX IF NOT EXISTS idx_vet_specializations_spec_id ON vet_specializations(specialization_id);
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);

-- ========== ОЧИСТКА ДУБЛИКАТОВ И ДОБАВЛЕНИЕ ОГРАНИЧЕНИЙ ==========

-- Удаляем дубликаты врачей, оставляя только первую запись
DELETE FROM veterinarians 
WHERE id NOT IN (
    SELECT MIN(id) 
    FROM veterinarians 
    GROUP BY first_name, last_name, phone
);

-- Удаляем дубликаты городов
DELETE FROM cities 
WHERE id NOT IN (
    SELECT MIN(id) 
    FROM cities 
    GROUP BY name
);

-- Добавляем уникальные ограничения для предотвращения будущих дублей
ALTER TABLE veterinarians ADD CONSTRAINT IF NOT EXISTS unique_vet_identity UNIQUE (first_name, last_name, phone);
ALTER TABLE cities ADD CONSTRAINT IF NOT EXISTS unique_city_name UNIQUE (name);
ALTER TABLE clinics ADD CONSTRAINT IF NOT EXISTS unique_clinic_address UNIQUE (name, address);
ALTER TABLE specializations ADD CONSTRAINT IF NOT EXISTS unique_specialization_name UNIQUE (name);

-- Обновляем существующие INSERT запросы с ON CONFLICT
INSERT INTO veterinarians (first_name, last_name, phone, email, experience_years, city_id) VALUES 
('Иван', 'Петров', '+79161234567', 'ivan.petrov@vetclinic.ru', 10, 1),
('Мария', 'Сидорова', '+79167654321', 'maria.sidorova@vetclinic.ru', 8, 1),
('Алексей', 'Кузнецов', '+79169998877', 'alexey.kuznetsov@vetclinic.ru', 12, 1)
ON CONFLICT (first_name, last_name, phone) DO NOTHING;

-- Аналогично обновите другие INSERT запросы...
INSERT INTO cities (name, region) VALUES 
('Москва', 'Центральный федеральный округ'),
('Санкт-Петербург', 'Северо-Западный федеральный округ')
-- ... остальные города
ON CONFLICT (name) DO NOTHING;