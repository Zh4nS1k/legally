package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"legally/api"
	"legally/db"
	"log"
	"os"
)

func main() {
	_ = godotenv.Load()
	db.InitMongo()

	if err := os.MkdirAll("./temp", os.ModePerm); err != nil {
		log.Fatal("❌ ERROR: Не удалось создать временную папку:", err)
	}

	router := gin.Default()
	api.SetupRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("✅ SUCCESS: Сервер запущен на http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("❌ ERROR: Ошибка при запуске сервера:", err)
	}
}
