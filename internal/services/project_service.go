package services

import (
	"errors"

	"SystemContorlBackend/internal/database"
	"SystemContorlBackend/internal/models"
	"gorm.io/gorm"
)

// CreateProject создает новый проект
func CreateProject(projectData models.ProjectCreate, createdBy uint) (*models.Project, error) {
	project := models.Project{
		Name:        projectData.Name,
		Description: projectData.Description,
		Address:     projectData.Address,
		Status:      models.ProjectStatusActive,
		StartDate:   projectData.StartDate,
		EndDate:     projectData.EndDate,
		CreatedBy:   createdBy,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		return nil, err
	}

	// Загружаем информацию о создателе
	database.DB.Preload("Creator").First(&project, project.ID)

	return &project, nil
}

// GetProjects получает список проектов с фильтрацией
func GetProjects(status string, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := database.DB.Model(&models.Project{}).Preload("Creator")

	// Фильтр по статусу
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Подсчет общего количества
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получение записей с пагинацией
	if err := query.Offset(offset).Limit(limit).Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// GetProjectByID получает проект по ID
func GetProjectByID(id uint) (*models.Project, error) {
	var project models.Project
	if err := database.DB.Preload("Creator").First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("проект не найден")
		}
		return nil, err
	}
	return &project, nil
}

// UpdateProject обновляет проект
func UpdateProject(id uint, updateData models.ProjectUpdate) (*models.Project, error) {
	var project models.Project
	if err := database.DB.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("проект не найден")
		}
		return nil, err
	}

	// Обновляем только переданные поля
	if updateData.Name != "" {
		project.Name = updateData.Name
	}
	if updateData.Description != "" {
		project.Description = updateData.Description
	}
	if updateData.Address != "" {
		project.Address = updateData.Address
	}
	if updateData.Status != "" {
		project.Status = updateData.Status
	}
	if updateData.StartDate != nil {
		project.StartDate = updateData.StartDate
	}
	if updateData.EndDate != nil {
		project.EndDate = updateData.EndDate
	}

	if err := database.DB.Save(&project).Error; err != nil {
		return nil, err
	}

	// Загружаем информацию о создателе
	database.DB.Preload("Creator").First(&project, project.ID)

	return &project, nil
}

// DeleteProject удаляет проект (мягкое удаление)
func DeleteProject(id uint) error {
	var project models.Project
	if err := database.DB.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("проект не найден")
		}
		return err
	}

	return database.DB.Delete(&project).Error
}