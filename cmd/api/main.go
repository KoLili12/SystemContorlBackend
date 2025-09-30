package main

import (
	"log"
	"os"

	"SystemContorlBackend/internal/config"
	"SystemContorlBackend/internal/database"
	"SystemContorlBackend/internal/handlers"
	"SystemContorlBackend/internal/middleware"
	"SystemContorlBackend/internal/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию
	config.LoadConfig()

	// Инициализируем базу данных
	database.InitDB()

	// Создаем Gin роутер
	router := gin.Default()

	// Добавляем CORS middleware для работы с мобильным приложением
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API группа
	api := router.Group("/api/v1")
	{
		// Публичные эндпоинты (без аутентификации)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// Защищенные эндпоинты (требуют аутентификации)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Профиль пользователя
			protected.GET("/profile", handlers.GetProfile)

			// Работа с файлами
			protected.POST("/files/upload", handlers.UploadFiles) // Загрузка файлов
			protected.GET("/files/:id", handlers.GetFile)               // Скачать файл
			protected.GET("/files", handlers.GetEntityFiles)    // Список файлов сущности
			protected.DELETE("/files/:id", handlers.DeleteFile) // Удалить файл

			protected.GET("/projects/:id/files", handlers.GetProjectFiles)
			protected.GET("/projects/:id/image", handlers.GetProjectMainImage)

			// Проекты (доступ для всех авторизованных пользователей)
			protected.GET("/projects", handlers.GetProjects)    // Список проектов
			protected.GET("/projects/:id", handlers.GetProject) // Один проект

			// Эндпоинты только для менеджеров
			manager := protected.Group("/")
			manager.Use(middleware.RoleMiddleware(models.RoleManager))
			{
				// Управление проектами
				manager.POST("/projects", handlers.CreateProject)       // Создание проекта
				manager.PUT("/projects/:id", handlers.UpdateProject)    // Обновление проекта
				manager.DELETE("/projects/:id", handlers.DeleteProject) // Удаление проекта
			}

			// Эндпоинты для менеджеров и инженеров
			staff := protected.Group("/")
			staff.Use(middleware.RoleMiddleware(models.RoleManager, models.RoleEngineer))
			{
				// Здесь будут эндпоинты для создания и редактирования дефектов
				// staff.POST("/defects", handlers.CreateDefect)
				// staff.PUT("/defects/:id", handlers.UpdateDefect)
			}
		}
	}

	// Запускаем сервер
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на порту %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
