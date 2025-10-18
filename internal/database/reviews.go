package database

import (
	"database/sql"
	"time"

	"github.com/drerr0r/vetbot/internal/models"
)

// ReviewRepository содержит методы для работы с отзывами
type ReviewRepository struct {
	db *sql.DB
}

// NewReviewRepository создает новый репозиторий отзывов
func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// CreateReview создает новый отзыв
func (r *ReviewRepository) CreateReview(review *models.Review) error {
	query := `INSERT INTO reviews (veterinarian_id, user_id, rating, comment, status, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := r.db.QueryRow(query,
		review.VeterinarianID,
		review.UserID,
		review.Rating,
		review.Comment,
		review.Status,
		review.CreatedAt,
	).Scan(&review.ID)

	return err
}

// GetReviewByID возвращает отзыв по ID с полной информацией
func (r *ReviewRepository) GetReviewByID(reviewID int) (*models.Review, error) {
	query := `
		SELECT r.id, r.veterinarian_id, r.user_id, r.rating, r.comment, 
		       r.status, r.created_at, r.moderated_at, r.moderated_by,
		       v.id, v.first_name, v.last_name, v.phone,
		       u.id, u.telegram_id, u.username, u.first_name, u.last_name,
		       m.id, m.telegram_id, m.username, m.first_name, m.last_name
		FROM reviews r
		LEFT JOIN veterinarians v ON r.veterinarian_id = v.id
		LEFT JOIN users u ON r.user_id = u.id
		LEFT JOIN users m ON r.moderated_by = m.id
		WHERE r.id = $1`

	var review models.Review
	var vetID sql.NullInt64
	var userID, moderatorID sql.NullInt64
	var moderatedAt sql.NullTime

	err := r.db.QueryRow(query, reviewID).Scan(
		&review.ID, &review.VeterinarianID, &userID, &review.Rating,
		&review.Comment, &review.Status, &review.CreatedAt, &moderatedAt, &moderatorID,
		&vetID, &review.Veterinarian.FirstName, &review.Veterinarian.LastName, &review.Veterinarian.Phone,
		&review.User.ID, &review.User.TelegramID, &review.User.Username, &review.User.FirstName, &review.User.LastName,
		&review.Moderator.ID, &review.Moderator.TelegramID, &review.Moderator.Username, &review.Moderator.FirstName, &review.Moderator.LastName,
	)

	if err != nil {
		return nil, err
	}

	if moderatedAt.Valid {
		review.ModeratedAt = moderatedAt
	}
	if moderatorID.Valid {
		review.ModeratedBy = sql.NullInt64{Int64: moderatorID.Int64, Valid: true}
	}

	return &review, nil
}

// GetApprovedReviewsByVet возвращает одобренные отзывы по врачу
func (r *ReviewRepository) GetApprovedReviewsByVet(vetID int) ([]*models.Review, error) {
	query := `
		SELECT r.id, r.rating, r.comment, r.created_at,
		       u.first_name, u.last_name
		FROM reviews r
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.veterinarian_id = $1 AND r.status = 'approved'
		ORDER BY r.created_at DESC
		LIMIT 50`

	rows, err := r.db.Query(query, vetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*models.Review
	for rows.Next() {
		var review models.Review
		var user models.User

		err := rows.Scan(
			&review.ID, &review.Rating, &review.Comment, &review.CreatedAt,
			&user.FirstName, &user.LastName,
		)
		if err != nil {
			return nil, err
		}

		review.User = &user
		review.VeterinarianID = vetID
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

// GetPendingReviews возвращает отзывы ожидающие модерации
func (r *ReviewRepository) GetPendingReviews() ([]*models.Review, error) {
	query := `
		SELECT r.id, r.veterinarian_id, r.user_id, r.rating, r.comment, r.created_at,
		       v.first_name, v.last_name, v.phone,
		       u.first_name, u.last_name
		FROM reviews r
		LEFT JOIN veterinarians v ON r.veterinarian_id = v.id
		LEFT JOIN users u ON r.user_id = u.id
		WHERE r.status = 'pending'
		ORDER BY r.created_at ASC
		LIMIT 100`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*models.Review
	for rows.Next() {
		var review models.Review
		var vet models.Veterinarian
		var user models.User

		err := rows.Scan(
			&review.ID, &review.VeterinarianID, &review.UserID, &review.Rating,
			&review.Comment, &review.CreatedAt,
			&vet.FirstName, &vet.LastName, &vet.Phone,
			&user.FirstName, &user.LastName,
		)
		if err != nil {
			return nil, err
		}

		review.Veterinarian = &vet
		review.User = &user
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

// UpdateReviewStatus обновляет статус отзыва
func (r *ReviewRepository) UpdateReviewStatus(reviewID int, status string, moderatorID int) error {
	var query string
	var err error

	if moderatorID > 0 {
		query = `UPDATE reviews SET status = $1, moderated_at = $2, moderated_by = $3 WHERE id = $4`
		_, err = r.db.Exec(query, status, time.Now(), moderatorID, reviewID)
	} else {
		query = `UPDATE reviews SET status = $1, moderated_at = $2 WHERE id = $3`
		_, err = r.db.Exec(query, status, time.Now(), reviewID)
	}

	return err
}

// HasUserReviewForVet проверяет, есть ли у пользователя отзыв на врача
func (r *ReviewRepository) HasUserReviewForVet(userID int, vetID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM reviews WHERE user_id = $1 AND veterinarian_id = $2)`

	var exists bool
	err := r.db.QueryRow(query, userID, vetID).Scan(&exists)
	return exists, err
}

// GetReviewStats возвращает статистику отзывов по врачу
func (r *ReviewRepository) GetReviewStats(vetID int) (*models.ReviewStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_reviews,
			COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_reviews,
			COALESCE(AVG(CASE WHEN status = 'approved' THEN rating END), 0) as avg_rating
		FROM reviews 
		WHERE veterinarian_id = $1`

	var stats models.ReviewStats
	var total, approved int
	var avgRating sql.NullFloat64

	err := r.db.QueryRow(query, vetID).Scan(&total, &approved, &avgRating)
	if err != nil {
		return nil, err
	}

	stats.VeterinarianID = vetID
	stats.TotalReviews = total
	stats.ApprovedReviews = approved
	stats.AverageRating = 0.0
	if avgRating.Valid {
		stats.AverageRating = avgRating.Float64
	}

	return &stats, nil
}

// GetUserByTelegramID возвращает пользователя по Telegram ID
func (r *ReviewRepository) GetUserByTelegramID(telegramID int64) (*models.User, error) {
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

// GetUserByID возвращает пользователя по ID
func (r *ReviewRepository) GetUserByID(userID int) (*models.User, error) {
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
