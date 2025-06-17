package api

import (
	"github.com/gin-gonic/gin"
	"legally/api/controllers"
	"legally/api/middleware"
	"legally/db"
	"legally/models"
)

func SetupRoutes(router *gin.Engine) {
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())

	router.Static("/static", "./public")
	router.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})
	public := router.Group("/api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.GET("/laws", controllers.GetRelevantLaws)
	}

	// Приватные маршруты для пользователей
	private := router.Group("/api")
	private.Use(middleware.AuthRequired(models.RoleUser))
	{
		private.POST("/analyze", controllers.AnalyzeDocument)
		private.GET("/history", controllers.GetHistory)
	}

	// Админские маршруты
	admin := router.Group("/api/admin")
	admin.Use(middleware.AuthRequired(models.RoleAdmin))
	{
		// Здесь можно добавить админские эндпоинты
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(200, gin.H{"status": "healthy"})
	})

}
