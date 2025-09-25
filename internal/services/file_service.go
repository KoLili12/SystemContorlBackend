package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"SystemContorlBackend/internal/database"
	"SystemContorlBackend/internal/models"
)

const (
	// Максимальный размер файла: 10MB
	MaxFileSize = 10 << 20 // 10MB
	
	// Директория для загрузки файлов
	UploadDir = "uploads"
)

// Разрешенные типы файлов
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var AllowedDocumentTypes = map[string]bool{
	"application/pdf": true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"text/plain": true,
}

// UploadFile загружает файл и сохраняет информацию в БД
func UploadFile(file *multipart.FileHeader, entityType string, entityID uint, uploadedBy uint) (*models.Attachment, error) {
	// Проверяем размер файла
	if file.Size > MaxFileSize {
		return nil, errors.New("файл слишком большой. Максимальный размер: 10MB")
	}

	// Определяем тип файла
	contentType := file.Header.Get("Content-Type")
	fileType := getFileType(contentType)
	if fileType == "" {
		return nil, errors.New("неподдерживаемый тип файла")
	}

	// Создаем директорию если не существует
	entityDir := filepath.Join(UploadDir, entityType)
	if err := os.MkdirAll(entityDir, 0755); err != nil {
		return nil, fmt.Errorf("ошибка создания директории: %v", err)
	}

	// Генерируем уникальное имя файла
	timestamp := time.Now().Unix()
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d_%d%s", timestamp, entityID, ext)
	filePath := filepath.Join(entityDir, fileName)

	// Открываем загруженный файл
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer src.Close()

	// Создаем файл на диске
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("ошибка сохранения файла: %v", err)
	}

	// Создаем запись в БД
	attachment := models.Attachment{
		FileName:     fileName,
		OriginalName: file.Filename,
		FilePath:     filePath,
		FileSize:     file.Size,
		ContentType:  contentType,
		FileType:     fileType,
		EntityType:   entityType,
		EntityID:     entityID,
		UploadedBy:   uploadedBy,
	}

	if err := database.DB.Create(&attachment).Error; err != nil {
		// Удаляем файл если не удалось сохранить в БД
		os.Remove(filePath)
		return nil, fmt.Errorf("ошибка сохранения в базе данных: %v", err)
	}

	return &attachment, nil
}

// GetAttachments получает список файлов для сущности
func GetAttachments(entityType string, entityID uint) ([]models.AttachmentResponse, error) {
	var attachments []models.Attachment
	
	if err := database.DB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Find(&attachments).Error; err != nil {
		return nil, err
	}

	// Преобразуем в response формат
	var responses []models.AttachmentResponse
	for _, att := range attachments {
		responses = append(responses, models.AttachmentResponse{
			ID:           att.ID,
			FileName:     att.FileName,
			OriginalName: att.OriginalName,
			FileSize:     att.FileSize,
			ContentType:  att.ContentType,
			FileType:     att.FileType,
			UploadedBy:   att.UploadedBy,
			CreatedAt:    att.CreatedAt,
			URL:          fmt.Sprintf("/api/v1/files/%d", att.ID),
		})
	}

	return responses, nil
}

// DeleteAttachment удаляет файл
func DeleteAttachment(id uint, userID uint) error {
	var attachment models.Attachment
	if err := database.DB.First(&attachment, id).Error; err != nil {
		return errors.New("файл не найден")
	}

	// Проверяем права (только загрузивший может удалить, или менеджер)
	// TODO: добавить проверку роли менеджера
	if attachment.UploadedBy != userID {
		return errors.New("недостаточно прав для удаления файла")
	}

	// Удаляем файл с диска
	if err := os.Remove(attachment.FilePath); err != nil {
		// Логируем ошибку, но продолжаем удаление из БД
		fmt.Printf("Ошибка удаления файла %s: %v\n", attachment.FilePath, err)
	}

	// Удаляем запись из БД
	return database.DB.Delete(&attachment).Error
}

// GetAttachmentByID получает файл по ID для скачивания
func GetAttachmentByID(id uint) (*models.Attachment, error) {
	var attachment models.Attachment
	if err := database.DB.Where("entity_id = ?", id).Find(&attachment).Error; err != nil {
		return nil, errors.New("файлы не найдены")
	}
	return &attachment, nil
}

// getFileType определяет тип файла по MIME типу
func getFileType(contentType string) string {
	if AllowedImageTypes[contentType] {
		return models.FileTypeImage
	}
	if AllowedDocumentTypes[contentType] {
		return models.FileTypeDocument
	}
	return ""
}

// IsImageFile проверяет, является ли файл изображением
func IsImageFile(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}

// ReplaceAttachment заменяет существующий файл новым
func ReplaceAttachment(id uint, file *multipart.FileHeader, userID uint) (*models.Attachment, error) {
	// Получаем старый файл
	var oldAttachment models.Attachment
	if err := database.DB.First(&oldAttachment, id).Error; err != nil {
		return nil, errors.New("файл не найден")
	}

	// Проверяем права (только загрузивший может заменить, или менеджер)  
	// TODO: добавить проверку роли менеджера
	if oldAttachment.UploadedBy != userID {
		return nil, errors.New("недостаточно прав для замены файла")
	}

	// Проверяем размер нового файла
	if file.Size > MaxFileSize {
		return nil, errors.New("файл слишком большой. Максимальный размер: 10MB")
	}

	// Определяем тип нового файла
	contentType := file.Header.Get("Content-Type")
	fileType := getFileType(contentType)
	if fileType == "" {
		return nil, errors.New("неподдерживаемый тип файла")
	}

	// Генерируем новое имя файла
	timestamp := time.Now().Unix()
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d_%d%s", timestamp, oldAttachment.EntityID, ext)
	filePath := filepath.Join(UploadDir, oldAttachment.EntityType, fileName)

	// Открываем загруженный файл
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %v", err)
	}
	defer src.Close()

	// Создаем новый файл на диске
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания файла: %v", err)
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("ошибка сохранения файла: %v", err)
	}

	// Сохраняем старый путь для удаления
	oldFilePath := oldAttachment.FilePath

	// Обновляем запись в БД
	oldAttachment.FileName = fileName
	oldAttachment.OriginalName = file.Filename
	oldAttachment.FilePath = filePath
	oldAttachment.FileSize = file.Size
	oldAttachment.ContentType = contentType
	oldAttachment.FileType = fileType

	// Обновляем в базе данных
	if err := database.DB.Save(&oldAttachment).Error; err != nil {
		// Удаляем новый файл если не удалось обновить БД
		os.Remove(filePath)
		return nil, fmt.Errorf("ошибка обновления в базе данных: %v", err)
	}

	// Удаляем старый файл с диска
	if err := os.Remove(oldFilePath); err != nil {
		// Логируем ошибку, но не прерываем операцию
		fmt.Printf("Предупреждение: не удалось удалить старый файл %s: %v\n", oldFilePath, err)
	}

	return &oldAttachment, nil
}