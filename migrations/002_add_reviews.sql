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

-- Ограничение на длину комментария
ALTER TABLE reviews ADD CONSTRAINT check_comment_length CHECK (length(comment) <= 500);

-- Вставляем тестовые отзывы (опционально)
INSERT INTO reviews (veterinarian_id, user_id, rating, comment, status, created_at) VALUES 
(1, 1, 5, 'Отличный врач! Очень помог нашему котику. Профессионал своего дела!', 'approved', NOW() - INTERVAL '5 days'),
(1, 2, 4, 'Хороший специалист, но пришлось немного подождать. В целом доволен.', 'approved', NOW() - INTERVAL '3 days'),
(2, 1, 5, 'Очень внимательная и заботливая врач. Наш питомец чувствует себя прекрасно!', 'approved', NOW() - INTERVAL '2 days'),
(3, 2, 3, 'Нормальный врач, но цены немного завышены. Лечение помогло.', 'approved', NOW() - INTERVAL '1 day'),
(1, 3, 5, 'Лучший ветеринар в городе! Спасибо за помощь!', 'pending', NOW())
ON CONFLICT (user_id, veterinarian_id) DO NOTHING;