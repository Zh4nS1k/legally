// analysis_repository.go

package repositories

import (
	"context"
	"fmt"
	"legally/db"
	"legally/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveAnalysis(filename, docType, analysis, text string) error {
	utils.LogAction("Сохранение анализа в БД")

	_, err := db.GetCollection("analyses").InsertOne(context.TODO(), bson.M{
		"filename":   filename,
		"type":       docType,
		"analysis":   analysis,
		"text":       text,
		"created_at": time.Now(),
	})

	if err != nil {
		utils.LogError(fmt.Sprintf("Ошибка сохранения анализа: %v", err))
	} else {
		utils.LogSuccess("Анализ успешно сохранён в БД")
	}

	return err
}

func GetHistory() ([]map[string]interface{}, error) {
	utils.LogAction("Получение истории анализов")

	coll := db.GetCollection("analyses")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(50)

	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		utils.LogError(fmt.Sprintf("Ошибка получения истории: %v", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		utils.LogError(fmt.Sprintf("Ошибка декодирования истории: %v", err))
		return nil, err
	}

	for _, result := range results {
		delete(result, "_id")
	}

	utils.LogSuccess(fmt.Sprintf("Получено %d записей истории", len(results)))
	return results, nil
}
