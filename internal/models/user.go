package models

import (
	"time"

	"gorm.io/gorm"
)

// User - модель пользователя для регистрации и аутентификации
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Email     string         `gorm:"size:255;not null;unique" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"` // Пароль не отдаем в JSON
	FirstName string         `gorm:"size:100;not null" json:"first_name"`
	LastName  string         `gorm:"size:100;not null" json:"last_name"`
	Phone     string         `gorm:"size:20" json:"phone"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	RoleID    uint           `json:"role_id"`
	Role      Role           `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserLogin - структура для входа
type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserRegister - структура для регистрации
type UserRegister struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	RoleCode  string `json:"role_code" binding:"required,oneof=manager engineer observer"`
}
