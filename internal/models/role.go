package models

import (
	"time"

	"gorm.io/gorm"
)

// Role - роль пользователя согласно ТЗ
type Role struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"size:100;not null;unique" json:"name"`
	Code      string         `gorm:"size:50;not null;unique" json:"code"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Константы ролей согласно ТЗ
const (
	RoleManager  = "manager"  // Менеджер - назначение задач, контроль сроков, формирование отчётов
	RoleEngineer = "engineer" // Инженер - регистрация дефектов, обновление информации
	RoleObserver = "observer" // Наблюдатель - просмотр прогресса и отчётности
)
