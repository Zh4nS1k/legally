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

// @Summary Регистрация
// @Description Регистрация нового пользователя
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param input body AuthRequest true "Данные для регистрации"
// @Success 200 {object} gin.H "Сообщение об успехе"
// @Router /api/register [post]
func Register(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := services.Register(req.Email, req.Password, models.RoleUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Регистрация прошла успешно",
		"success": true,
	})
}

// @Summary Логин
// @Description Аутентификация пользователя
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param input body AuthRequest true "Данные для входа"
// @Success 200 {object} gin.H "Токен и сообщение"
// @Router /api/login [post]
func Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := services.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Неверные учетные данные",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "Вход выполнен успешно",
		"success": true,
	})
}
