package services

import (
	"errors"
	"fmt"

	"SystemContorlBackend/internal/database"
	"SystemContorlBackend/internal/models"
	"gorm.io/gorm"
)

// CreateDefect создает новый дефект (только для инженеров)
func CreateDefect(defectData models.DefectCreate, createdBy uint) (*models.Defect, error) {
	// Проверяем существование проекта
	var project models.Project
	if err := database.DB.First(&project, defectData.ProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("проект не найден")
		}
		return nil, err
	}

	// Устанавливаем приоритет по умолчанию, если не указан
	priority := defectData.Priority
	if priority == "" {
		priority = models.DefectPriorityMedium
	}

	// Создаем дефект
	defect := models.Defect{
		ProjectID:   defectData.ProjectID,
		Title:       defectData.Title,
		Description: defectData.Description,
		Status:      models.DefectStatusRegistered,
		Priority:    priority,
		CreatedBy:   createdBy,
	}

	if err := database.DB.Create(&defect).Error; err != nil {
		return nil, err
	}

	// Загружаем связанные данные
	database.DB.Preload("Creator").Preload("Project").First(&defect, defect.ID)

	return &defect, nil
}

// GetDefects получает список дефектов с фильтрацией
func GetDefects(projectID *uint, status string, priority string, limit, offset int) ([]models.Defect, int64, error) {
	var defects []models.Defect
	var total int64

	query := database.DB.Model(&models.Defect{}).
		Preload("Creator").
		Preload("Project")

	// Фильтр по проекту
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	// Фильтр по статусу
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Фильтр по приоритету
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// Подсчет общего количества
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получение записей с пагинацией (сортировка по дате создания, новые первыми)
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&defects).Error; err != nil {
		return nil, 0, err
	}

	return defects, total, nil
}

// GetDefectByID получает дефект по ID
func GetDefectByID(id uint) (*models.Defect, error) {
	var defect models.Defect
	if err := database.DB.
		Preload("Creator").
		Preload("Project").
		Preload("Attachments").
		First(&defect, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("дефект не найден")
		}
		return nil, err
	}
	return &defect, nil
}

// UpdateDefect обновляет дефект (только для инженеров - создателей)
func UpdateDefect(id uint, updateData models.DefectUpdate, userID uint) (*models.Defect, error) {
	var defect models.Defect
	if err := database.DB.First(&defect, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("дефект не найден")
		}
		return nil, err
	}

	// Проверяем, что пользователь - создатель дефекта
	if defect.CreatedBy != userID {
		return nil, errors.New("только создатель может редактировать дефект")
	}

	// Обновляем только переданные поля
	if updateData.Title != "" {
		defect.Title = updateData.Title
	}
	if updateData.Description != "" {
		defect.Description = updateData.Description
	}
	if updateData.Priority != "" {
		defect.Priority = updateData.Priority
	}

	if err := database.DB.Save(&defect).Error; err != nil {
		return nil, err
	}

	// Загружаем связанные данные
	database.DB.Preload("Creator").Preload("Project").First(&defect, defect.ID)

	return &defect, nil
}

// UpdateDefectStatus изменяет статус дефекта (только для инженеров)
func UpdateDefectStatus(id uint, newStatus string, userID uint) (*models.Defect, error) {
	var defect models.Defect
	if err := database.DB.First(&defect, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("дефект не найден")
		}
		return nil, err
	}

	// Проверяем валидность нового статуса
	if !isValidStatus(newStatus) {
		return nil, errors.New("недопустимый статус")
	}

	// Проверяем возможность смены статуса
	if err := validateStatusTransition(defect.Status, newStatus); err != nil {
		return nil, err
	}

	// Обновляем статус
	defect.Status = newStatus

	if err := database.DB.Save(&defect).Error; err != nil {
		return nil, err
	}

	// Загружаем связанные данные
	database.DB.Preload("Creator").Preload("Project").First(&defect, defect.ID)

	return &defect, nil
}

// DeleteDefect удаляет дефект (только для инженеров - создателей)
func DeleteDefect(id uint, userID uint) error {
	var defect models.Defect
	if err := database.DB.First(&defect, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("дефект не найден")
		}
		return err
	}

	// Проверяем, что пользователь - создатель дефекта
	if defect.CreatedBy != userID {
		return errors.New("только создатель может удалить дефект")
	}

	// Удаляем связанные файлы
	if err := DeleteAttachmentsByEntity(models.EntityTypeDefect, id); err != nil {
		fmt.Printf("Предупреждение при удалении файлов дефекта: %v\n", err)
	}

	// Удаляем дефект (soft delete)
	if err := database.DB.Delete(&defect).Error; err != nil {
		return errors.New("не удалось удалить дефект")
	}

	return nil
}

// GetProjectDefectsStats получает статистику по дефектам проекта
func GetProjectDefectsStats(projectID uint) (map[string]interface{}, error) {
	var totalCount int64
	var registeredCount int64
	var inProgressCount int64
	var completedCount int64

	// Общее количество
	database.DB.Model(&models.Defect{}).Where("project_id = ?", projectID).Count(&totalCount)

	// По статусам
	database.DB.Model(&models.Defect{}).
		Where("project_id = ? AND status = ?", projectID, models.DefectStatusRegistered).
		Count(&registeredCount)

	database.DB.Model(&models.Defect{}).
		Where("project_id = ? AND status = ?", projectID, models.DefectStatusInProgress).
		Count(&inProgressCount)

	database.DB.Model(&models.Defect{}).
		Where("project_id = ? AND status = ?", projectID, models.DefectStatusCompleted).
		Count(&completedCount)

	return map[string]interface{}{
		"total":       totalCount,
		"registered":  registeredCount,
		"in_progress": inProgressCount,
		"completed":   completedCount,
	}, nil
}

// Вспомогательные функции

func isValidStatus(status string) bool {
	validStatuses := []string{
		models.DefectStatusRegistered,
		models.DefectStatusInProgress,
		models.DefectStatusCompleted,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

func validateStatusTransition(currentStatus, newStatus string) error {
	// Правила перехода статусов
	// registered -> in_progress, completed
	// in_progress -> completed, registered (откат)
	// completed -> (нельзя изменить завершенный)

	if currentStatus == newStatus {
		return errors.New("новый статус совпадает с текущим")
	}

	switch currentStatus {
	case models.DefectStatusRegistered:
		if newStatus != models.DefectStatusInProgress && newStatus != models.DefectStatusCompleted {
			return errors.New("из статуса 'зарегистрирован' можно перейти только в 'в работе' или 'завершен'")
		}
	case models.DefectStatusInProgress:
		if newStatus != models.DefectStatusCompleted && newStatus != models.DefectStatusRegistered {
			return errors.New("из статуса 'в работе' можно перейти только в 'завершен' или вернуться в 'зарегистрирован'")
		}
	case models.DefectStatusCompleted:
		return errors.New("завершенный дефект нельзя изменить")
	}

	return nil
}

// ConvertToDefectResponse преобразует Defect в DefectResponse
func ConvertToDefectResponse(defect *models.Defect) models.DefectResponse {
	response := models.DefectResponse{
		ID:          defect.ID,
		ProjectID:   defect.ProjectID,
		Title:       defect.Title,
		Description: defect.Description,
		Status:      defect.Status,
		Priority:    defect.Priority,
		CreatedAt:   defect.CreatedAt,
		UpdatedAt:   defect.UpdatedAt,
	}

	// Добавляем информацию о проекте
	if defect.Project.ID != 0 {
		response.ProjectName = defect.Project.Name
	}

	// Добавляем информацию о создателе
	if defect.Creator.ID != 0 {
		response.Creator = models.DefectCreatorInfo{
			ID:        defect.Creator.ID,
			FirstName: defect.Creator.FirstName,
			LastName:  defect.Creator.LastName,
			Email:     defect.Creator.Email,
		}
	}

	return response
}