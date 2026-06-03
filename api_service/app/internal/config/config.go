package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"path/filepath"

	"myproject/pkg/logging"
	"sync"
)

type Config struct {
	IsDebug *bool `yaml:"is_debug"`
	JWT     struct {
		Secret string `yaml:"secret" env:"API_SERVICE_JWT_SECRET" env-required:"true"`
	}
	CategoryService struct {
		URL string `yaml:"url" env:"API_SERVICE_CATEGORY_URL" env-required:"true"`
	} `yaml:"category_service" env-required:"true"`
	NoteService struct {
		URL string `yaml:"url" env:"API_SERVICE_NOTE_URL" env-required:"true"`
	} `yaml:"note_service" env-required:"true"`
	FileService struct {
		URL string `yaml:"url" env:"API_SERVICE_FILE_URL" env-required:"true"`
	} `yaml:"file_service" env-required:"true"`
	UserService struct {
		URL string `yaml:"url" env:"API_SERVICE_USER_URL" env-required:"true"`
	} `yaml:"user_service" env-required:"true"`
	SearchService struct {
		URL string `yaml:"url" env:"API_SERVICE_SEARCH_URL" env-required:"true"`
	} `yaml:"search_service" env-required:"true"`

	Listen struct {
		Type   string `yaml:"type" env:"API_SERVICE_LISTEN_TYPE" env-default:"port"`
		BindIP string `yaml:"bind_ip" env:"API_SERVICE_BIND_IP" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env:"API_SERVICE_PORT" env-default:"8080"`
	}
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("read application config")

		instance = &Config{}

		if err := cleanenv.ReadConfig(configPath(), instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
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
