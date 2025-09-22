-- Миграция 001: Инициализация базы данных VetBot
-- Создание таблиц и начальных данных

-- Таблица пользователей Telegram бота
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    chat_id BIGINT UNIQUE NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Индексы для быстрого поиска
    CONSTRAINT unique_chat_id UNIQUE (chat_id)
);

COMMENT ON TABLE users IS 'Таблица пользователей Telegram бота';
COMMENT ON COLUMN users.chat_id IS 'Уникальный идентификатор чата пользователя в Telegram';

-- Таблица ветеринарных врачей
CREATE TABLE IF NOT EXISTS veterinarians (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    specialty VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    phone VARCHAR(50) NOT NULL,
    work_hours VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ограничения для валидации данных
    CONSTRAINT valid_phone CHECK (phone ~ '^\+?[0-9\s\-\(\)]+$')
);

COMMENT ON TABLE veterinarians IS 'Таблица ветеринарных врачей';
COMMENT ON COLUMN veterinarians.specialty IS 'Специализация врача (терапевт, хирург, стоматолог и т.д.)';

-- Создание индексов для оптимизации поиска
CREATE INDEX IF NOT EXISTS idx_veterinarians_specialty ON veterinarians (specialty);
CREATE INDEX IF NOT EXISTS idx_veterinarians_name ON veterinarians (name);
CREATE INDEX IF NOT EXISTS idx_users_chat_id ON users (chat_id);

-- Вставка начальных данных
-- Администратор по умолчанию (chat_id нужно заменить на реальный)
INSERT INTO users (username, chat_id, is_admin) 
VALUES ('admin', 0, true)
ON CONFLICT (chat_id) DO NOTHING;

-- Примеры ветеринарных врачей
INSERT INTO veterinarians (name, specialty, address, phone, work_hours) VALUES
('Доктор Айболит', 'терапевт', 'ул. Центральная, 1', '+7 (999) 123-45-67', '09:00-18:00'),
('Доктор Бобров', 'хирург', 'ул. Ветеринарная, 15', '+7 (999) 765-43-21', '10:00-19:00'),
('Доктор Котов', 'стоматолог', 'пр. Животных, 33', '+7 (999) 555-44-33', '08:00-17:00'),
('Доктор Птичкин', 'ортопед', 'ул. Костистая, 7', '+7 (999) 888-77-66', '09:00-18:00')
ON CONFLICT DO NOTHING;

-- Логирование успешного выполнения миграции
DO $$ 
BEGIN
    RAISE NOTICE 'Миграция 001 успешно выполнена: созданы таблицы users и veterinarians';
END $$;