package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"

	"file_service/pkg/logging"
)

type Config struct {
	IsDebug bool `yaml:"is_debug" env:"FILE_SERVICE_DEBUG" env-default:"true"`
	Minio   struct {
		Endpoint  string `yaml:"endpoint" env:"FILE_SERVICE_MINIO_ENDPOINT" env-default:"127.0.0.1:9000"`
		AccessKey string `env:"MINIO_ROOT_USER"`
		SecretKey string `env:"MINIO_ROOT_PASSWORD"`
		UseSSL    bool   `yaml:"use_ssl" env:"FILE_SERVICE_MINIO_USE_SSL" env-default:"false"`
		Bucket    string `yaml:"bucket" env:"FILE_SERVICE_MINIO_BUCKET" env-default:"notes-files"`
	} `yaml:"minio"`
	RabbitMQ struct {
		Enabled  bool   `yaml:"enabled" env:"FILE_SERVICE_RABBITMQ_ENABLED" env-default:"false"`
		URL      string `yaml:"url" env:"FILE_SERVICE_RABBITMQ_URL" env-default:"amqp://guest:guest@localhost:5672/"`
		Exchange string `yaml:"exchange" env:"FILE_SERVICE_RABBITMQ_EXCHANGE" env-default:"notes.events"`
	} `yaml:"rabbitmq"`
	Listen struct {
		Type   string `yaml:"type" env:"FILE_SERVICE_LISTEN_TYPE" env-default:"port"`
		BindIP string `yaml:"bind_ip" env:"FILE_SERVICE_BIND_IP" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env:"FILE_SERVICE_PORT" env-default:"8085"`
	} `yaml:"listen"`
	Upload struct {
		MaxFileSizeMB int64 `yaml:"max_file_size_mb" env:"FILE_SERVICE_MAX_FILE_SIZE_MB" env-default:"32"`
	} `yaml:"upload"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		_ = godotenv.Load(envPath())

		instance = &Config{}
		if err := cleanenv.ReadConfig(configPath(), instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info("config help", "help", help)
			logger.Fatal("failed to read config", "error", err)
		}

		logger.Info(
			"config loaded",
			"address", instance.Listen.BindIP+":"+instance.Listen.Port,
			"minio_endpoint", instance.Minio.Endpoint,
			"minio_bucket", instance.Minio.Bucket,
			"max_file_size_mb", instance.Upload.MaxFileSizeMB,
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

func envPath() string {
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ".env"
}
