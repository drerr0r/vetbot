package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/drerr0r/vetbot/internal/models"
	_ "github.com/lib/pq"
)

// Database обертка вокруг sql.DB с методами для работы с данными
type Database struct {
	db *sql.DB
}

// NewDatabase создает новый экземпляр Database
func NewDatabase(dataSourceName string) (*Database, error) {
	sqlDB, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Устанавливаем настройки пула соединений
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Проверяем соединение
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Database connection established")
	return &Database{db: sqlDB}, nil
}

// Close закрывает соединение с базой данных
func (d *Database) Close() error {
	if d.db != nil {
		if err := d.db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
			return err
		}
		log.Println("Database connection closed")
	}
	return nil
}

// GetDB возвращает внутренний *sql.DB для прямого доступа
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// User methods
func (d *Database) CreateUser(user *models.User) error {
	query := `INSERT INTO users (telegram_id, username, first_name, last_name, phone) 
	          VALUES ($1, $2, $3, $4, $5) 
			  ON CONFLICT (telegram_id) DO UPDATE 
			  SET username = EXCLUDED.username, first_name = EXCLUDED.first_name, 
			      last_name = EXCLUDED.last_name, phone = EXCLUDED.phone
			  RETURNING id, created_at`

	err := d.db.QueryRow(query, user.TelegramID, user.Username, user.FirstName, user.LastName, user.Phone).
		Scan(&user.ID, &user.CreatedAt)
	return err
}

func (d *Database) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	query := `SELECT id, telegram_id, username, first_name, last_name, phone, created_at 
	          FROM users WHERE telegram_id = $1`
	row := d.db.QueryRow(query, telegramID)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.TelegramID, &user.Username, &user.FirstName,
		&user.LastName, &user.Phone, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// Specialization methods
func (d *Database) GetAllSpecializations() ([]*models.Specialization, error) {
	log.Printf("Getting all specializations from database")

	query := `SELECT id, name, description, created_at FROM specializations ORDER BY name`
	rows, err := d.db.Query(query)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var specializations []*models.Specialization
	count := 0
	for rows.Next() {
		spec := &models.Specialization{}
		err := rows.Scan(&spec.ID, &spec.Name, &spec.Description, &spec.CreatedAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		specializations = append(specializations, spec)
		count++
	}

	log.Printf("Successfully retrieved %d specializations", count)
	return specializations, nil
}

func (d *Database) GetSpecializationByID(id int) (*models.Specialization, error) {
	query := `SELECT id, name, description, created_at FROM specializations WHERE id = $1`
	row := d.db.QueryRow(query, id)

	spec := &models.Specialization{}
	err := row.Scan(&spec.ID, &spec.Name, &spec.Description, &spec.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return spec, nil
}

// SpecializationExists проверяет существование специализации по ID
func (d *Database) SpecializationExists(specID int) (bool, error) {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM specializations WHERE id = $1)", specID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking if specialization exists (ID: %d): %v", specID, err)
		return false, err
	}
	return exists, nil
}

// Veterinarian methods
func (d *Database) GetVeterinariansBySpecialization(specializationID int) ([]*models.Veterinarian, error) {
	query := `
		SELECT v.id, v.first_name, v.last_name, v.phone, v.email, 
		       v.description, v.experience_years, v.is_active, v.created_at
		FROM veterinarians v
		JOIN vet_specializations vs ON v.id = vs.vet_id
		WHERE vs.specialization_id = $1 AND v.is_active = true
		ORDER BY v.experience_years DESC`

	rows, err := d.db.Query(query, specializationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vets []*models.Veterinarian
	for rows.Next() {
		vet := &models.Veterinarian{}
		err := rows.Scan(
			&vet.ID,
			&vet.FirstName,
			&vet.LastName,
			&vet.Phone,
			&vet.Email,
			&vet.Description,
			&vet.ExperienceYears,
			&vet.IsActive,
			&vet.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Загружаем специализации для врача
		specs, err := d.GetSpecializationsByVetID(vet.ID)
		if err == nil {
			vet.Specializations = specs
		}

		vets = append(vets, vet)
	}
	return vets, nil
}

// GetSpecializationsByVetID возвращает специализации врача
func (d *Database) GetSpecializationsByVetID(vetID int) ([]*models.Specialization, error) {
	query := `
		SELECT s.id, s.name, s.description, s.created_at
		FROM specializations s
		JOIN vet_specializations vs ON s.id = vs.specialization_id
		WHERE vs.vet_id = $1`

	rows, err := d.db.Query(query, vetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var specializations []*models.Specialization
	for rows.Next() {
		spec := &models.Specialization{}
		err := rows.Scan(&spec.ID, &spec.Name, &spec.Description, &spec.CreatedAt)
		if err != nil {
			return nil, err
		}
		specializations = append(specializations, spec)
	}
	return specializations, nil
}

func (d *Database) GetVeterinarianByID(id int) (*models.Veterinarian, error) {
	query := `
		SELECT id, first_name, last_name, phone, email, description, 
		       experience_years, is_active, created_at
		FROM veterinarians WHERE id = $1`

	row := d.db.QueryRow(query, id)
	vet := &models.Veterinarian{}
	err := row.Scan(
		&vet.ID,
		&vet.FirstName,
		&vet.LastName,
		&vet.Phone,
		&vet.Email,
		&vet.Description,
		&vet.ExperienceYears,
		&vet.IsActive,
		&vet.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return vet, nil
}

// Schedule methods
func (d *Database) GetSchedulesByVetID(vetID int) ([]*models.Schedule, error) {
	query := `
        SELECT s.id, s.vet_id, s.clinic_id, s.day_of_week, 
               TO_CHAR(s.start_time, 'HH24:MI') as start_time,
               TO_CHAR(s.end_time, 'HH24:MI') as end_time,
               s.is_available, s.created_at,
               c.name as clinic_name, c.address, c.phone, c.working_hours
        FROM schedules s
        JOIN clinics c ON s.clinic_id = c.id
        WHERE s.vet_id = $1 AND s.is_available = true
        ORDER BY s.day_of_week, s.start_time`

	rows, err := d.db.Query(query, vetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.Schedule
	for rows.Next() {
		schedule := &models.Schedule{
			Clinic: &models.Clinic{},
		}
		err := rows.Scan(
			&schedule.ID,
			&schedule.VetID,
			&schedule.ClinicID,
			&schedule.DayOfWeek,
			&schedule.StartTime,
			&schedule.EndTime,
			&schedule.IsAvailable,
			&schedule.CreatedAt,
			&schedule.Clinic.Name,
			&schedule.Clinic.Address,
			&schedule.Clinic.Phone,
			&schedule.Clinic.WorkingHours,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

func (d *Database) FindAvailableVets(criteria *models.SearchCriteria) ([]*models.Veterinarian, error) {
	query := `
		SELECT DISTINCT v.id, v.first_name, v.last_name, v.phone, v.email,
		       v.description, v.experience_years, v.created_at
		FROM veterinarians v
		JOIN vet_specializations vs ON v.id = vs.vet_id
		JOIN schedules s ON v.id = s.vet_id
		WHERE v.is_active = true AND s.is_available = true
		AND ($1 = 0 OR vs.specialization_id = $1)
		AND ($2 = 0 OR s.day_of_week = $2)
		AND ($3 = '' OR ($3::time BETWEEN s.start_time AND s.end_time))
		AND ($4 = 0 OR s.clinic_id = $4)
		ORDER BY v.experience_years DESC`

	rows, err := d.db.Query(query, criteria.SpecializationID, criteria.DayOfWeek,
		criteria.Time, criteria.ClinicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vets []*models.Veterinarian
	for rows.Next() {
		vet := &models.Veterinarian{}
		err := rows.Scan(
			&vet.ID,
			&vet.FirstName,
			&vet.LastName,
			&vet.Phone,
			&vet.Email,
			&vet.Description,
			&vet.ExperienceYears,
			&vet.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		vets = append(vets, vet)
	}
	return vets, nil
}

// Clinic methods
func (d *Database) GetAllClinics() ([]*models.Clinic, error) {
	query := `SELECT id, name, address, phone, working_hours, created_at FROM clinics ORDER BY name`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clinics []*models.Clinic
	for rows.Next() {
		clinic := &models.Clinic{}
		err := rows.Scan(
			&clinic.ID,
			&clinic.Name,
			&clinic.Address,
			&clinic.Phone,
			&clinic.WorkingHours,
			&clinic.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		clinics = append(clinics, clinic)
	}
	return clinics, nil
}

// UserRequest methods
func (d *Database) CreateUserRequest(request *models.UserRequest) error {
	query := `INSERT INTO user_requests (user_id, specialization_id, search_query) 
	          VALUES ($1, $2, $3) RETURNING id, created_at`
	err := d.db.QueryRow(query, request.UserID, request.SpecializationID, request.SearchQuery).
		Scan(&request.ID, &request.CreatedAt)
	return err
}

// AddUserIfNotExists добавляет пользователя если его еще нет (для совместимости со старым кодом)
func (d *Database) AddUserIfNotExists(telegramID int64, username, firstName, lastName string) error {
	query := `INSERT INTO users (telegram_id, username, first_name, last_name) 
	          VALUES ($1, $2, $3, $4) 
			  ON CONFLICT (telegram_id) DO UPDATE 
			  SET username = EXCLUDED.username, first_name = EXCLUDED.first_name, 
			      last_name = EXCLUDED.last_name
			  RETURNING id, created_at`

	var userID int
	var createdAt time.Time
	err := d.db.QueryRow(query, telegramID, username, firstName, lastName).
		Scan(&userID, &createdAt)
	return err
}
