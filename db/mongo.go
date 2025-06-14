package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func InitMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Ошибка подключения к MongoDB:", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ MongoDB недоступна:", err)
	}

	MongoClient = client
	log.Println("✅ MongoDB подключена")
}

func GetCollection(name string) *mongo.Collection {
	return MongoClient.Database("legally").Collection(name)
}
