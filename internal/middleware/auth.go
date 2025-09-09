package middleware

import (
	"net/http"
	"strings"

	"SystemContorlBackend/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет JWT токен
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		// Проверяем формат: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
			c.Abort()
			return
		}

		// Валидируем токен
		claims, err := services.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный или истекший токен"})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role_code", claims.RoleCode)
		c.Next()
	}
}

// RoleMiddleware проверяет роль пользователя (разграничение прав доступа согласно ТЗ)
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode, exists := c.Get("role_code")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Роль не определена"})
			c.Abort()
			return
		}

		userRole := roleCode.(string)
		allowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав для выполнения операции"})
			c.Abort()
			return
		}

		c.Next()
	}
}
