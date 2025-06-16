package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"legally/repositories"
	"legally/utils"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	model       = "deepseek/deepseek-r1-0528:free"
	apiEndpoint = "https://openrouter.ai/api/v1/chat/completions"
	maxFileSize = 10 << 20 // 10MB
)

type HttpError struct {
	Status  int
	Message string
}

func AnalyzeDocument(c *gin.Context) (interface{}, *HttpError) {
	utils.LogAction("Получен запрос на анализ документа")

	text, filename, err := utils.ProcessUploadedFile(c)
	if err != nil {
		utils.LogError(err.Error())
		return nil, &HttpError{Status: http.StatusBadRequest, Message: err.Error()}
	}

	utils.LogInfo(fmt.Sprintf("Извлечено %d символов из документа", len(text)))

	analysis, docType, err := AnalyzeText(text)
	if err != nil {
		utils.LogError(err.Error())
		return nil, &HttpError{Status: http.StatusInternalServerError, Message: err.Error()}
	}

	err = repositories.SaveAnalysis(filename, docType, analysis, text)
	if err != nil {
		utils.LogWarning(fmt.Sprintf("Ошибка сохранения в MongoDB: %v", err))
	}

	utils.LogSuccess("Полный анализ готов, отправляем ответ клиенту")
	utils.LogInfo(fmt.Sprintf("Тип документа: %s, длина анализа: %d символов", docType, len(analysis)))

	return gin.H{
		"analysis":      analysis,
		"timestamp":     time.Now().Format(time.RFC3339),
		"document_type": docType,
		"filename":      filename,
	}, nil
}

func AnalyzeText(text string) (string, string, error) {
	parts := utils.SplitText(text, 12000)
	utils.LogInfo(fmt.Sprintf("Документ разбит на %d частей для анализа", len(parts)))

	var analysisResults []string
	for i, part := range parts {
		partNum := i + 1
		utils.LogAction(fmt.Sprintf("Анализ части %d/%d...", partNum, len(parts)))

		result, err := analyzeDocumentPart(part)
		if err != nil {
			utils.LogError(fmt.Sprintf("При анализе части %d: %v", partNum, err))
			return "", "", err
		}

		utils.LogSuccess(fmt.Sprintf("Анализ части %d завершён, результат длиной %d символов", partNum, len(result)))
		analysisResults = append(analysisResults, result)
	}

	fullAnalysis := strings.Join(analysisResults, "\n\n---\n\n")
	docType := detectDocumentType(text)

	return fullAnalysis, docType, nil
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

	utils.LogInfo(fmt.Sprintf("Отправка запроса к AI с текстом длиной %d символов", len(text)))

	result, err := queryOpenRouter(prompt)
	if err != nil {
		return "", err
	}

	utils.LogSuccess(fmt.Sprintf("Успешно получен ответ от AI длиной %d символов", len(result)))
	return result, nil
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

	utils.LogRequest("out", apiEndpoint, len(body))

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

	utils.LogRequest("in", fmt.Sprintf("OpenRouter (статус: %d)", resp.StatusCode), len(resBody))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка от OpenRouter: статус %d", resp.StatusCode)
	}

	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(resBody, &res); err != nil {
		return "", fmt.Errorf("не удалось распарсить ответ AI: %w", err)
	}

	if len(res.Choices) == 0 || res.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("пустой ответ от OpenRouter")
	}

	return res.Choices[0].Message.Content, nil
}

func GetRelevantLaws() []map[string]string {
	return []map[string]string{
		{"name": "Гражданский кодекс РК", "url": "https://adilet.zan.kz/rus/docs/K950001000_"},
		{"name": "Налоговый кодекс РК", "url": "https://adilet.zan.kz/rus/docs/K2100000409"},
		{"name": "Трудовой кодекс РК", "url": "https://adilet.zan.kz/rus/docs/K1500000011"},
		{"name": "Кодекс об административных правонарушениях РК", "url": "https://adilet.zan.kz/rus/docs/K1400000233"},
	}
}

func GetHistory() ([]map[string]interface{}, error) {
	return repositories.GetHistory()
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
