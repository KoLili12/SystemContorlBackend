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
		// ============================================
		// ПУБЛИЧНЫЕ ЭНДПОИНТЫ (без аутентификации)
		// ============================================
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register) // Регистрация
			auth.POST("/login", handlers.Login)       // Вход
		}

		// ============================================
		// ЗАЩИЩЕННЫЕ ЭНДПОИНТЫ (требуют аутентификации)
		// ============================================
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// === ПРОФИЛЬ ПОЛЬЗОВАТЕЛЯ ===
			protected.GET("/profile", handlers.GetProfile)

			// === РАБОТА С ФАЙЛАМИ ===
			protected.POST("/files/upload", handlers.UploadFiles)   // Загрузка файлов
			protected.GET("/files/:id", handlers.GetFile)           // Скачать файл
			protected.GET("/files", handlers.GetEntityFiles)        // Список файлов сущности
			protected.DELETE("/files/:id", handlers.DeleteFile)     // Удалить файл

			// === ПРОЕКТЫ (доступ для всех авторизованных) ===
			protected.GET("/projects", handlers.GetProjects)              // Список проектов
			protected.GET("/projects/:id", handlers.GetProject)           // Один проект
			protected.GET("/projects/:id/files", handlers.GetProjectFiles)       // Файлы проекта
			protected.GET("/projects/:id/image", handlers.GetProjectMainImage)   // Главное изображение
			protected.GET("/projects/:id/defects/stats", handlers.GetProjectDefectsStats) // Статистика по дефектам

			// === ДЕФЕКТЫ (доступ для всех авторизованных) ===
			protected.GET("/defects", handlers.GetDefects)     // Список дефектов (с фильтрацией)
			protected.GET("/defects/:id", handlers.GetDefect)  // Один дефект

			// ============================================
			// ЭНДПОИНТЫ ТОЛЬКО ДЛЯ МЕНЕДЖЕРОВ
			// ============================================
			manager := protected.Group("/")
			manager.Use(middleware.RoleMiddleware(models.RoleManager))
			{
				// Управление проектами
				manager.POST("/projects", handlers.CreateProject)       // Создание проекта
				manager.PUT("/projects/:id", handlers.UpdateProject)    // Обновление проекта
				manager.DELETE("/projects/:id", handlers.DeleteProject) // Удаление проекта
			}

			// ============================================
			// ЭНДПОИНТЫ ДЛЯ ИНЖЕНЕРОВ
			// ============================================
			engineer := protected.Group("/")
			engineer.Use(middleware.RoleMiddleware(models.RoleEngineer))
			{
				// Управление дефектами
				engineer.POST("/defects", handlers.CreateDefect)                      // Создание дефекта
				engineer.PUT("/defects/:id", handlers.UpdateDefect)                   // Обновление дефекта
				engineer.PATCH("/defects/:id/status", handlers.UpdateDefectStatus)    // Изменение статуса
				engineer.DELETE("/defects/:id", handlers.DeleteDefect)                // Удаление дефекта
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