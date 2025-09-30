package handlers

import (
	"net/http"
	"strconv"

	"SystemContorlBackend/internal/models"
	"SystemContorlBackend/internal/services"

	"github.com/gin-gonic/gin"
)

// GetProjectFiles получает все файлы проекта
func GetProjectFiles(c *gin.Context) {
    projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID проекта"})
        return
    }

    files, err := services.GetAttachments(models.EntityTypeProject, uint(projectID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"files": files})
}

// GetProjectMainImage возвращает первое изображение проекта
func GetProjectMainImage(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID проекта"})
		return
	}

	// Получаем файлы проекта
	files, err := services.GetAttachments(models.EntityTypeProject, uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Находим первое изображение
	var imageAttachment *models.Attachment
	for _, file := range files {
		if file.FileType == models.FileTypeImage {
			// Получаем полную информацию о файле
			att, err := services.GetAttachmentByID(file.ID)
			if err == nil {
				imageAttachment = att
				break
			}
		}
	}

	if imageAttachment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Изображение не найдено"})
		return
	}

	// Устанавливаем заголовки
	c.Header("Content-Type", imageAttachment.ContentType)
	c.Header("Content-Disposition", `inline; filename="`+imageAttachment.OriginalName+`"`)
	c.Header("Cache-Control", "public, max-age=3600")

	// Отдаем файл
	c.File(imageAttachment.FilePath)
}

// UploadFiles загружает файлы для сущности (проект, дефект)
func UploadFiles(c *gin.Context) {
	entityType := c.PostForm("entity_type")
	entityIDStr := c.PostForm("entity_id")

	// Валидация параметров
	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указаны entity_type или entity_id"})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный entity_id"})
		return
	}

	// Проверяем поддерживаемые типы сущностей
	if entityType != models.EntityTypeProject && entityType != models.EntityTypeDefect {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неподдерживаемый тип сущности"})
		return
	}

	// Получаем ID пользователя
	userID, _ := c.Get("user_id")

	// Получаем multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка обработки формы"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файлы не выбраны"})
		return
	}

	var uploadedFiles []models.AttachmentResponse
	var errors []string

	// Загружаем каждый файл
	for _, file := range files {
		attachment, err := services.UploadFile(file, entityType, uint(entityID), userID.(uint))
		if err != nil {
			errors = append(errors, file.Filename+": "+err.Error())
			continue
		}

		uploadedFiles = append(uploadedFiles, models.AttachmentResponse{
			ID:           attachment.ID,
			FileName:     attachment.FileName,
			OriginalName: attachment.OriginalName,
			FileSize:     attachment.FileSize,
			ContentType:  attachment.ContentType,
			FileType:     attachment.FileType,
			UploadedBy:   attachment.UploadedBy,
			CreatedAt:    attachment.CreatedAt,
			URL:          "/api/v1/files/" + strconv.Itoa(int(attachment.ID)),
		})
	}

	result := gin.H{
		"uploaded_files": uploadedFiles,
	}

	if len(errors) > 0 {
		result["errors"] = errors
	}

	c.JSON(http.StatusOK, result)
}

// GetFile отдает файл для скачивания
func GetFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID файла"})
		return
	}

	attachment, err := services.GetAttachmentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем заголовки для корректного отображения файла
	c.Header("Content-Type", attachment.ContentType)
	c.Header("Content-Disposition", `inline; filename="`+attachment.OriginalName+`"`)
	
	// Для изображений добавляем cache заголовки
	if services.IsImageFile(attachment.ContentType) {
		c.Header("Cache-Control", "public, max-age=3600")
	}

	// Отдаем файл
	c.File(attachment.FilePath)
}

// GetEntityFiles получает список файлов для сущности
func GetEntityFiles(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указаны entity_type или entity_id"})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный entity_id"})
		return
	}

	files, err := services.GetAttachments(entityType, uint(entityID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

func DeleteFile(c *gin.Context) {
	// Получаем ID файла из URL параметра
	attachmentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID файла"})
		return
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Удаляем файл
	err = services.DeleteAttachment(uint(attachmentID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Файл успешно удален",
	})
}

// ReplaceFile заменяет существующий файл новым
func ReplaceFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID файла"})
		return
	}

	userID, _ := c.Get("user_id")

	// Получаем загружаемый файл
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не выбран"})
		return
	}

	// Заменяем файл
	newAttachment, err := services.ReplaceAttachment(uint(id), file, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Файл успешно заменен",
		"file": models.AttachmentResponse{
			ID:           newAttachment.ID,
			FileName:     newAttachment.FileName,
			OriginalName: newAttachment.OriginalName,
			FileSize:     newAttachment.FileSize,
			ContentType:  newAttachment.ContentType,
			FileType:     newAttachment.FileType,
			UploadedBy:   newAttachment.UploadedBy,
			CreatedAt:    newAttachment.CreatedAt,
			URL:          "/api/v1/files/" + strconv.Itoa(int(newAttachment.ID)),
		},
	})
}