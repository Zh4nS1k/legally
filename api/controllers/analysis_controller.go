package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"legally/services"
)

func AnalyzeDocument(c *gin.Context) {
	result, err := services.AnalyzeDocument(c)
	if err != nil {
		c.JSON(err.Status, gin.H{"error": err.Message})
		return
	}
	c.JSON(http.StatusOK, result)
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
