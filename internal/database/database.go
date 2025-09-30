package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/drerr0r/vetbot/internal/models"
	_ "github.com/lib/pq"
)

// Database представляет обертку для работы с базой данных
type Database struct {
	db *sql.DB
}

// New создает новое подключение к базе данных
func New(dataSourceName string) (*Database, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to database")
	return &Database{db: db}, nil
}

// Close закрывает подключение к базе данных
func (d *Database) Close() error {
	if d == nil || d.db == nil {
		return nil // Безопасно возвращаем nil для nil объекта
	}
	return d.db.Close()
}

// GetDB возвращает объект базы данных
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// CreateUser создает нового пользователя
func (d *Database) CreateUser(user *models.User) error {
	query := `INSERT INTO users (telegram_id, username, first_name, last_name, phone) 
	          VALUES ($1, $2, $3, $4, $5) 
	          ON CONFLICT (telegram_id) DO UPDATE SET 
	          username = EXCLUDED.username, 
	          first_name = EXCLUDED.first_name, 
	          last_name = EXCLUDED.last_name, 
	          phone = EXCLUDED.phone
	          RETURNING id, created_at`

	err := d.db.QueryRow(query, user.TelegramID, user.Username, user.FirstName, user.LastName, user.Phone).
		Scan(&user.ID, &user.CreatedAt)
	return err
}

// GetAllSpecializations возвращает все специализации
func (d *Database) GetAllSpecializations() ([]*models.Specialization, error) {
	rows, err := d.db.Query("SELECT id, name, description, created_at FROM specializations ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var specializations []*models.Specialization
	for rows.Next() {
		var spec models.Specialization
		err := rows.Scan(&spec.ID, &spec.Name, &spec.Description, &spec.CreatedAt)
		if err != nil {
			return nil, err
		}
		specializations = append(specializations, &spec)
	}

	return specializations, nil
}

// GetSpecializationByID возвращает специализацию по ID
func (d *Database) GetSpecializationByID(id int) (*models.Specialization, error) {
	var spec models.Specialization
	err := d.db.QueryRow("SELECT id, name, description, created_at FROM specializations WHERE id = $1", id).
		Scan(&spec.ID, &spec.Name, &spec.Description, &spec.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

// SpecializationExists проверяет существование специализации
func (d *Database) SpecializationExists(id int) (bool, error) {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM specializations WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

// GetVeterinariansBySpecialization возвращает врачей по специализации
func (d *Database) GetVeterinariansBySpecialization(specializationID int) ([]*models.Veterinarian, error) {
	query := `
		SELECT DISTINCT v.id, v.first_name, v.last_name, v.phone, v.email, 
		       v.description, v.experience_years, v.is_active, v.created_at
		FROM veterinarians v
		INNER JOIN vet_specializations vs ON v.id = vs.vet_id
		WHERE vs.specialization_id = $1 AND v.is_active = true
		ORDER BY v.first_name, v.last_name`

	rows, err := d.db.Query(query, specializationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var veterinarians []*models.Veterinarian
	for rows.Next() {
		var vet models.Veterinarian
		err := rows.Scan(&vet.ID, &vet.FirstName, &vet.LastName, &vet.Phone, &vet.Email,
			&vet.Description, &vet.ExperienceYears, &vet.IsActive, &vet.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Загружаем специализации для каждого врача
		specs, err := d.GetSpecializationsByVetID(vet.ID)
		if err == nil {
			vet.Specializations = specs
		}

		veterinarians = append(veterinarians, &vet)
	}

	return veterinarians, nil
}

// GetSpecializationsByVetID возвращает специализации врача
func (d *Database) GetSpecializationsByVetID(vetID int) ([]*models.Specialization, error) {
	query := `
		SELECT s.id, s.name, s.description, s.created_at
		FROM specializations s
		INNER JOIN vet_specializations vs ON s.id = vs.specialization_id
		WHERE vs.vet_id = $1
		ORDER BY s.name`

	rows, err := d.db.Query(query, vetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var specializations []*models.Specialization
	for rows.Next() {
		var spec models.Specialization
		err := rows.Scan(&spec.ID, &spec.Name, &spec.Description, &spec.CreatedAt)
		if err != nil {
			return nil, err
		}
		specializations = append(specializations, &spec)
	}

	return specializations, nil
}

// GetAllClinics возвращает все клиники
func (d *Database) GetAllClinics() ([]*models.Clinic, error) {
	query := "SELECT id, name, address, phone, working_hours, is_active, created_at FROM clinics ORDER BY name"
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clinics []*models.Clinic
	for rows.Next() {
		var clinic models.Clinic
		err := rows.Scan(&clinic.ID, &clinic.Name, &clinic.Address, &clinic.Phone,
			&clinic.WorkingHours, &clinic.IsActive, &clinic.CreatedAt)
		if err != nil {
			return nil, err
		}
		clinics = append(clinics, &clinic)
	}

	return clinics, nil
}

// GetSchedulesByVetID возвращает расписание врача
func (d *Database) GetSchedulesByVetID(vetID int) ([]*models.Schedule, error) {
	query := `
		SELECT s.id, s.vet_id, s.clinic_id, s.day_of_week, s.start_time, s.end_time, 
		       s.is_available, s.created_at,
		       c.name, c.address, c.phone, c.working_hours, c.is_active, c.created_at
		FROM schedules s
		LEFT JOIN clinics c ON s.clinic_id = c.id
		WHERE s.vet_id = $1 AND s.is_available = true
		ORDER BY s.day_of_week, s.start_time`

	rows, err := d.db.Query(query, vetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		var clinic models.Clinic
		var clinicPhone, clinicWorkingHours sql.NullString

		err := rows.Scan(&schedule.ID, &schedule.VetID, &schedule.ClinicID, &schedule.DayOfWeek,
			&schedule.StartTime, &schedule.EndTime, &schedule.IsAvailable, &schedule.CreatedAt,
			&clinic.Name, &clinic.Address, &clinicPhone, &clinicWorkingHours, &clinic.IsActive, &clinic.CreatedAt)

		if err != nil {
			return nil, err
		}

		clinic.Phone = clinicPhone
		clinic.WorkingHours = clinicWorkingHours
		schedule.Clinic = &clinic
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// FindAvailableVets ищет доступных врачей по критериям
func (d *Database) FindAvailableVets(criteria *models.SearchCriteria) ([]*models.Veterinarian, error) {
	query := `
		SELECT DISTINCT v.id, v.first_name, v.last_name, v.phone, v.email, 
		       v.description, v.experience_years, v.is_active, v.created_at
		FROM veterinarians v
		LEFT JOIN vet_specializations vs ON v.id = vs.vet_id
		LEFT JOIN schedules s ON v.id = s.vet_id
		WHERE v.is_active = true AND s.is_available = true`

	args := []interface{}{}
	argCount := 0

	if criteria.SpecializationID > 0 {
		argCount++
		query += fmt.Sprintf(" AND vs.specialization_id = $%d", argCount)
		args = append(args, criteria.SpecializationID)
	}

	if criteria.DayOfWeek > 0 {
		argCount++
		query += fmt.Sprintf(" AND s.day_of_week = $%d", argCount)
		args = append(args, criteria.DayOfWeek)
	}

	if criteria.ClinicID > 0 {
		argCount++
		query += fmt.Sprintf(" AND s.clinic_id = $%d", argCount)
		args = append(args, criteria.ClinicID)
	}

	query += " ORDER BY v.first_name, v.last_name"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var veterinarians []*models.Veterinarian
	for rows.Next() {
		var vet models.Veterinarian
		err := rows.Scan(&vet.ID, &vet.FirstName, &vet.LastName, &vet.Phone, &vet.Email,
			&vet.Description, &vet.ExperienceYears, &vet.IsActive, &vet.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Загружаем специализации для каждого врача
		specs, err := d.GetSpecializationsByVetID(vet.ID)
		if err == nil {
			vet.Specializations = specs
		}

		veterinarians = append(veterinarians, &vet)
	}

	return veterinarians, nil
}

// ========== НОВЫЕ МЕТОДЫ ДЛЯ АДМИНКИ ==========

// GetAllVeterinarians возвращает всех врачей
func (d *Database) GetAllVeterinarians() ([]*models.Veterinarian, error) {
	query := `SELECT id, first_name, last_name, phone, email, description, experience_years, is_active, created_at 
              FROM veterinarians ORDER BY first_name, last_name`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vets []*models.Veterinarian
	for rows.Next() {
		var vet models.Veterinarian
		err := rows.Scan(&vet.ID, &vet.FirstName, &vet.LastName, &vet.Phone, &vet.Email,
			&vet.Description, &vet.ExperienceYears, &vet.IsActive, &vet.CreatedAt)
		if err != nil {
			return nil, err
		}
		vets = append(vets, &vet)
	}

	return vets, nil
}

// GetVeterinarianByID возвращает врача по ID
func (d *Database) GetVeterinarianByID(id int) (*models.Veterinarian, error) {
	query := `SELECT id, first_name, last_name, phone, email, description, experience_years, is_active, created_at 
              FROM veterinarians WHERE id = $1`

	var vet models.Veterinarian
	err := d.db.QueryRow(query, id).Scan(&vet.ID, &vet.FirstName, &vet.LastName, &vet.Phone,
		&vet.Email, &vet.Description, &vet.ExperienceYears,
		&vet.IsActive, &vet.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &vet, nil
}

// GetClinicByID возвращает клинику по ID
func (d *Database) GetClinicByID(id int) (*models.Clinic, error) {
	query := `SELECT id, name, address, phone, working_hours, is_active, created_at 
              FROM clinics WHERE id = $1`

	var clinic models.Clinic
	err := d.db.QueryRow(query, id).Scan(&clinic.ID, &clinic.Name, &clinic.Address,
		&clinic.Phone, &clinic.WorkingHours, &clinic.IsActive,
		&clinic.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &clinic, nil
}

func (d *Database) AddMissingColumns() error {
	// Добавляем is_active в clinics если не существует
	_, err := d.db.Exec(`
		DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
						  WHERE table_name = 'clinics' AND column_name = 'is_active') THEN
				ALTER TABLE clinics ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
			END IF;
		END $$;
	`)
	if err != nil {
		return err
	}

	// Добавляем is_active в veterinarians если не существует
	_, err = d.db.Exec(`
		DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
						  WHERE table_name = 'veterinarians' AND column_name = 'is_active') THEN
				ALTER TABLE veterinarians ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
			END IF;
		END $$;
	`)
	return err
}

// ========== МЕТОДЫ ДЛЯ ГОРОДОВ И ИМПОРТА ==========

// CreateCity создает новый город
func (d *Database) CreateCity(city *models.City) error {
	query := `INSERT INTO cities (name, region) VALUES ($1, $2) RETURNING id, created_at`
	return d.db.QueryRow(query, city.Name, city.Region).Scan(&city.ID, &city.CreatedAt)
}

// GetAllCities возвращает все города
func (d *Database) GetAllCities() ([]*models.City, error) {
	query := `SELECT id, name, region, created_at FROM cities ORDER BY name`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []*models.City
	for rows.Next() {
		var city models.City
		err := rows.Scan(&city.ID, &city.Name, &city.Region, &city.CreatedAt)
		if err != nil {
			return nil, err
		}
		cities = append(cities, &city)
	}
	return cities, nil
}

// GetCityByID возвращает город по ID
func (d *Database) GetCityByID(id int) (*models.City, error) {
	query := `SELECT id, name, region, created_at FROM cities WHERE id = $1`
	var city models.City
	err := d.db.QueryRow(query, id).Scan(&city.ID, &city.Name, &city.Region, &city.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &city, nil
}

// GetCityByName возвращает город по названию
func (d *Database) GetCityByName(name string) (*models.City, error) {
	query := `SELECT id, name, region, created_at FROM cities WHERE LOWER(name) = LOWER($1)`
	var city models.City
	err := d.db.QueryRow(query, name).Scan(&city.ID, &city.Name, &city.Region, &city.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &city, nil
}

// CreateClinicWithCity создает клинику с привязкой к городу
func (d *Database) CreateClinicWithCity(clinic *models.Clinic) error {
	query := `INSERT INTO clinics (name, address, phone, working_hours, is_active, city_id, district, metro_station) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`

	return d.db.QueryRow(query,
		clinic.Name,
		clinic.Address,
		clinic.Phone,
		clinic.WorkingHours,
		clinic.IsActive,
		clinic.CityID,
		clinic.District,
		clinic.MetroStation,
	).Scan(&clinic.ID, &clinic.CreatedAt)
}

// GetAllClinicsWithCities возвращает все клиники с информацией о городах
func (d *Database) GetAllClinicsWithCities() ([]*models.Clinic, error) {
	query := `
		SELECT c.id, c.name, c.address, c.phone, c.working_hours, c.is_active, 
		       c.city_id, c.district, c.metro_station, c.created_at,
		       ct.id, ct.name, ct.region, ct.created_at
		FROM clinics c
		LEFT JOIN cities ct ON c.city_id = ct.id
		ORDER BY c.name`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clinics []*models.Clinic
	for rows.Next() {
		var clinic models.Clinic
		var cityID sql.NullInt64
		var city models.City
		var cityCreatedAt time.Time

		err := rows.Scan(
			&clinic.ID, &clinic.Name, &clinic.Address, &clinic.Phone, &clinic.WorkingHours,
			&clinic.IsActive, &cityID, &clinic.District, &clinic.MetroStation, &clinic.CreatedAt,
			&city.ID, &city.Name, &city.Region, &cityCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		clinic.CityID = cityID
		if cityID.Valid {
			city.CreatedAt = cityCreatedAt
			clinic.City = &city
		}

		clinics = append(clinics, &clinic)
	}
	return clinics, nil
}

// UpdateClinic обновляет данные клиники
func (d *Database) UpdateClinic(clinic *models.Clinic) error {
	query := `UPDATE clinics SET 
		name = $1, address = $2, phone = $3, working_hours = $4, 
		is_active = $5, city_id = $6, district = $7, metro_station = $8
		WHERE id = $9`

	_, err := d.db.Exec(query,
		clinic.Name, clinic.Address, clinic.Phone, clinic.WorkingHours,
		clinic.IsActive, clinic.CityID, clinic.District, clinic.MetroStation, clinic.ID,
	)
	return err
}
