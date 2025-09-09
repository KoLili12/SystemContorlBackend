package database

import (
	"fmt"
	"log"
	"os"

	"SystemContorlBackend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB инициализирует подключение к базе данных
func InitDB() {
	var dsn string
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		dsn = fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_SSLMODE"),
		)
	} else {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			password,
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_SSLMODE"),
		)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Автомиграция моделей
	err = DB.AutoMigrate(
		&models.Role{},
		&models.User{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	// Создаем роли по умолчанию
	seedRoles()
}

// seedRoles создает роли согласно ТЗ
func seedRoles() {
	roles := []models.Role{
		{Name: "Менеджер", Code: models.RoleManager},
		{Name: "Инженер", Code: models.RoleEngineer},
		{Name: "Наблюдатель", Code: models.RoleObserver},
	}

	for _, role := range roles {
		var existingRole models.Role
		if err := DB.Where("code = ?", role.Code).First(&existingRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				DB.Create(&role)
				log.Printf("Role '%s' created", role.Name)
			}
		}
	}
}