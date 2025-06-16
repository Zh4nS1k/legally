package utils

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxFileSize    = 10 << 20 // 10MB
	tempFilePrefix = "temp_"
)

// ProcessUploadedFile обрабатывает загруженный файл и извлекает текст
func ProcessUploadedFile(c *gin.Context) (string, string, error) {
	LogAction("Начало обработки загруженного файла")

	// Проверка размера файла
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)
	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		LogError(fmt.Sprintf("Превышен максимальный размер файла (10MB): %v", err))
		return "", "", fmt.Errorf("размер файла не должен превышать 10MB")
	}

	// Получение файла из запроса
	file, header, err := c.Request.FormFile("document")
	if err != nil {
		LogError(fmt.Sprintf("Ошибка получения файла: %v", err))
		return "", "", fmt.Errorf("файл не получен")
	}
	defer file.Close()

	// Проверка расширения файла
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".pdf" {
		LogError(fmt.Sprintf("Неподдерживаемый формат файла: %s", ext))
		return "", "", fmt.Errorf("поддерживаются только PDF файлы")
	}

	// Создание временного файла
	tempPath := filepath.Join("./temp", tempFilePrefix+fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename))
	LogInfo(fmt.Sprintf("Создание временного файла: %s", tempPath))

	tempFile, err := os.Create(tempPath)
	if err != nil {
		LogError(fmt.Sprintf("Ошибка создания временного файла: %v", err))
		return "", "", fmt.Errorf("ошибка создания временного файла")
	}
	defer tempFile.Close()

	// Копирование содержимого во временный файл
	if _, err := io.Copy(tempFile, file); err != nil {
		LogError(fmt.Sprintf("Ошибка сохранения файла: %v", err))
		return "", "", fmt.Errorf("ошибка сохранения файла")
	}

	// Удаление временного файла после завершения
	defer func() {
		if err := os.Remove(tempPath); err != nil {
			LogWarning(fmt.Sprintf("Не удалось удалить временный файл: %v", err))
		} else {
			LogInfo(fmt.Sprintf("Временный файл удален: %s", tempPath))
		}
	}()

	// Извлечение текста из PDF
	LogInfo(fmt.Sprintf("Извлечение текста из файла: %s", tempPath))
	text, err := ExtractTextFromPDF(tempPath)
	if err != nil {
		LogError(fmt.Sprintf("Ошибка извлечения текста: %v", err))
		return "", "", fmt.Errorf("ошибка извлечения текста: %v", err)
	}

	if len(text) == 0 {
		LogWarning("Документ не содержит текста")
		return "", "", fmt.Errorf("документ не содержит текста")
	}

	LogSuccess(fmt.Sprintf("Успешно обработан файл: %s (символов: %d)", header.Filename, len(text)))
	return text, header.Filename, nil
}

// ExtractTextFromPDF извлекает текст из PDF файла
func ExtractTextFromPDF(path string) (string, error) {
	LogAction(fmt.Sprintf("Извлечение текста из PDF: %s", path))

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

	// Нормализация пробелов
	text = strings.Join(strings.Fields(text), " ")
	LogInfo(fmt.Sprintf("Извлечено %d символов из PDF", len(text)))

	return text, nil
}

// SplitText разделяет текст на части по максимальному количеству символов
func SplitText(text string, maxChars int) []string {
	LogAction(fmt.Sprintf("Разделение текста (макс. %d символов на часть)", maxChars))

	runes := []rune(text)
	var parts []string

	for start := 0; start < len(runes); start += maxChars {
		end := start + maxChars
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[start:end]))
	}

	LogInfo(fmt.Sprintf("Текст разделен на %d частей", len(parts)))
	return parts
}
