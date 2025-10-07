package handlers

import (
	"net/http"
	"strconv"

	"SystemContorlBackend/internal/models"
	"SystemContorlBackend/internal/services"

	"github.com/gin-gonic/gin"
)

// CreateDefect создает новый дефект (только для инженеров)
func CreateDefect(c *gin.Context) {
	var defectData models.DefectCreate

	// Валидация входных данных
	if err := c.ShouldBindJSON(&defectData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID текущего пользователя из контекста
	userID, _ := c.Get("user_id")

	// Создаем дефект
	defect, err := services.CreateDefect(defectData, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем в response
	response := services.ConvertToDefectResponse(defect)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Дефект успешно зарегистрирован",
		"defect":  response,
	})
}

// GetDefects получает список дефектов
func GetDefects(c *gin.Context) {
	// Параметры фильтрации и пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	priority := c.Query("priority")
	projectIDStr := c.Query("project_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Обработка project_id
	var projectID *uint
	if projectIDStr != "" {
		id, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный project_id"})
			return
		}
		idUint := uint(id)
		projectID = &idUint
	}

	defects, total, err := services.GetDefects(projectID, status, priority, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем в response
	var defectResponses []models.DefectResponse
	for _, defect := range defects {
		defectResponses = append(defectResponses, services.ConvertToDefectResponse(&defect))
	}

	// Подсчет общего количества страниц
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"defects": defectResponses,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  total,
			"limit":        limit,
		},
	})
}

// GetDefect получает дефект по ID
func GetDefect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID дефекта"})
		return
	}

	defect, err := services.GetDefectByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем в response
	response := services.ConvertToDefectResponse(defect)

	// Добавляем информацию о файлах
	files, _ := services.GetAttachments(models.EntityTypeDefect, defect.ID)

	c.JSON(http.StatusOK, gin.H{
		"defect": response,
		"files":  files,
	})
}

// UpdateDefect обновляет дефект (только создатель)
func UpdateDefect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID дефекта"})
		return
	}

	var updateData models.DefectUpdate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	defect, err := services.UpdateDefect(uint(id), updateData, userID.(uint))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем в response
	response := services.ConvertToDefectResponse(defect)

	c.JSON(http.StatusOK, gin.H{
		"message": "Дефект успешно обновлен",
		"defect":  response,
	})
}

// UpdateDefectStatus изменяет статус дефекта (только инженеры)
func UpdateDefectStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID дефекта"})
		return
	}

	var statusUpdate models.DefectStatusUpdate
	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	defect, err := services.UpdateDefectStatus(uint(id), statusUpdate.Status, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем в response
	response := services.ConvertToDefectResponse(defect)

	c.JSON(http.StatusOK, gin.H{
		"message": "Статус дефекта успешно изменен",
		"defect":  response,
	})
}

// DeleteDefect удаляет дефект (только создатель)
func DeleteDefect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID дефекта"})
		return
	}

	userID, _ := c.Get("user_id")

	err = services.DeleteDefect(uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Дефект успешно удален",
	})
}

// GetProjectDefectsStats получает статистику по дефектам проекта
func GetProjectDefectsStats(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID проекта"})
		return
	}

	stats, err := services.GetProjectDefectsStats(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project_id": projectID,
		"stats":      stats,
	})
}