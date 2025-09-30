package handlers

import (
	"net/http"
	"strconv"

	"SystemContorlBackend/internal/models"
	"SystemContorlBackend/internal/services"

	"github.com/gin-gonic/gin"
)

// CreateProject создает новый проект (только для менеджеров)
func CreateProject(c *gin.Context) {
	var projectData models.ProjectCreate

	// Валидация входных данных
	if err := c.ShouldBindJSON(&projectData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID текущего пользователя из контекста
	userID, _ := c.Get("user_id")

	// Создаем проект
	project, err := services.CreateProject(projectData, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Проект успешно создан",
		"project": project,
	})
}

// GetProjects получает список проектов (доступно всем ролям)
func GetProjects(c *gin.Context) {
	// Параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	projects, total, err := services.GetProjects(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Подсчет общего количества страниц
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  total,
			"limit":        limit,
		},
	})
}

// GetProject получает проект по ID (доступно всем ролям)
func GetProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID проекта"})
		return
	}

	project, err := services.GetProjectByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": project,
	})
}

// UpdateProject обновляет проект (только для менеджеров)
func UpdateProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID проекта"})
		return
	}

	var updateData models.ProjectUpdate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := services.UpdateProject(uint(id), updateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Проект успешно обновлен",
		"project": project,
	})
}

// DeleteProject удаляет проект (только для менеджеров)
func DeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID проекта"})
		return
	}

	err = services.DeleteProject(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Проект успешно удален",
	})

}