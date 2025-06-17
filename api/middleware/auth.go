package middleware

import (
	"github.com/gin-gonic/gin"
	"legally/models"
	"legally/utils"
	"net/http"
)

func AuthRequired(requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			return
		}

		if requiredRole != "" && claims.Role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}
