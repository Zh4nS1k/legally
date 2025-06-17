package services

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"legally/db"
	"legally/models"
	"legally/utils"
	"time"
)

var (
	ErrUserExists         = errors.New("пользователь уже существует")
	ErrInvalidCredentials = errors.New("неверные учетные данные")
	ErrUserNotFound       = errors.New("пользователь не найден")
	ErrTokenGeneration    = errors.New("ошибка генерации токена")
)

// Register регистрирует нового пользователя и возвращает пару токенов
func Register(email, password string, role models.UserRole) (map[string]string, error) {
	// Проверяем существование пользователя
	var existingUser models.User
	err := db.GetCollection("users").FindOne(
		context.Background(),
		bson.M{"email": email},
	).Decode(&existingUser)

	if err == nil {
		return nil, ErrUserExists
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем нового пользователя
	user := models.User{
		ID:        primitive.NewObjectID(),
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: time.Now(),
	}

	// Сохраняем в БД
	_, err = db.GetCollection("users").InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}

	// Генерируем токены
	accessToken, refreshToken, err := utils.GenerateTokenPair(user.ID.Hex(), user.Role)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	return map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}, nil
}

// Login аутентифицирует пользователя и возвращает пару токенов
func Login(email, password string) (map[string]string, error) {
	var user models.User
	err := db.GetCollection("users").FindOne(
		context.Background(),
		bson.M{"email": email},
	).Decode(&user)

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Генерируем токены
	accessToken, refreshToken, err := utils.GenerateTokenPair(user.ID.Hex(), user.Role)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	return map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}, nil
}

// RefreshTokens обновляет пару токенов
func RefreshTokens(refreshToken string) (map[string]string, error) {
	claims, err := utils.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Проверяем существование пользователя
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	var user models.User
	err = db.GetCollection("users").FindOne(
		context.Background(),
		bson.M{"_id": userID},
	).Decode(&user)

	if err != nil {
		return nil, ErrUserNotFound
	}

	// Генерируем новые токены
	accessToken, refreshToken, err := utils.GenerateTokenPair(user.ID.Hex(), user.Role)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	return map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}, nil
}

// ValidateUser проверяет существование пользователя по ID
func ValidateUser(userID string) (*models.User, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = db.GetCollection("users").FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&user)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return &user, nil
}
