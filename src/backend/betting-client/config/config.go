package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	ExternalAPIbaseURL string // Base URL for the external betting API (e.g., "http://external-betting.com/api/events")
	ListenPort         string // Port for our microservice to listen on for callbacks (e.g., "8080")
	CallbackBaseURL    string // The base URL where our service is reachable for callbacks (e.g., "http://my-service.com")
}

// LoadConfig loads configuration from environment variables or defaults.
func LoadConfig() *Config {
	cfg := &Config{
		ExternalAPIbaseURL: getEnv("EXTERNAL_API_BASE_URL", "https://arlan-api.azurewebsites.net/api/events"), // Default to a mock server URL
		ListenPort:         getEnv("LISTEN_PORT", "8080"),
		CallbackBaseURL:    getEnv("CALLBACK_BASE_URL", "http://localhost:8080"),
	}

	log.Printf("Configuration loaded:")
	log.Printf("  External API Base URL: %s", cfg.ExternalAPIbaseURL)
	log.Printf("  Listen Port: %s", cfg.ListenPort)
	log.Printf("  Callback Base URL: %s", cfg.CallbackBaseURL)

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

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Environment variable %s not set, using default: %s", key, fallback)
	return fallback
}
