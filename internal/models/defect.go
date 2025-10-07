package models

import (
	"time"

	"gorm.io/gorm"
)

// Defect - модель дефекта на строительном объекте
type Defect struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	ProjectID   uint           `gorm:"not null;index" json:"project_id"`
	Project     Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Description string         `gorm:"text;not null" json:"description"`
	Status      string         `gorm:"size:50;default:'registered';index" json:"status"`
	Priority    string         `gorm:"size:50;default:'medium'" json:"priority"`
	CreatedBy   uint           `json:"created_by"`
	Creator     User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Attachments []Attachment   `gorm:"polymorphic:Entity;polymorphicValue:defect" json:"attachments,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// DefectCreate - структура для создания дефекта
type DefectCreate struct {
	ProjectID   uint   `json:"project_id" binding:"required"`
	Title       string `json:"title" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"required,min=10"`
	Priority    string `json:"priority" binding:"omitempty,oneof=low medium high critical"`
}

// DefectUpdate - структура для обновления дефекта
type DefectUpdate struct {
	Title       string `json:"title" binding:"omitempty,min=3,max=255"`
	Description string `json:"description" binding:"omitempty,min=10"`
	Priority    string `json:"priority" binding:"omitempty,oneof=low medium high critical"`
}

// DefectStatusUpdate - структура для изменения статуса
type DefectStatusUpdate struct {
	Status string `json:"status" binding:"required,oneof=registered in_progress completed"`
}

// DefectResponse - ответ с информацией о дефекте
type DefectResponse struct {
	ID          uint              `json:"id"`
	ProjectID   uint              `json:"project_id"`
	ProjectName string            `json:"project_name"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Priority    string            `json:"priority"`
	Creator     DefectCreatorInfo `json:"creator"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// DefectCreatorInfo - информация о создателе дефекта
type DefectCreatorInfo struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// Константы статусов дефекта
const (
	DefectStatusRegistered = "registered"  // Зарегистрирован
	DefectStatusInProgress = "in_progress" // В работе
	DefectStatusCompleted  = "completed"   // Завершен
)

// Константы приоритетов дефекта
const (
	DefectPriorityLow      = "low"      // Низкий
	DefectPriorityMedium   = "medium"   // Средний
	DefectPriorityHigh     = "high"     // Высокий
	DefectPriorityCritical = "critical" // Критический
)

// GetStatusDisplayName возвращает читаемое название статуса
func GetStatusDisplayName(status string) string {
	switch status {
	case DefectStatusRegistered:
		return "Зарегистрирован"
	case DefectStatusInProgress:
		return "В работе"
	case DefectStatusCompleted:
		return "Завершен"
	default:
		return status
	}
}

// GetPriorityDisplayName возвращает читаемое название приоритета
func GetPriorityDisplayName(priority string) string {
	switch priority {
	case DefectPriorityLow:
		return "Низкий"
	case DefectPriorityMedium:
		return "Средний"
	case DefectPriorityHigh:
		return "Высокий"
	case DefectPriorityCritical:
		return "Критический"
	default:
		return priority
	}
}