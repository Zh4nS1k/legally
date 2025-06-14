package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ledongthuc/pdf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"legally/db"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	model       = "deepseek/deepseek-r1-0528:free"
	apiEndpoint = "https://openrouter.ai/api/v1/chat/completions"
)

func main() {
	_ = godotenv.Load()
	db.InitMongo()
	// Создаем временную папку, если ее нет
	if err := os.MkdirAll("./temp", os.ModePerm); err != nil {
		log.Fatal("Не удалось создать временную папку:", err)
	}

	router := gin.Default()
	router.Static("/static", "./public")
	router.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})
	router.POST("/api/analyze", analyzeDocumentHandler)
	router.GET("/api/laws", getRelevantLawsHandler)
	router.GET("/api/history", getHistoryHandler)

	// Добавляем middleware для логирования запросов
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		log.Printf("%s %s %d %v", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
	})

	log.Println("✅ SUCCESS: Сервер запущен на http://localhost:3000")
	if err := router.Run(":3000"); err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
}

func analyzeDocumentHandler(c *gin.Context) {
	log.Println("🔄 Получен запрос на анализ документа")

	file, err := c.FormFile("document")
	if err != nil {
		log.Println("❌ ERROR: Не удалось получить файл:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не получен"})
		return
	}

	tempPath := fmt.Sprintf("./temp/%d_%s", time.Now().Unix(), file.Filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		log.Println("❌ ERROR: Не удалось сохранить файл:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла"})
		return
	}
	defer func() {
		if err := os.Remove(tempPath); err != nil {
			log.Println("❌ WARNING: Не удалось удалить временный файл:", err)
		}
	}()

	text, err := extractTextFromDocument(tempPath)
	if err != nil {
		log.Println("❌ ERROR: Ошибка извлечения текста:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("ℹ️ Извлечено %d символов из документа", len(text))

	parts := splitTextByChars(text, 12000)
	log.Printf("ℹ️ Документ разбит на %d частей для анализа", len(parts))

	var analysisResults []string
	for i, part := range parts {
		log.Printf("🔄 Анализ части %d/%d...", i+1, len(parts))
		result, err := analyzeDocumentPart(part)
		if err != nil {
			log.Println("❌ ERROR при анализе части:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.Printf("✅ Анализ части %d завершён, результат длиной %d символов", i+1, len(result))
		analysisResults = append(analysisResults, result)
	}

	fullAnalysis := strings.Join(analysisResults, "\n\n---\n\n")
	docType := detectDocumentType(text)

	log.Println("✅ Полный анализ готов, отправляем ответ клиенту")
	log.Printf("ℹ️ Тип документа: %s, длина анализа: %d символов", docType, len(fullAnalysis))

	_, err = db.GetCollection("analyses").InsertOne(context.TODO(), bson.M{
		"filename":   file.Filename,
		"type":       docType,
		"analysis":   fullAnalysis,
		"text":       text,
		"created_at": time.Now(),
	})
	if err != nil {
		log.Println("❌ Ошибка сохранения в Mongo:", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis":      fullAnalysis,
		"timestamp":     time.Now().Format(time.RFC3339),
		"document_type": docType,
	})
}

func extractTextFromDocument(path string) (string, error) {
	log.Printf("ℹ️ Извлечение текста из файла: %s", path)

	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("не удалось открыть документ: %v", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("ошибка чтения документа: %v", err)
	}
	if _, err := io.Copy(&buf, b); err != nil {
		return "", fmt.Errorf("ошибка копирования текста: %v", err)
	}

	text := buf.String()
	if len(text) == 0 {
		return "", fmt.Errorf("файл пуст")
	}

	// Очистка лишних пробелов и переносов строк
	text = strings.Join(strings.Fields(text), " ")
	return text, nil
}

func analyzeDocumentPart(text string) (string, error) {
	prompt := fmt.Sprintf(`Проанализируй следующий юридический документ на соответствие законодательству Казахстана. 
Выяви потенциальные риски, несоответствия и проблемные формулировки. 
Сгруппируй результаты по категориям:
1. Правовые риски
2. Неясные формулировки
3. Возможные нарушения
4. Рекомендации

Для каждой проблемы укажи:
- Описание
- Закон/статью
- Уровень риска (высокий, средний, низкий)
- Рекомендации по исправлению

Документ:
%s`, text)

	log.Printf("ℹ️ Отправка запроса к AI с текстом длиной %d символов", len(text))
	return queryOpenRouter(prompt)
}

func queryOpenRouter(prompt string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY не установлен")
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "Ты — юридический эксперт по законодательству Казахстана. Анализируй документы и давай развернутые ответы с конкретными ссылками на законы."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  4000,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ошибка маршалинга payload: %w", err)
	}

	log.Printf("➡️ Отправляем запрос к OpenRouter, длина тела: %d байт", len(body))

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://legally.kz")
	req.Header.Set("X-Title", "Legally AI Risk Analyzer")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса к OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа OpenRouter: %w", err)
	}

	log.Printf("⬅️ Получен ответ от OpenRouter, статус: %d, длина тела: %d байт", resp.StatusCode, len(resBody))

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Тело ответа с ошибкой: %s", string(resBody))
		return "", fmt.Errorf("ошибка от OpenRouter: статус %d", resp.StatusCode)
	}

	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(resBody, &res); err != nil {
		return "", fmt.Errorf("не удалось распарсить ответ AI: %w", err)
	}

	if res.Error.Message != "" {
		return "", fmt.Errorf("ошибка от API: %s", res.Error.Message)
	}

	if len(res.Choices) == 0 || res.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("пустой ответ от OpenRouter")
	}

	log.Printf("✅ Успешно получен ответ от AI длиной %d символов", len(res.Choices[0].Message.Content))
	return res.Choices[0].Message.Content, nil
}

func splitTextByChars(text string, maxChars int) []string {
	var parts []string
	runes := []rune(text)
	for start := 0; start < len(runes); start += maxChars {
		end := start + maxChars
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[start:end]))
	}
	return parts
}

func detectDocumentType(text string) string {
	text = strings.ToLower(text)
	switch {
	case strings.Contains(text, "договор"):
		return "Договор"
	case strings.Contains(text, "приказ"):
		return "Приказ"
	case strings.Contains(text, "постановление"):
		return "Постановление"
	case strings.Contains(text, "закон"):
		return "Закон"
	case strings.Contains(text, "решение"):
		return "Решение"
	default:
		return "Неизвестно"
	}
}

func getRelevantLawsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"laws": []map[string]string{
			{"name": "Гражданский кодекс РК", "url": "https://adilet.zan.kz/rus/docs/K950001000_"},
			{"name": "Налоговый кодекс РК", "url": "https://adilet.zan.kz/rus/docs/K2100000409"},
			{"name": "Трудовой кодекс РК", "url": "https://adilet.zan.kz/rus/docs/K1500000011"},
			{"name": "Кодекс об административных правонарушениях РК", "url": "https://adilet.zan.kz/rus/docs/K1400000233"},
		},
	})
}
func getHistoryHandler(c *gin.Context) {
	coll := db.GetCollecti  on("analyses")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{"created_at", -1}}))
	if err != nil {
		log.Println("❌ Ошибка чтения из Mongo:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения истории"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных"})
		return
	}

	c.JSON(http.StatusOK, results)
}
