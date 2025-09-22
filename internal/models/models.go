package models

import (
	"database/sql"
	"time"
)

// Veterinarian представляет модель ветеринарного врача в системе
type Veterinarian struct {
	ID        int64     `json:"id"`         // Уникальный идентификатор врача
	Name      string    `json:"name"`       // ФИО врача
	Specialty string    `json:"specialty"`  // Специализация (терапевт, хирург и т.д.)
	Address   string    `json:"address"`    // Адрес приема
	Phone     string    `json:"phone"`      // Контактный телефон
	WorkHours string    `json:"work_hours"` // График работы
	CreatedAt time.Time `json:"created_at"` // Время создания записи
	UpdatedAt time.Time `json:"updated_at"` // Время последнего обновления
}

// User представляет модель пользователя Telegram бота
type User struct {
	ID       int64  `json:"id"`       // Уникальный идентификатор пользователя
	Username string `json:"username"` // Имя пользователя в Telegram
	ChatID   int64  `json:"chat_id"`  // ID чата с пользователем
	IsAdmin  bool   `json:"is_admin"` // Флаг администратора
}

// CreateVeterinarianRequest структура для создания нового врача
type CreateVeterinarianRequest struct {
	Name      string `json:"name" binding:"required"`
	Specialty string `json:"specialty" binding:"required"`
	Address   string `json:"address" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
	WorkHours string `json:"work_hours"`
}

// UpdateVeterinarianRequest структура для обновления данных врача
type UpdateVeterinarianRequest struct {
	Name      string `json:"name"`
	Specialty string `json:"specialty"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	WorkHours string `json:"work_hours"`
}

// ToVeterinarian преобразует CreateVeterinarianRequest в Veterinarian
func (r *CreateVeterinarianRequest) ToVeterinarian() *Veterinarian {
	return &Veterinarian{
		Name:      r.Name,
		Specialty: r.Specialty,
		Address:   r.Address,
		Phone:     r.Phone,
		WorkHours: r.WorkHours,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Scan считывает данные врача из SQL строки
func (v *Veterinarian) Scan(rows *sql.Rows) error {
	return rows.Scan(
		&v.ID,
		&v.Name,
		&v.Specialty,
		&v.Address,
		&v.Phone,
		&v.WorkHours,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
}
