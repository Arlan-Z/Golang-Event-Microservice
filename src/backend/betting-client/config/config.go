package config

import (
	"log"
	"os"
	"strconv"

	// Импортируем пакет godotenv
	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	ExternalAPIbaseURL string // Base URL for the external betting API (e.g., "http://external-betting.com/api/events")
	ListenPort         string // Port for our microservice to listen on for callbacks (e.g., "8080")
	CallbackBaseURL    string // The base URL where our service is reachable for callbacks (e.g., "http://my-service.com")
}

// LoadConfig loads configuration from environment variables or defaults.
// It now attempts to load from a .env file first.
func LoadConfig() *Config {
	// --- Загрузка из .env файла ---
	// Пытаемся загрузить переменные из файла .env в текущей директории
	// godotenv.Load() НЕ перезаписывает переменные окружения, которые УЖЕ установлены в системе.
	// Если нужно, чтобы .env перезаписывал системные, используйте godotenv.Overload()
	err := godotenv.Load()
	if err != nil {
		// Обычно ошибку "файл .env не найден" можно игнорировать,
		// так как мы можем полагаться на системные переменные или значения по умолчанию.
		// Но мы можем залогировать другие возможные ошибки (например, неверный формат файла).
		// Используем os.IsNotExist для проверки, что ошибка именно "файл не найден".
		if !os.IsNotExist(err) {
			log.Printf("WARN: Error loading .env file: %v", err)
		} else {
			log.Printf("INFO: .env file not found, using environment variables or defaults.")
		}
	}
	// --- Конец загрузки из .env ---

	// Дальнейший код остается без изменений:
	// Сначала берем значения по умолчанию
	cfg := &Config{
		ExternalAPIbaseURL: getEnv("EXTERNAL_API_BASE_URL", "https://arlan-api.azurewebsites.net/api/events"),
		ListenPort:         getEnv("LISTEN_PORT", "8080"),
		CallbackBaseURL:    getEnv("CALLBACK_BASE_URL", "http://localhost:8080"),
	}

	log.Printf("Configuration loaded:")
	log.Printf("  External API Base URL: %s", cfg.ExternalAPIbaseURL)
	log.Printf("  Listen Port: %s", cfg.ListenPort)
	log.Printf("  Callback Base URL: %s", cfg.CallbackBaseURL)

	// Валидация
	if _, err := strconv.Atoi(cfg.ListenPort); err != nil {
		log.Fatalf("Invalid LISTEN_PORT: %s", cfg.ListenPort)
	}
	if cfg.ExternalAPIbaseURL == "" {
		log.Fatalf("EXTERNAL_API_BASE_URL cannot be empty")
	}
	if cfg.CallbackBaseURL == "" {
		log.Fatalf("CALLBACK_BASE_URL cannot be empty")
	}

	return cfg
}

// getEnv читает переменную окружения или возвращает значение по умолчанию.
// Теперь она будет видеть переменные, загруженные из .env (если они не были переопределены системными)
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("INFO: Using value for %s from environment.", key) // Добавим лог
		return value
	}
	log.Printf("INFO: Environment variable %s not set, using default: %s", key, fallback)
	return fallback
}
