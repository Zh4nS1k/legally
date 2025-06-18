package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"legally/models"
	"legally/utils"
	"net/http"
	"strings"
)

func AuthRequired(requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Требуется авторизация",
				"code":  "MISSING_AUTH_HEADER",
			})
			return
		}

		// Проверяем формат "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Неверный формат токена",
				"code":  "INVALID_TOKEN_FORMAT",
			})
			return
		}

		token := tokenParts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			utils.LogError(fmt.Sprintf("Ошибка токена: %v", err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Неверный или истекший токен",
				"code":    "INVALID_OR_EXPIRED_TOKEN",
				"details": err.Error(),
			})
			return
		}

		// Проверка роли
		if requiredRole != "" && claims.Role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Недостаточно прав",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("userId", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}
