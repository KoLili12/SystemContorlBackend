package services

import (
	"errors"

	"SystemContorlBackend/internal/database"
	"SystemContorlBackend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterUser регистрирует нового пользователя
func RegisterUser(userData models.UserRegister) (*models.User, error) {
	// Проверяем, существует ли пользователь с таким email
	var existingUser models.User
	if err := database.DB.Where("email = ?", userData.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	// Хешируем пароль с использованием bcrypt согласно ТЗ
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Находим роль по коду
	var role models.Role
	if err := database.DB.Where("code = ?", userData.RoleCode).First(&role).Error; err != nil {
		return nil, errors.New("указанная роль не найдена")
	}

	// Создаем пользователя
	user := models.User{
		Email:     userData.Email,
		Password:  string(hashedPassword),
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Phone:     userData.Phone,
		RoleID:    role.ID,
		IsActive:  true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	// Загружаем роль для ответа
	database.DB.Preload("Role").First(&user, user.ID)

	return &user, nil
}

// LoginUser выполняет вход пользователя
func LoginUser(loginData models.UserLogin) (*models.User, string, error) {
	var user models.User
	
	// Находим пользователя по email с подгрузкой роли
	if err := database.DB.Preload("Role").Where("email = ?", loginData.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", errors.New("неверный email или пароль")
		}
		return nil, "", err
	}

	// Проверяем активен ли пользователь
	if !user.IsActive {
		return nil, "", errors.New("пользователь заблокирован")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		return nil, "", errors.New("неверный email или пароль")
	}

	// Генерируем JWT токен
	token, err := GenerateToken(&user)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

// GetUserByID получает пользователя по ID
func GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := database.DB.Preload("Role").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}