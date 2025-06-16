package api

import (
	"github.com/gin-gonic/gin"
	"legally/api/controllers"
	"legally/api/middleware"
	"legally/db"
)

func SetupRoutes(router *gin.Engine) {
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

	router.Static("/static", "./public")
	router.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(200, gin.H{"status": "healthy"})
	})

	api := router.Group("/api")
	{
		api.POST("/analyze", controllers.AnalyzeDocument)
		api.GET("/laws", controllers.GetRelevantLaws)
		api.GET("/history", controllers.GetHistory)
	}
}
