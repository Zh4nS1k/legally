// analysis_controller.go

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
			"error":  "File is required",
			"code":   "FILE_REQUIRED",
			"detail": err.Error(),
		})
		return
	}

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File must be PDF format",
			"code":  "INVALID_FILE_TYPE",
		})
		return
	}

	// Get user from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
			"code":  "AUTH_ERROR",
		})
		return
	}

	// Create a copy of the context with userId
	newC := c.Copy()
	newC.Set("userId", userID)

	// Process the file using AnalyzePDFDocument
	result, serviceErr := services.AnalyzeDocument(newC)
	if serviceErr != nil {
		c.JSON(serviceErr.Status, gin.H{
			"error": serviceErr.Message,
			"code":  "ANALYSIS_ERROR",
		})
		return
	}

	// Type assert the result to access its fields
	analysisResult, ok := result.(gin.H)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "RESULT_TYPE_ERROR",
		})
		return
	}

	// Return successful response with data
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"analysis":     analysisResult["analysis"],
		"documentType": analysisResult["document_type"],
		"filename":     analysisResult["filename"],
		"timestamp":    analysisResult["timestamp"],
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
