package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"

	"search_service/pkg/logging"
)

type Config struct {
	IsDebug   *bool `yaml:"is_debug"`
	Typesense struct {
		URL        string `yaml:"url" env:"SEARCH_SERVICE_TYPESENSE_URL" env-default:"http://localhost:8108"`
		APIKey     string `yaml:"api_key" env:"SEARCH_SERVICE_TYPESENSE_API_KEY" env-default:"local-typesense-api-key"`
		Collection string `yaml:"collection" env:"SEARCH_SERVICE_TYPESENSE_COLLECTION" env-default:"notes"`
	} `yaml:"typesense"`
	RabbitMQ struct {
		Enabled  bool   `yaml:"enabled" env:"SEARCH_SERVICE_RABBITMQ_ENABLED" env-default:"false"`
		URL      string `yaml:"url" env:"SEARCH_SERVICE_RABBITMQ_URL" env-default:"amqp://guest:guest@localhost:5672/"`
		Exchange string `yaml:"exchange" env:"SEARCH_SERVICE_RABBITMQ_EXCHANGE" env-default:"notes.events"`
	} `yaml:"rabbitmq"`
	NoteService struct {
		URL string `yaml:"url" env:"SEARCH_SERVICE_NOTE_URL" env-default:"http://localhost:8082/api"`
	} `yaml:"note_service"`
	CategoryService struct {
		URL string `yaml:"url" env:"SEARCH_SERVICE_CATEGORY_URL" env-default:"http://localhost:8081/api"`
	} `yaml:"category_service"`
	FileService struct {
		URL string `yaml:"url" env:"SEARCH_SERVICE_FILE_URL" env-default:"http://localhost:8085"`
	} `yaml:"file_service"`
	Listen struct {
		Type   string `yaml:"type" env:"SEARCH_SERVICE_LISTEN_TYPE" env-default:"port"`
		BindIP string `yaml:"bind_ip" env:"SEARCH_SERVICE_BIND_IP" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env:"SEARCH_SERVICE_PORT" env-default:"8086"`
	} `yaml:"listen"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		instance = &Config{}

		if err := cleanenv.ReadConfig(configPath(), instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info("config help", "help", help)
			logger.Fatal("failed to read config", "error", err)
		}

		logger.Info(
			"config loaded",
			"address", instance.Listen.BindIP+":"+instance.Listen.Port,
			"typesense_url", instance.Typesense.URL,
			"collection", instance.Typesense.Collection,
			"rabbitmq_enabled", instance.RabbitMQ.Enabled,
		)
	})

	return instance
}

func configPath() string {
	candidates := []string{
		"config.yml",
		filepath.Join("..", "config.yml"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return "config.yml"
}
