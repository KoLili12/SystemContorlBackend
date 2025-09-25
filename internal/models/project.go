package models

import (
	"time"

	"gorm.io/gorm"
)

// Project - модель строительного объекта/проекта
type Project struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Description string         `gorm:"text" json:"description"`
	Address     string         `gorm:"size:500" json:"address"`
	Status      string         `gorm:"size:50;default:'active'" json:"status"`
	StartDate   *time.Time     `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	CreatedBy   uint           `json:"created_by"`
	Creator     User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Attachments []Attachment   `gorm:"foreignKey:EntityID;where:entity_type = 'project'" json:"attachments,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProjectCreate - структура для создания проекта
type ProjectCreate struct {
	Name        string     `json:"name" binding:"required,min=3,max=255"`
	Description string     `json:"description"`
	Address     string     `json:"address"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// ProjectUpdate - структура для обновления проекта
type ProjectUpdate struct {
	Name        string     `json:"name" binding:"omitempty,min=3,max=255"`
	Description string     `json:"description"`
	Address     string     `json:"address"`
	Status      string     `json:"status" binding:"omitempty,oneof=active completed suspended"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// Константы статусов проекта
const (
	ProjectStatusActive    = "active"     // Активный
	ProjectStatusCompleted = "completed"  // Завершенный
	ProjectStatusSuspended = "suspended"  // Приостановленный
)