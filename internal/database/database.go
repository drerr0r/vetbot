package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"vetbot/internal/models"
	"vetbot/pkg/utils"

	_ "github.com/lib/pq" // Драйвер PostgreSQL
)

// Database представляет обертку для работы с базой данных
type Database struct {
	DB *sql.DB
}

// InitDB инициализирует подключение к базе данных PostgreSQL
func InitDB(config *utils.Config) (*Database, error) {
	// Получаем строку подключения из конфигурации
	connStr := config.GetDBConnectionString()

	// Открываем соединение с базой данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ошибка ping базы данных: %v", err)
	}

	log.Println("✅ Успешное подключение к PostgreSQL")

	return &Database{DB: db}, nil
}

// Close закрывает соединение с базой данных
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// RunMigrations выполняет миграции базы данных
func (d *Database) RunMigrations() error {
	// Читаем файл миграций
	migrationSQL, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("ошибка чтения файла миграций: %v", err)
	}

	// Выполняем SQL миграций
	_, err = d.DB.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("ошибка выполнения миграций: %v", err)
	}

	log.Println("✅ Миграции базы данных успешно выполнены")
	return nil
}

// CreateVeterinarian создает нового ветеринарного врача в базе данных
func (d *Database) CreateVeterinarian(vet *models.Veterinarian) error {
	query := `
		INSERT INTO veterinarians (name, specialty, address, phone, work_hours, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err := d.DB.QueryRow(
		query,
		vet.Name,
		vet.Specialty,
		vet.Address,
		vet.Phone,
		vet.WorkHours,
		vet.CreatedAt,
		vet.UpdatedAt,
	).Scan(&vet.ID)

	if err != nil {
		return fmt.Errorf("ошибка создания врача: %v", err)
	}

	return nil
}

// GetVeterinarianByID возвращает врача по ID
func (d *Database) GetVeterinarianByID(id int64) (*models.Veterinarian, error) {
	query := `
		SELECT id, name, specialty, address, phone, work_hours, created_at, updated_at
		FROM veterinarians WHERE id = $1
	`

	row := d.DB.QueryRow(query, id)
	vet := &models.Veterinarian{}

	err := row.Scan(
		&vet.ID,
		&vet.Name,
		&vet.Specialty,
		&vet.Address,
		&vet.Phone,
		&vet.WorkHours,
		&vet.CreatedAt,
		&vet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("врач с ID %d не найден", id)
		}
		return nil, fmt.Errorf("ошибка получения врача: %v", err)
	}

	return vet, nil
}

// GetAllVeterinarians возвращает всех ветеринарных врачей
func (d *Database) GetAllVeterinarians() ([]models.Veterinarian, error) {
	query := `
		SELECT id, name, specialty, address, phone, work_hours, created_at, updated_at
		FROM veterinarians ORDER BY name
	`

	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса врачей: %v", err)
	}
	defer rows.Close()

	var veterinarians []models.Veterinarian

	for rows.Next() {
		var vet models.Veterinarian
		err := rows.Scan(
			&vet.ID,
			&vet.Name,
			&vet.Specialty,
			&vet.Address,
			&vet.Phone,
			&vet.WorkHours,
			&vet.CreatedAt,
			&vet.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования врача: %v", err)
		}
		veterinarians = append(veterinarians, vet)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по врачам: %v", err)
	}

	return veterinarians, nil
}

// FindVeterinariansBySpecialty ищет врачей по специализации
func (d *Database) FindVeterinariansBySpecialty(specialty string) ([]models.Veterinarian, error) {
	query := `
		SELECT id, name, specialty, address, phone, work_hours, created_at, updated_at
		FROM veterinarians WHERE LOWER(specialty) LIKE LOWER($1) ORDER BY name
	`

	rows, err := d.DB.Query(query, "%"+specialty+"%")
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска врачей: %v", err)
	}
	defer rows.Close()

	var veterinarians []models.Veterinarian

	for rows.Next() {
		var vet models.Veterinarian
		err := rows.Scan(
			&vet.ID,
			&vet.Name,
			&vet.Specialty,
			&vet.Address,
			&vet.Phone,
			&vet.WorkHours,
			&vet.CreatedAt,
			&vet.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования врача: %v", err)
		}
		veterinarians = append(veterinarians, vet)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по врачам: %v", err)
	}

	return veterinarians, nil
}

// UpdateVeterinarian обновляет данные врача
func (d *Database) UpdateVeterinarian(id int64, updateData map[string]interface{}) error {
	// Начинаем построение SQL запроса
	query := "UPDATE veterinarians SET "
	params := []interface{}{}
	paramCount := 1

	// Добавляем обновляемые поля
	for field, value := range updateData {
		query += fmt.Sprintf("%s = $%d, ", field, paramCount)
		params = append(params, value)
		paramCount++
	}

	// Добавляем обновление времени и условие WHERE
	query += "updated_at = NOW() WHERE id = $" + fmt.Sprintf("%d", paramCount)
	params = append(params, id)

	// Выполняем запрос
	result, err := d.DB.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("ошибка обновления врача: %v", err)
	}

	// Проверяем, что запись была обновлена
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки обновления: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("врач с ID %d не найден", id)
	}

	return nil
}

// DeleteVeterinarian удаляет врача по ID
func (d *Database) DeleteVeterinarian(id int64) error {
	query := "DELETE FROM veterinarians WHERE id = $1"

	result, err := d.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления врача: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки удаления: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("врач с ID %d не найден", id)
	}

	return nil
}

// UserExists проверяет существование пользователя
func (d *Database) UserExists(chatID int64) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE chat_id = $1"

	var count int
	err := d.DB.QueryRow(query, chatID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки пользователя: %v", err)
	}

	return count > 0, nil
}

// CreateUser создает нового пользователя
func (d *Database) CreateUser(username string, chatID int64, isAdmin bool) error {
	query := "INSERT INTO users (username, chat_id, is_admin) VALUES ($1, $2, $3)"

	_, err := d.DB.Exec(query, username, chatID, isAdmin)
	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %v", err)
	}

	return nil
}

// IsAdmin проверяет, является ли пользователь администратором
func (d *Database) IsAdmin(chatID int64) (bool, error) {
	query := "SELECT is_admin FROM users WHERE chat_id = $1"

	var isAdmin bool
	err := d.DB.QueryRow(query, chatID).Scan(&isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("ошибка проверки администратора: %v", err)
	}

	return isAdmin, nil
}
