-- Миграция для добавления системы отзывов

-- Таблица отзывов
CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    veterinarian_id INTEGER NOT NULL REFERENCES veterinarians(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    moderated_at TIMESTAMP,
    moderated_by INTEGER REFERENCES users(id) ON DELETE SET NULL
);

-- Индексы для улучшения производительности
CREATE INDEX IF NOT EXISTS idx_reviews_veterinarian_id ON reviews(veterinarian_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews(status);
CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews(created_at);
CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews(rating);

-- Уникальный индекс чтобы пользователь мог оставить только один отзыв на врача
CREATE UNIQUE INDEX IF NOT EXISTS idx_reviews_user_vet_unique ON reviews(user_id, veterinarian_id);

-- Ограничение на длину комментария (с обработкой если уже существует)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_comment_length') THEN
        ALTER TABLE reviews ADD CONSTRAINT check_comment_length CHECK (length(comment) <= 500);
    END IF;
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Вставляем тестовые отзывы ТОЛЬКО если есть тестовые пользователи
DO $$ 
DECLARE
    user1_exists BOOLEAN;
    user2_exists BOOLEAN;
    user3_exists BOOLEAN;
BEGIN
    -- Проверяем существование пользователей
    SELECT EXISTS(SELECT 1 FROM users WHERE id = 1) INTO user1_exists;
    SELECT EXISTS(SELECT 1 FROM users WHERE id = 2) INTO user2_exists;
    SELECT EXISTS(SELECT 1 FROM users WHERE id = 3) INTO user3_exists;
    
    -- Вставляем отзывы только для существующих пользователей
    IF user1_exists THEN
        INSERT INTO reviews (veterinarian_id, user_id, rating, comment, status, created_at) VALUES 
        (1, 1, 5, 'Отличный врач! Очень помог нашему котику. Профессионал своего дела!', 'approved', NOW() - INTERVAL '5 days')
        ON CONFLICT (user_id, veterinarian_id) DO NOTHING;
    END IF;
    
    IF user2_exists THEN
        INSERT INTO reviews (veterinarian_id, user_id, rating, comment, status, created_at) VALUES 
        (1, 2, 4, 'Хороший специалист, но пришлось немного подождать. В целом доволен.', 'approved', NOW() - INTERVAL '3 days'),
        (3, 2, 3, 'Нормальный врач, но цены немного завышены. Лечение помогло.', 'approved', NOW() - INTERVAL '1 day')
        ON CONFLICT (user_id, veterinarian_id) DO NOTHING;
    END IF;
    
    IF user1_exists THEN
        INSERT INTO reviews (veterinarian_id, user_id, rating, comment, status, created_at) VALUES 
        (2, 1, 5, 'Очень внимательная и заботливая врач. Наш питомец чувствует себя прекрасно!', 'approved', NOW() - INTERVAL '2 days')
        ON CONFLICT (user_id, veterinarian_id) DO NOTHING;
    END IF;
    
    IF user3_exists THEN
        INSERT INTO reviews (veterinarian_id, user_id, rating, comment, status, created_at) VALUES 
        (1, 3, 5, 'Лучший ветеринар в городе! Спасибо за помощь!', 'pending', NOW())
        ON CONFLICT (user_id, veterinarian_id) DO NOTHING;
    END IF;
END $$;