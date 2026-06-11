package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"

	"user_service/pkg/logging"
)

type Config struct {
	IsDebug  *bool `yaml:"is_debug"`
	Postgres struct {
		Host     string `yaml:"host" env:"USER_SERVICE_POSTGRES_HOST" env-default:"localhost"`
		Port     string `yaml:"port" env:"USER_SERVICE_POSTGRES_PORT" env-default:"5432"`
		Username string `yaml:"username" env:"USER_SERVICE_POSTGRES_USERNAME" env-default:"notes_user"`
		Password string `yaml:"password" env:"USER_SERVICE_POSTGRES_PASSWORD" env-default:"notes_password"`
		Database string `yaml:"database" env:"USER_SERVICE_POSTGRES_DATABASE" env-default:"users_system"`
		SSLMode  string `yaml:"ssl_mode" env:"USER_SERVICE_POSTGRES_SSL_MODE" env-default:"disable"`
	} `yaml:"postgres"`
	RabbitMQ struct {
		Enabled  bool   `yaml:"enabled" env:"USER_SERVICE_RABBITMQ_ENABLED" env-default:"false"`
		URL      string `yaml:"url" env:"USER_SERVICE_RABBITMQ_URL" env-default:"amqp://guest:guest@localhost:5672/"`
		Exchange string `yaml:"exchange" env:"USER_SERVICE_RABBITMQ_EXCHANGE" env-default:"notes.events"`
	} `yaml:"rabbitmq"`
	Listen struct {
		Type   string `yaml:"type" env:"USER_SERVICE_LISTEN_TYPE" env-default:"port"`
		BindIP string `yaml:"bind_ip" env:"USER_SERVICE_BIND_IP" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env:"USER_SERVICE_PORT" env-default:"8083"`
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
			"postgres", instance.Postgres.Host+":"+instance.Postgres.Port,
			"database", instance.Postgres.Database,
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
