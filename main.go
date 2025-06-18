// main.go

package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"legally/api"
	"legally/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	_ = godotenv.Load()
	checkEnvVars()
	db.InitMongo()

	if err := os.MkdirAll("./temp", os.ModePerm); err != nil {
		log.Fatal("❌ ERROR: Не удалось создать временную папку:", err)
	}

	router := gin.Default()
	api.SetupRoutes(router)
	
	router.Use(func(c *gin.Context) {
		log.Printf("Incoming request: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("✅ Сервер запущен на http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Ошибка сервера: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🔄 Завершение работы сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Принудительное завершение работы:", err)
	}

	log.Println("✅ Сервер успешно остановлен")
}

func checkEnvVars() {
	required := []string{"MONGO_URI", "OPENROUTER_API_KEY"}
	for _, env := range required {
		if os.Getenv(env) == "" {
			log.Fatalf("❌ ERROR: Необходимо установить переменную окружения %s", env)
		}
	}
}
