package services

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"legally/db"
	"legally/models"
	"legally/utils"
)

func Register(email, password string, role models.UserRole) (bool, error) {
	// Проверяем, существует ли уже пользователь
	var existingUser models.User
	err := db.GetCollection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&existingUser)
	if err == nil {
		return false, models.ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	user := models.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	_, err = db.GetCollection("users").InsertOne(context.Background(), user)
	if err != nil {
		return false, err
	}

	return true, nil
}

func Login(email, password string) (string, error) {
	var user models.User
	err := db.GetCollection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return "", models.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", models.ErrInvalidCredentials
	}

	return utils.GenerateToken(user.ID.Hex(), user.Role)
}
