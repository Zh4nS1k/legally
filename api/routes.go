package api

import (
	"github.com/gin-gonic/gin"
	"legally/api/controllers"
	"legally/api/middleware"
)

func SetupRoutes(router *gin.Engine) {
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

	router.Static("/static", "./public")
	router.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})

	api := router.Group("/api")
	{
		api.POST("/analyze", controllers.AnalyzeDocument)
		api.GET("/laws", controllers.GetRelevantLaws)
		api.GET("/history", controllers.GetHistory)
	}
}
