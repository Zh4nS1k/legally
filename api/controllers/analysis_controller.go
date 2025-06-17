package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"legally/services"
)

func AnalyzeDocument(c *gin.Context) {
	// Get file from request
	file, err := c.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Необходимо загрузить файл",
			"code":  "FILE_REQUIRED",
		})
		return
	}

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Файл должен быть в формате PDF",
			"code":  "INVALID_FILE_TYPE",
		})
		return
	}

	// Get user from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{ // 401 is more appropriate
			"error": "Ошибка аутентификации",
			"code":  "AUTH_ERROR",
		})
		return
	}

	// Delegate to service layer
	result, serviceErr := services.AnalyzePDFDocument(file, userID.(string))
	if serviceErr != nil {
		c.JSON(serviceErr.Status, gin.H{
			"error": serviceErr.Message,
			"code":  serviceErr.Code,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"analysis":   result.Analysis,
		"documentId": result.DocumentID,
	})
}

func GetRelevantLaws(c *gin.Context) {
	laws := services.GetRelevantLaws()
	c.JSON(http.StatusOK, gin.H{"laws": laws})
}

func GetHistory(c *gin.Context) {
	history, err := services.GetHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения истории"})
		return
	}
	c.JSON(http.StatusOK, history)
}
