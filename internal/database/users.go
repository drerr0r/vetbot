package database

import (
	"database/sql"
	"time"

	"github.com/drerr0r/vetbot/internal/models"
)

// UserRepository содержит методы для работы с пользователями
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByID возвращает пользователя по ID
func (r *UserRepository) GetUserByID(userID int) (*models.User, error) {
	query := `SELECT id, telegram_id, username, first_name, last_name, phone, created_at 
              FROM users WHERE id = $1`

	var user models.User
	err := r.db.QueryRow(query, userID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByTelegramID возвращает пользователя по Telegram ID
func (r *UserRepository) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	query := `SELECT id, telegram_id, username, first_name, last_name, phone, created_at 
              FROM users WHERE telegram_id = $1`

	var user models.User
	err := r.db.QueryRow(query, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser создает нового пользователя
func (r *UserRepository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (telegram_id, username, first_name, last_name, phone, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := r.db.QueryRow(query,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Phone,
		time.Now(),
	).Scan(&user.ID)

	return err
}

// UpdateUser обновляет данные пользователя
func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `UPDATE users SET username = $1, first_name = $2, last_name = $3, phone = $4 
              WHERE id = $5`

	_, err := r.db.Exec(query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.ID,
	)

	return err
}
