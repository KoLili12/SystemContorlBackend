package handlers

import (
	"net/http"

	"SystemContorlBackend/internal/models"
	"SystemContorlBackend/internal/services"

	"github.com/gin-gonic/gin"
)

// Register обрабатывает регистрацию пользователя
func Register(c *gin.Context) {
	var userData models.UserRegister

	// Валидация входных данных
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Регистрация пользователя
	user, err := services.RegisterUser(userData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Генерируем токен для нового пользователя
	token, err := services.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Пользователь успешно зарегистрирован",
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"phone":      user.Phone,
			"role": gin.H{
				"id":   user.Role.ID,
				"name": user.Role.Name,
				"code": user.Role.Code,
			},
		},
		"token": token,
	})
}

// Login обрабатывает вход пользователя
func Login(c *gin.Context) {
	var loginData models.UserLogin

	// Валидация входных данных
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Выполняем вход
	user, token, err := services.LoginUser(loginData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Успешный вход",
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"phone":      user.Phone,
			"role": gin.H{
				"id":   user.Role.ID,
				"name": user.Role.Name,
				"code": user.Role.Code,
			},
		},
		"token": token,
	})
}

// GetProfile получает профиль текущего пользователя
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := services.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"phone":      user.Phone,
			"role": gin.H{
				"id":   user.Role.ID,
				"name": user.Role.Name,
				"code": user.Role.Code,
			},
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt,
		},
	})
}
