-- Таблица специализаций врачей
CREATE TABLE IF NOT EXISTS specializations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
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

-- Таблица запросов пользователей (вместо consultations)
CREATE TABLE IF NOT EXISTS user_requests (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    specialization_id INTEGER REFERENCES specializations(id) ON DELETE SET NULL,
    search_query TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Вставляем тестовые данные
INSERT INTO specializations (name, description) VALUES 
('Терапевт', 'Общее лечение животных'),
('Хирург', 'Проведение операций'),
('Стоматолог', 'Лечение зубов и полости рта'),
('Дерматолог', 'Лечение кожных заболеваний'),
('Офтальмолог', 'Лечение заболеваний глаз'),
('Кардиолог', 'Лечение сердечных заболеваний'),
('Ортопед', 'Лечение опорно-двигательного аппарата');

INSERT INTO veterinarians (first_name, last_name, phone, email, experience_years) VALUES 
('Иван', 'Петров', '+79161234567', 'ivan.petrov@vetclinic.ru', 10),
('Мария', 'Сидорова', '+79167654321', 'maria.sidorova@vetclinic.ru', 8),
('Алексей', 'Кузнецов', '+79169998877', 'alexey.kuznetsov@vetclinic.ru', 12);

INSERT INTO clinics (name, address, phone, working_hours) VALUES 
('ВетКлиника Центр', 'ул. Центральная, д. 1', '+74950000001', 'Пн-Пт: 9:00-21:00, Сб-Вс: 10:00-18:00'),
('ВетКлиника Север', 'ул. Северная, д. 25', '+74950000002', 'Пн-Вс: 8:00-20:00');

INSERT INTO vet_specializations (vet_id, specialization_id) VALUES 
(1, 1), (1, 2), -- Иван Петров: Терапевт, Хирург
(2, 1), (2, 4), -- Мария Сидорова: Терапевт, Дерматолог
(3, 2), (3, 6); -- Алексей Кузнецов: Хирург, Кардиолог

INSERT INTO schedules (vet_id, clinic_id, day_of_week, start_time, end_time) VALUES 
(1, 1, 1, '09:00', '15:00'),
(1, 1, 3, '09:00', '15:00'),
(1, 1, 5, '09:00', '15:00'),
(2, 1, 2, '12:00', '18:00'),
(2, 1, 4, '12:00', '18:00'),
(3, 2, 1, '10:00', '16:00'),
(3, 2, 3, '10:00', '16:00'),
(3, 2, 5, '10:00', '16:00');

-- ========== ДОБАВЛЕНИЕ ГОРОДОВ И АДРЕСОВ ==========

-- Таблица населенных пунктов
CREATE TABLE IF NOT EXISTS cities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    region VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Добавляем поля к клиникам
ALTER TABLE clinics ADD COLUMN IF NOT EXISTS city_id INTEGER REFERENCES cities(id);
ALTER TABLE clinics ADD COLUMN IF NOT EXISTS district VARCHAR(100);
ALTER TABLE clinics ADD COLUMN IF NOT EXISTS metro_station VARCHAR(100);

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
('Ростов-на-Дону', 'Южный федеральный округ');