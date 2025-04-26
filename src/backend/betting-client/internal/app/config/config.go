package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPServer struct {
		Port    string        `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
		Timeout time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"5s"`
	} `yaml:"http_server"`
	Database struct {
		Path string `yaml:"path" env:"DB_PATH" env-required:"true"`
	} `yaml:"database"`
	PayoutService struct {
		URL     string        `yaml:"url" env:"PAYOUT_SVC_URL" env-required:"true"`
		Timeout time.Duration `yaml:"timeout" env:"PAYOUT_SVC_TIMEOUT" env-default:"3s"`
	} `yaml:"payout_service"`
	EventSourceAPI struct { // <-- Новый раздел
		URL          string        `yaml:"url" env:"EVENT_SOURCE_URL" env-required:"true"`
		Timeout      time.Duration `yaml:"timeout" env:"EVENT_SOURCE_TIMEOUT" env-default:"10s"`
		SyncInterval time.Duration `yaml:"sync_interval" env:"EVENT_SYNC_INTERVAL" env-default:"5m"`
	} `yaml:"event_source_api"`
}

func Load() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml" // Default path
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
