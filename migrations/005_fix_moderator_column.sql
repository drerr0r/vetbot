-- Исправляем имя колонки в таблице reviews
ALTER TABLE reviews RENAME COLUMN moderated_by TO moderator_id;
