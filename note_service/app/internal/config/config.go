package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"

	"note_service/pkg/logging"
)

type Config struct {
	IsDebug *bool `yaml:"is_debug"`
	Mongo   struct {
		Scheme     string `yaml:"scheme" env:"NOTE_SERVICE_MONGO_SCHEME" env-default:"mongodb"`
		Host       string `yaml:"host" env:"NOTE_SERVICE_MONGO_HOST" env-default:"localhost"`
		Port       string `yaml:"port" env:"NOTE_SERVICE_MONGO_PORT" env-default:"27017"`
		Username   string `yaml:"username" env:"NOTE_SERVICE_MONGO_USERNAME"`
		Password   string `yaml:"password" env:"NOTE_SERVICE_MONGO_PASSWORD"`
		AuthSource string `yaml:"auth_source" env:"NOTE_SERVICE_MONGO_AUTH_SOURCE" env-default:"admin"`
		Database   string `yaml:"database" env:"NOTE_SERVICE_MONGO_DATABASE" env-default:"notes_system"`
		Collection string `yaml:"collection" env:"NOTE_SERVICE_MONGO_COLLECTION" env-default:"notes"`
	} `yaml:"mongo"`
	RabbitMQ struct {
		Enabled               bool   `yaml:"enabled" env:"NOTE_SERVICE_RABBITMQ_ENABLED" env-default:"false"`
		URL                   string `yaml:"url" env:"NOTE_SERVICE_RABBITMQ_URL" env-default:"amqp://guest:guest@localhost:5672/"`
		Exchange              string `yaml:"exchange" env:"NOTE_SERVICE_RABBITMQ_EXCHANGE" env-default:"notes.events"`
		NoteUpdatedRoutingKey string `yaml:"note_updated_routing_key" env:"NOTE_SERVICE_RABBITMQ_NOTE_UPDATED_ROUTING_KEY" env-default:"note.updated"`
	} `yaml:"rabbitmq"`
	Listen struct {
		Type   string `yaml:"type" env:"NOTE_SERVICE_LISTEN_TYPE" env-default:"port"`
		BindIP string `yaml:"bind_ip" env:"NOTE_SERVICE_BIND_IP" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env:"NOTE_SERVICE_PORT" env-default:"8082"`
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
			"mongo", instance.Mongo.Host+":"+instance.Mongo.Port,
			"database", instance.Mongo.Database,
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
