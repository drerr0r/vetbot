package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/drerr0r/vetbot/internal/models"
	_ "github.com/lib/pq"
)

// Database представляет обертку для работы с базой данных
type Database struct {
	db      *sql.DB
	users   *UserRepository
	reviews *ReviewRepository
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
	return &Database{
		db:      db,
		users:   NewUserRepository(db),
		reviews: NewReviewRepository(db),
	}, nil
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
		var vetID sql.NullInt64
		var email, description sql.NullString
		var experienceYears sql.NullInt64

		err := rows.Scan(&vetID, &vet.FirstName, &vet.LastName, &vet.Phone, &email,
			&description, &experienceYears, &vet.IsActive, &vet.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Заполняем nullable поля
		vet.ID = vetID
		vet.Email = email
		vet.Description = description
		vet.ExperienceYears = experienceYears

		// Загружаем специализации для каждого врача только если ID валиден
		if vetID.Valid {
			specs, err := d.GetSpecializationsByVetID(int(vetID.Int64))
			if err == nil {
				vet.Specializations = specs
			}
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
	query := "SELECT id, name, address, phone, working_hours, is_active, city_id, district, metro_station, created_at FROM clinics ORDER BY name"
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clinics []*models.Clinic
	for rows.Next() {
		var clinic models.Clinic
		var cityID sql.NullInt64
		var phone, workingHours sql.NullString

		err := rows.Scan(&clinic.ID, &clinic.Name, &clinic.Address, &phone,
			&workingHours, &clinic.IsActive, &cityID,
			&clinic.District, &clinic.MetroStation, &clinic.CreatedAt)
		if err != nil {
			return nil, err
		}

		clinic.Phone = phone
		clinic.WorkingHours = workingHours
		clinic.CityID = cityID
		clinics = append(clinics, &clinic)
	}

	return clinics, nil
}

// GetSchedulesByVetID возвращает расписание врача
func (d *Database) GetSchedulesByVetID(vetID int) ([]*models.Schedule, error) {
	query := `
        SELECT s.id, s.vet_id, s.clinic_id, s.day_of_week, 
               TO_CHAR(s.start_time, 'HH24:MI') as start_time,
               TO_CHAR(s.end_time, 'HH24:MI') as end_time,
               s.is_available, s.created_at,
               c.id, c.name, c.address, c.phone, c.working_hours, 
               c.is_active, c.city_id, c.district, c.metro_station, c.created_at
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
		var startTimeStr, endTimeStr string

		err := rows.Scan(
			&schedule.ID, &schedule.VetID, &schedule.ClinicID, &schedule.DayOfWeek,
			&startTimeStr, &endTimeStr,
			&schedule.IsAvailable, &schedule.CreatedAt,
			&clinic.ID, &clinic.Name, &clinic.Address, &clinic.Phone,
			&clinic.WorkingHours, &clinic.IsActive, &clinic.CityID,
			&clinic.District, &clinic.MetroStation, &clinic.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Используем строковое представление времени
		schedule.StartTime = startTimeStr
		schedule.EndTime = endTimeStr
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
		specs, err := d.GetSpecializationsByVetID(models.GetVetIDAsIntOrZero(&vet))
		if err == nil {
			vet.Specializations = specs
		}

		veterinarians = append(veterinarians, &vet)
	}

	return veterinarians, nil
}

// ========== НОВЫЕ МЕТОДЫ ДЛЯ АДМИНКИ ==========

func (d *Database) GetAllVeterinarians() ([]*models.Veterinarian, error) {
	query := `SELECT v.id, v.first_name, v.last_name, v.patronymic, v.phone, v.email, v.description, 
                     v.experience_years, v.is_active, v.city_id, v.created_at,
                     c.id, c.name, c.region, c.created_at
              FROM veterinarians v
              LEFT JOIN cities c ON v.city_id = c.id
              ORDER BY v.first_name, v.last_name`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса veterinarians: %v", err)
	}
	defer rows.Close()

	var vets []*models.Veterinarian
	for rows.Next() {
		var vet models.Veterinarian
		var cityID sql.NullInt64
		var city models.City
		var cityCreatedAt sql.NullTime
		var email, description, patronymic sql.NullString
		var experienceYears sql.NullInt64
		var vetID sql.NullInt64
		var cityID2 sql.NullInt64
		var cityName, cityRegion sql.NullString

		err := rows.Scan(
			&vetID, &vet.FirstName, &vet.LastName, &patronymic, &vet.Phone, &email,
			&description, &experienceYears, &vet.IsActive, &cityID, &vet.CreatedAt,
			&cityID2, &cityName, &cityRegion, &cityCreatedAt,
		)
		if err != nil {
			log.Printf("Ошибка сканирования строки veterinarians: %v", err)

			vet = models.Veterinarian{
				ID:        vetID,
				FirstName: "ОШИБКА_ДАННЫХ",
				LastName:  "Требует_редактирования",
				Phone:     "Не указан",
				IsActive:  false,
			}
		} else {
			vet.ID = vetID
			vet.Email = email
			vet.Description = description
			vet.ExperienceYears = experienceYears
			vet.CityID = cityID
			vet.Patronymic = patronymic

			// ИСПРАВЛЕНО: правильно заполняем город
			if cityID2.Valid {
				city.ID = int(cityID2.Int64)
				if cityName.Valid {
					city.Name = cityName.String
				} else {
					city.Name = "Не указан"
				}
				if cityRegion.Valid {
					city.Region = cityRegion.String
				} else {
					city.Region = "Не указан"
				}
				if cityCreatedAt.Valid {
					city.CreatedAt = cityCreatedAt.Time
				}
				vet.City = &city
			}

			if !d.hasCompleteRequiredData(&vet) {
				vet.IsActive = false
			}
		}

		vets = append(vets, &vet)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по результатам: %v", err)
	}

	log.Printf("Успешно загружено %d врачей", len(vets))
	return vets, nil
}

// hasCompleteRequiredData проверяет, что у врача заполнены все обязательные поля
func (d *Database) hasCompleteRequiredData(vet *models.Veterinarian) bool {
	// Обязательные поля: имя, фамилия, телефон
	if strings.TrimSpace(vet.FirstName) == "" {
		return false
	}
	if strings.TrimSpace(vet.LastName) == "" {
		return false
	}
	if strings.TrimSpace(vet.Phone) == "" {
		return false
	}

	return true
}

func (d *Database) GetVeterinarianByID(id int) (*models.Veterinarian, error) {
	query := `SELECT id, first_name, last_name, patronymic, phone, email, description, experience_years, is_active, city_id, created_at 
              FROM veterinarians WHERE id = $1`

	var vet models.Veterinarian
	var cityID sql.NullInt64
	var patronymic, email, description sql.NullString
	var experienceYears sql.NullInt64

	err := d.db.QueryRow(query, id).Scan(&vet.ID, &vet.FirstName, &vet.LastName,
		&patronymic, &vet.Phone, &email, &description, &experienceYears,
		&vet.IsActive, &cityID, &vet.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Заполняем nullable поля
	vet.Patronymic = patronymic
	vet.Email = email
	vet.Description = description
	vet.ExperienceYears = experienceYears
	vet.CityID = cityID

	// Загружаем информацию о городе если есть
	if cityID.Valid {
		city, err := d.GetCityByID(int(cityID.Int64))
		if err == nil {
			vet.City = city
		}
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

// GetClinicsByCity возвращает клиники по городу
func (d *Database) GetClinicsByCity(cityID int) ([]*models.Clinic, error) {
	query := `
        SELECT c.id, c.name, c.address, c.phone, c.working_hours, 
               c.is_active, c.city_id, c.district, c.metro_station, c.created_at
        FROM clinics c
        WHERE c.city_id = $1 AND c.is_active = true
        ORDER BY c.name`

	rows, err := d.db.Query(query, cityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clinics []*models.Clinic
	for rows.Next() {
		var clinic models.Clinic
		err := rows.Scan(&clinic.ID, &clinic.Name, &clinic.Address, &clinic.Phone,
			&clinic.WorkingHours, &clinic.IsActive, &clinic.CityID,
			&clinic.District, &clinic.MetroStation, &clinic.CreatedAt)
		if err != nil {
			return nil, err
		}
		clinics = append(clinics, &clinic)
	}

	return clinics, nil
}

// FindVetsByCity ищет врачей по городу с дополнительными критериями
// FindVetsByCity ищет врачей по городу с дополнительными критериями
func (d *Database) FindVetsByCity(criteria *models.SearchCriteria) ([]*models.Veterinarian, error) {
	query := `
        SELECT DISTINCT v.id, v.first_name, v.last_name, v.phone, v.email, 
               v.description, v.experience_years, v.is_active, v.city_id, v.created_at,
               c.id, c.name, c.region, c.created_at
        FROM veterinarians v
        LEFT JOIN cities c ON v.city_id = c.id
        WHERE v.is_active = true AND v.city_id = $1`

	args := []interface{}{criteria.CityID}

	// Дополнительные критерии поиска
	if criteria.SpecializationID > 0 {
		query += ` AND EXISTS (
            SELECT 1 FROM vet_specializations vs 
            WHERE vs.vet_id = v.id AND vs.specialization_id = $2
        )`
		args = append(args, criteria.SpecializationID)
	}

	if criteria.DayOfWeek > 0 {
		query += ` AND EXISTS (
            SELECT 1 FROM schedules s 
            WHERE s.vet_id = v.id AND s.day_of_week = $3 AND s.is_available = true
        )`
		args = append(args, criteria.DayOfWeek)
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
		var cityID sql.NullInt64
		var city models.City
		var cityCreatedAt time.Time

		err := rows.Scan(&vet.ID, &vet.FirstName, &vet.LastName, &vet.Phone, &vet.Email,
			&vet.Description, &vet.ExperienceYears, &vet.IsActive, &cityID, &vet.CreatedAt,
			&city.ID, &city.Name, &city.Region, &cityCreatedAt)
		if err != nil {
			return nil, err
		}

		vet.CityID = cityID
		if cityID.Valid {
			city.CreatedAt = cityCreatedAt
			vet.City = &city
		}

		// Загружаем специализации для каждого врача
		specs, err := d.GetSpecializationsByVetID(models.GetVetIDAsIntOrZero(&vet))
		if err == nil {
			vet.Specializations = specs
		}

		veterinarians = append(veterinarians, &vet)
	}

	return veterinarians, nil
}

// GetCitiesByRegion возвращает города по региону
func (d *Database) GetCitiesByRegion(region string) ([]*models.City, error) {
	query := `SELECT id, name, region, created_at FROM cities WHERE LOWER(region) LIKE LOWER($1) ORDER BY name`
	rows, err := d.db.Query(query, "%"+region+"%")
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

// SearchCities ищет города по названию
func (d *Database) SearchCities(queryStr string) ([]*models.City, error) {
	query := `SELECT id, name, region, created_at FROM cities WHERE LOWER(name) LIKE LOWER($1) ORDER BY name`
	rows, err := d.db.Query(query, "%"+queryStr+"%")
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

// UpdateVeterinarian обновляет данные врача
func (d *Database) UpdateVeterinarian(vet *models.Veterinarian) error {
	query := `UPDATE veterinarians SET 
		first_name = $1, last_name = $2, phone = $3, email = $4, 
		description = $5, experience_years = $6, is_active = $7, city_id = $8
		WHERE id = $9`

	_, err := d.db.Exec(query,
		vet.FirstName, vet.LastName, vet.Phone, vet.Email,
		vet.Description, vet.ExperienceYears, vet.IsActive, vet.CityID, vet.ID,
	)
	return err
}

func (d *Database) CreateVeterinarian(vet *models.Veterinarian) error {
	// Сначала проверяем, нет ли уже врача с таким именем и телефоном
	var existingID int
	err := d.db.QueryRow(
		"SELECT id FROM veterinarians WHERE first_name = $1 AND last_name = $2 AND phone = $3",
		vet.FirstName, vet.LastName, vet.Phone,
	).Scan(&existingID)

	if err == nil {
		return fmt.Errorf("врач с такими данными уже существует (ID: %d)", existingID)
	}
	if err != sql.ErrNoRows {
		return err
	}

	query := `INSERT INTO veterinarians 
        (first_name, last_name, patronymic, phone, email, description, experience_years, is_active, city_id) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
        RETURNING id, created_at`

	return d.db.QueryRow(query,
		vet.FirstName, vet.LastName, vet.Patronymic, vet.Phone, vet.Email,
		vet.Description, vet.ExperienceYears, vet.IsActive, vet.CityID,
	).Scan(&vet.ID, &vet.CreatedAt)
}

// DeleteCity удаляет город по ID
func (d *Database) DeleteCity(id int) error {
	query := "DELETE FROM cities WHERE id = $1"
	_, err := d.db.Exec(query, id)
	return err
}

// UpdateCity обновляет данные города
func (d *Database) UpdateCity(city *models.City) error {
	query := "UPDATE cities SET name = $1, region = $2 WHERE id = $3"
	_, err := d.db.Exec(query, city.Name, city.Region, city.ID)
	return err
}

// GetVeterinarianWithDetails возвращает врача с полной информацией о городе и клиниках
func (d *Database) GetVeterinarianWithDetails(id int) (*models.Veterinarian, error) {
	// Получаем основную информацию о враче с городом
	query := `
        SELECT v.id, v.first_name, v.last_name, v.phone, v.email, 
               v.description, v.experience_years, v.is_active, v.city_id, v.created_at,
               c.id, c.name, c.region, c.created_at
        FROM veterinarians v
        LEFT JOIN cities c ON v.city_id = c.id
        WHERE v.id = $1`

	var vet models.Veterinarian
	var cityID sql.NullInt64
	var city models.City
	var cityCreatedAt time.Time
	var vetID sql.NullInt64

	err := d.db.QueryRow(query, id).Scan(
		&vetID, &vet.FirstName, &vet.LastName, &vet.Phone, &vet.Email,
		&vet.Description, &vet.ExperienceYears, &vet.IsActive, &cityID, &vet.CreatedAt,
		&city.ID, &city.Name, &city.Region, &cityCreatedAt,
	)
	if err != nil {
		return nil, err
	}

	vet.ID = vetID
	vet.CityID = cityID
	if cityID.Valid {
		city.CreatedAt = cityCreatedAt
		vet.City = &city
	}

	// Загружаем специализации
	specs, err := d.GetSpecializationsByVetID(models.GetVetIDAsIntOrZero(&vet))
	if err == nil {
		vet.Specializations = specs
	}

	// Загружаем расписание и клиники
	schedules, err := d.GetSchedulesByVetID(models.GetVetIDAsIntOrZero(&vet))
	if err == nil {
		vet.Schedules = schedules
	}
	// Загружаем клиники врача
	clinics, err := d.GetClinicsByVetID(models.GetVetIDAsIntOrZero(&vet))
	if err == nil {
		vet.Clinics = clinics
	}
	return &vet, nil
}

// GetSpecializationByName возвращает специализацию по имени
func (d *Database) GetSpecializationByName(name string) (*models.Specialization, error) {
	query := `SELECT id, name, description, created_at FROM specializations WHERE LOWER(name) = LOWER($1)`
	var spec models.Specialization
	err := d.db.QueryRow(query, name).Scan(&spec.ID, &spec.Name, &spec.Description, &spec.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

// CreateSpecialization создает новую специализацию
func (d *Database) CreateSpecialization(spec *models.Specialization) error {
	query := `INSERT INTO specializations (name, description, created_at) 
              VALUES ($1, $2, $3) RETURNING id`
	return d.db.QueryRow(query, spec.Name, spec.Description, spec.CreatedAt).Scan(&spec.ID)
}

// AddVeterinarianSpecialization добавляет специализацию врачу
func (d *Database) AddVeterinarianSpecialization(vetID int, specID int) error {
	query := `INSERT INTO vet_specializations (vet_id, specialization_id) VALUES ($1, $2)`
	_, err := d.db.Exec(query, vetID, specID)
	return err
}

// SearchCitiesByRegion ищет города по региону
func (d *Database) SearchCitiesByRegion(region string) ([]*models.City, error) {
	query := `SELECT id, name, region, created_at FROM cities 
              WHERE region ILIKE $1 ORDER BY name`

	rows, err := d.db.Query(query, "%"+region+"%")
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

// Добавьте методы для работы с отзывами в структуру Database

func (d *Database) CreateReview(review *models.Review) error {
	repo := NewReviewRepository(d.db)
	return repo.CreateReview(review)
}

func (d *Database) GetReviewByID(reviewID int) (*models.Review, error) {
	repo := NewReviewRepository(d.db)
	return repo.GetReviewByID(reviewID)
}

func (d *Database) GetApprovedReviewsByVet(vetID int) ([]*models.Review, error) {
	repo := NewReviewRepository(d.db)
	return repo.GetApprovedReviewsByVet(vetID)
}

func (d *Database) GetPendingReviews() ([]*models.Review, error) {
	repo := NewReviewRepository(d.db)
	return repo.GetPendingReviews()
}

func (d *Database) UpdateReviewStatus(reviewID int, status string, moderatorID int) error {
	repo := NewReviewRepository(d.db)
	return repo.UpdateReviewStatus(reviewID, status, moderatorID)
}

func (d *Database) HasUserReviewForVet(userID int, vetID int) (bool, error) {
	repo := NewReviewRepository(d.db)
	return repo.HasUserReviewForVet(userID, vetID)
}

func (d *Database) GetReviewStats(vetID int) (*models.ReviewStats, error) {
	repo := NewReviewRepository(d.db)
	return repo.GetReviewStats(vetID)
}

// DebugSpecializationVetsCount - диагностическая функция для отладки количества врачей по специализациям
func (d *Database) DebugSpecializationVetsCount() (map[int]int, error) {
	query := `
        SELECT vs.specialization_id, COUNT(DISTINCT v.id) as vet_count 
        FROM vet_specializations vs 
        LEFT JOIN veterinarians v ON vs.vet_id = v.id AND v.is_active = true
        GROUP BY vs.specialization_id 
        ORDER BY vs.specialization_id`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var specID, count int
		err := rows.Scan(&specID, &count)
		if err != nil {
			return nil, err
		}
		result[specID] = count
	}

	return result, nil
}

// ========== СТАТИСТИЧЕСКИЕ МЕТОДЫ ==========

// GetActiveClinicCount возвращает количество активных клиник
func (db *Database) GetActiveClinicCount() (int, error) {
	query := "SELECT COUNT(*) FROM clinics WHERE is_active = true"
	var count int
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества активных клиник: %v", err)
	}
	return count, nil
}

// GetTotalClinicCount возвращает общее количество клиник
func (db *Database) GetTotalClinicCount() (int, error) {
	query := "SELECT COUNT(*) FROM clinics"
	var count int
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения общего количества клиник: %v", err)
	}
	return count, nil
}

// GetActiveVetCount возвращает количество активных врачей
func (db *Database) GetActiveVetCount() (int, error) {
	query := "SELECT COUNT(*) FROM veterinarians WHERE is_active = true"
	var count int
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества активных врачей: %v", err)
	}
	return count, nil
}

// GetTotalVetCount возвращает общее количество врачей
func (db *Database) GetTotalVetCount() (int, error) {
	query := "SELECT COUNT(*) FROM veterinarians"
	var count int
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения общего количества врачей: %v", err)
	}
	return count, nil
}

// GetUserCount возвращает количество пользователей
func (db *Database) GetUserCount() (int, error) {
	query := "SELECT COUNT(*) FROM users"
	var count int
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества пользователей: %v", err)
	}
	return count, nil
}

// GetRequestCount возвращает количество запросов
func (db *Database) GetRequestCount() (int, error) {
	query := "SELECT COUNT(*) FROM user_requests"
	var count int
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества запросов: %v", err)
	}
	return count, nil
}

// GetCitiesCount возвращает количество городов
func (db *Database) GetCitiesCount() (int, error) {
	query := "SELECT COUNT(*) FROM cities"
	var count int
	err := db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества городов: %v", err)
	}
	return count, nil
}

// GetVetsCountByCity возвращает количество врачей в городе
func (db *Database) GetVetsCountByCity(cityID int) (int, error) {
	query := "SELECT COUNT(*) FROM veterinarians WHERE city_id = $1"
	var count int
	err := db.db.QueryRow(query, cityID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества врачей в городе: %v", err)
	}
	return count, nil
}

// GetClinicsCountByCity возвращает количество клиник в городе
func (db *Database) GetClinicsCountByCity(cityID int) (int, error) {
	query := "SELECT COUNT(*) FROM clinics WHERE city_id = $1"
	var count int
	err := db.db.QueryRow(query, cityID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения количества клиник в городе: %v", err)
	}
	return count, nil
}

// ========== МЕТОДЫ ДЛЯ УДАЛЕНИЯ ==========

// ========== МЕТОДЫ ДЛЯ ОБНОВЛЕНИЯ ПОЛЕЙ ==========

func (db *Database) UpdateVeterinarianField(vetID int, field string, value interface{}) error {
	var query string
	var args []interface{}

	switch field {
	case "first_name", "last_name", "phone":
		query = fmt.Sprintf("UPDATE veterinarians SET %s = $1 WHERE id = $2", field)
		args = []interface{}{value, vetID}
	case "patronymic", "email", "description":
		if value == nil || value == "" {
			query = fmt.Sprintf("UPDATE veterinarians SET %s = NULL WHERE id = $1", field)
			args = []interface{}{vetID}
		} else {
			query = fmt.Sprintf("UPDATE veterinarians SET %s = $1 WHERE id = $2", field)
			args = []interface{}{value, vetID}
		}
	case "experience_years":
		if value == nil {
			query = "UPDATE veterinarians SET experience_years = NULL WHERE id = $1"
			args = []interface{}{vetID}
		} else {
			query = "UPDATE veterinarians SET experience_years = $1 WHERE id = $2"
			args = []interface{}{value, vetID}
		}
	case "is_active":
		query = "UPDATE veterinarians SET is_active = $1 WHERE id = $2"
		args = []interface{}{value, vetID}
	case "city_id":
		if value == nil {
			query = "UPDATE veterinarians SET city_id = NULL WHERE id = $1"
			args = []interface{}{vetID}
		} else {
			query = "UPDATE veterinarians SET city_id = $1 WHERE id = $2"
			args = []interface{}{value, vetID}
		}
	default:
		return fmt.Errorf("неизвестное поле: %s", field)
	}

	result, err := db.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("ошибка обновления поля %s: %v", field, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки обновленных строк: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("врач с ID %d не найден", vetID)
	}

	return nil
}

// UpdateClinicField обновляет конкретное поле клиники
func (db *Database) UpdateClinicField(clinicID int, field string, value interface{}) error {
	var query string
	var args []interface{}

	switch field {
	case "name", "address":
		query = fmt.Sprintf("UPDATE clinics SET %s = $1 WHERE id = $2", field)
		args = []interface{}{value, clinicID}
	case "phone", "working_hours":
		if value == nil || value == "" {
			query = fmt.Sprintf("UPDATE clinics SET %s = NULL WHERE id = $1", field)
			args = []interface{}{clinicID}
		} else {
			query = fmt.Sprintf("UPDATE clinics SET %s = $1 WHERE id = $2", field)
			args = []interface{}{value, clinicID}
		}
	case "is_active":
		query = "UPDATE clinics SET is_active = $1 WHERE id = $2"
		args = []interface{}{value, clinicID}
	case "city_id":
		if value == nil {
			query = "UPDATE clinics SET city_id = NULL WHERE id = $1"
			args = []interface{}{clinicID}
		} else {
			query = "UPDATE clinics SET city_id = $1 WHERE id = $2"
			args = []interface{}{value, clinicID}
		}
	default:
		return fmt.Errorf("неизвестное поле: %s", field)
	}

	result, err := db.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("ошибка обновления поля %s: %v", field, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки обновленных строк: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("клиника с ID %d не найдена", clinicID)
	}

	return nil
}

// DeleteClinic удаляет клинику
func (db *Database) DeleteClinic(clinicID int) error {
	query := "DELETE FROM clinics WHERE id = $1"
	result, err := db.db.ExecContext(context.Background(), query, clinicID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении клиники: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке удаленных строк: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("клиника с ID %d не найдена", clinicID)
	}

	return nil
}

// DeleteVeterinarian удаляет ветеринара
func (db *Database) DeleteVeterinarian(vetID int) error {
	// Сначала удаляем связанные записи из таблицы связей
	deleteSpecsQuery := "DELETE FROM veterinarian_specializations WHERE veterinarian_id = $1"
	_, err := db.db.ExecContext(context.Background(), deleteSpecsQuery, vetID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении специализаций врача: %v", err)
	}

	// Затем удаляем расписание
	deleteScheduleQuery := "DELETE FROM schedules WHERE vet_id = $1"
	_, err = db.db.ExecContext(context.Background(), deleteScheduleQuery, vetID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении расписания врача: %v", err)
	}

	// Затем удаляем самого ветеринара
	deleteVetQuery := "DELETE FROM veterinarians WHERE id = $1"
	result, err := db.db.ExecContext(context.Background(), deleteVetQuery, vetID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении ветеринара: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при проверке удаленных строк: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ветеринар с ID %d не найден", vetID)
	}

	return nil
}

// ========== МЕТОДЫ ДЛЯ РАБОТЫ С ПОЛЬЗОВАТЕЛЯМИ ==========

// GetUserByID возвращает пользователя по ID
func (d *Database) GetUserByID(userID int) (*models.User, error) {
	query := `SELECT id, telegram_id, username, first_name, last_name, phone, created_at 
              FROM users WHERE id = $1`

	var user models.User
	err := d.db.QueryRow(query, userID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByTelegramID возвращает пользователя по Telegram ID
func (d *Database) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	query := `SELECT id, telegram_id, username, first_name, last_name, phone, created_at 
              FROM users WHERE telegram_id = $1`

	var user models.User
	err := d.db.QueryRow(query, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetClinicsByVetID возвращает клиники, в которых работает врач
func (d *Database) GetClinicsByVetID(vetID int) ([]*models.Clinic, error) {
	query := `
        SELECT c.id, c.name, c.address, c.phone, c.working_hours, 
               c.is_active, c.city_id, c.district, c.metro_station, c.created_at
        FROM clinics c
        INNER JOIN vet_clinics vc ON c.id = vc.clinic_id
        WHERE vc.vet_id = $1 AND c.is_active = true
        ORDER BY c.name`

	rows, err := d.db.Query(query, vetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clinics []*models.Clinic
	for rows.Next() {
		var clinic models.Clinic
		var cityID sql.NullInt64
		var phone, workingHours sql.NullString

		err := rows.Scan(&clinic.ID, &clinic.Name, &clinic.Address, &phone,
			&workingHours, &clinic.IsActive, &cityID,
			&clinic.District, &clinic.MetroStation, &clinic.CreatedAt)
		if err != nil {
			return nil, err
		}

		clinic.Phone = phone
		clinic.WorkingHours = workingHours
		clinic.CityID = cityID
		clinics = append(clinics, &clinic)
	}

	return clinics, nil
}

// AddVetToClinic добавляет врача в клинику
func (d *Database) AddVetToClinic(vetID int, clinicID int) error {
	query := `INSERT INTO vet_clinics (vet_id, clinic_id) VALUES ($1, $2) 
              ON CONFLICT (vet_id, clinic_id) DO NOTHING`
	_, err := d.db.Exec(query, vetID, clinicID)
	return err
}

// RemoveVetFromClinic удаляет врача из клиники
func (d *Database) RemoveVetFromClinic(vetID int, clinicID int) error {
	query := `DELETE FROM vet_clinics WHERE vet_id = $1 AND clinic_id = $2`
	_, err := d.db.Exec(query, vetID, clinicID)
	return err
}

// GetVetsByClinic возвращает врачей работающих в клинике
func (d *Database) GetVetsByClinic(clinicID int) ([]*models.Veterinarian, error) {
	query := `
        SELECT DISTINCT v.id, v.first_name, v.last_name, v.phone, v.email, 
               v.description, v.experience_years, v.is_active, v.city_id, v.created_at
        FROM veterinarians v
        INNER JOIN vet_clinics vc ON v.id = vc.vet_id
        WHERE vc.clinic_id = $1 AND v.is_active = true
        ORDER BY v.first_name, v.last_name`

	rows, err := d.db.Query(query, clinicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var veterinarians []*models.Veterinarian
	for rows.Next() {
		var vet models.Veterinarian
		var vetID sql.NullInt64
		var email, description sql.NullString
		var experienceYears sql.NullInt64
		var cityID sql.NullInt64

		err := rows.Scan(&vetID, &vet.FirstName, &vet.LastName, &vet.Phone, &email,
			&description, &experienceYears, &vet.IsActive, &cityID, &vet.CreatedAt)
		if err != nil {
			return nil, err
		}

		vet.ID = vetID
		vet.Email = email
		vet.Description = description
		vet.ExperienceYears = experienceYears
		vet.CityID = cityID

		// Загружаем специализации для каждого врача
		specs, err := d.GetSpecializationsByVetID(models.GetVetIDAsIntOrZero(&vet))
		if err == nil {
			vet.Specializations = specs
		}

		veterinarians = append(veterinarians, &vet)
	}

	return veterinarians, nil
}

// GetAllActiveVeterinarians возвращает всех активных ветеринаров
func (d *Database) GetAllActiveVeterinarians() ([]*models.Veterinarian, error) {
	query := `
        SELECT v.id, v.first_name, v.last_name, v.phone, v.email, 
               v.experience_years, v.description, v.city_id, v.is_active, v.created_at,
               c.id, c.name, c.region, c.created_at
        FROM veterinarians v
        LEFT JOIN cities c ON v.city_id = c.id
        WHERE v.is_active = true
        ORDER BY v.first_name, v.last_name
    `

	rows, err := d.db.Query(query) // Исправлено: d.db.Query вместо db.Query
	if err != nil {
		return nil, fmt.Errorf("error querying active veterinarians: %v", err)
	}
	defer rows.Close()

	var vets []*models.Veterinarian
	for rows.Next() {
		var vet models.Veterinarian
		var city models.City
		var email, description sql.NullString
		var experienceYears sql.NullInt64
		var cityID sql.NullInt64
		var vetID sql.NullInt64 // Добавлено для ID врача

		err := rows.Scan(
			&vetID, &vet.FirstName, &vet.LastName, &vet.Phone, &email,
			&experienceYears, &description, &cityID, &vet.IsActive, &vet.CreatedAt,
			&city.ID, &city.Name, &city.Region, &city.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning veterinarian: %v", err)
		}

		vet.ID = vetID
		vet.Email = email
		vet.ExperienceYears = experienceYears
		vet.Description = description
		vet.CityID = cityID

		if city.ID != 0 {
			vet.City = &city
		}

		// Загружаем специализации для врача
		vetIDInt := models.GetVetIDAsIntOrZero(&vet) // Исправлено: используем функцию для получения ID
		specs, err := d.GetSpecializationsByVetID(vetIDInt)
		if err == nil {
			vet.Specializations = specs
		}

		vets = append(vets, &vet)
	}

	return vets, nil
}
