package models

import (
	"time"

	"gorm.io/gorm"
)

// Attachment - модель для хранения файлов (фото, документы)
type Attachment struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	FileName     string         `gorm:"size:255;not null" json:"file_name"`
	OriginalName string         `gorm:"size:255;not null" json:"original_name"`
	FilePath     string         `gorm:"size:500;not null" json:"file_path"`
	FileSize     int64          `json:"file_size"`
	ContentType  string         `gorm:"size:100" json:"content_type"`
	FileType     string         `gorm:"size:50" json:"file_type"` // "image", "document"
	
	// Полиморфные связи - к чему прикреплен файл
	EntityType   string         `gorm:"size:50;not null" json:"entity_type"` // "project", "defect"
	EntityID     uint           `gorm:"not null" json:"entity_id"`
	
	UploadedBy   uint           `json:"uploaded_by"`
	Uploader     User           `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Константы типов сущностей
const (
	EntityTypeProject = "project"
	EntityTypeDefect  = "defect"
)

// Константы типов файлов
const (
	FileTypeImage    = "image"
	FileTypeDocument = "document"
)

// AttachmentResponse - ответ с информацией о файле
type AttachmentResponse struct {
	ID           uint   `json:"id"`
	FileName     string `json:"file_name"`
	OriginalName string `json:"original_name"`
	FileSize     int64  `json:"file_size"`
	ContentType  string `json:"content_type"`
	FileType     string `json:"file_type"`
	UploadedBy   uint   `json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
	URL          string `json:"url"` // Полный URL для доступа к файлу
}