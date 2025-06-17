package controllers

import (
	"github.com/gin-gonic/gin"
	"legally/models"
	"legally/services"
	"net/http"
)

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := services.Register(req.Email, req.Password, models.RoleUser)
	if err != nil {
		status := http.StatusBadRequest
		if err == services.ErrUserExists {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Регистрация прошла успешно",
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
	})
}

func Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := services.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Вход выполнен успешно",
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
		"success":      true,
	})
}

func Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := services.RefreshTokens(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
		"success":      true,
	})
}

func ValidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"valid": false})
		return
	}

	// Проверка токена происходит в middleware
	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Токен действителен",
	})
}

func Logout(c *gin.Context) {
	// В реальном приложении здесь можно добавить токен в blacklist
	c.JSON(http.StatusOK, gin.H{
		"message": "Выход выполнен успешно",
		"success": true,
	})
}
