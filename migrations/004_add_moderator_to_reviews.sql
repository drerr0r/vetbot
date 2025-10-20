-- migrations/004_add_moderation_fields.sql
ALTER TABLE reviews 
ADD COLUMN moderator_id INTEGER REFERENCES users(id),
ADD COLUMN moderated_at TIMESTAMP WITH TIME ZONE;

-- Создаем индекс для быстрого поиска отзывов по статусу
CREATE INDEX idx_reviews_status ON reviews(status);
CREATE INDEX idx_reviews_moderated_at ON reviews(moderated_at);