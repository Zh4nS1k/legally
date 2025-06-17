package models

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

var (
	ErrUserExists         = errors.New("пользователь с таким email уже существует")
	ErrInvalidCredentials = errors.New("неверные учетные данные")
)

const (
	RoleAdmin     UserRole = "admin"
	RoleUser      UserRole = "user"
	RoleAnonymous UserRole = "anonymous"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"-"`
	Role     UserRole           `bson:"role" json:"role"`
}
