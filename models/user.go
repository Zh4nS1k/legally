// user.go

package models

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
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
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	Role      UserRole           `bson:"role"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}
