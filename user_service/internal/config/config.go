package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"

	"user_service/pkg/logging"
)

type Config struct {
	IsDebug *bool `yaml:"is_debug"`
	Mongo   struct {
		Scheme     string `yaml:"scheme" env-default:"mongodb"`
		Host       string `yaml:"host" env-default:"localhost"`
		Port       string `yaml:"port" env-default:"27017"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		AuthSource string `yaml:"auth_source" env-default:"admin"`
		Database   string `yaml:"database" env-default:"notes_system"`
		Collection string `yaml:"collection" env-default:"users"`
	} `yaml:"mongo"`
	Listen struct {
		Type   string `yaml:"type" env-default:"port"`
		BindIP string `yaml:"bind_ip" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env-default:"8083"`
	} `yaml:"listen"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		instance = &Config{}

		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info("config help", "help", help)
			logger.Fatal("failed to read config", "error", err)
		}

		logger.Info(
			"config loaded",
			"address", instance.Listen.BindIP+":"+instance.Listen.Port,
			"mongo", instance.Mongo.Host+":"+instance.Mongo.Port,
			"database", instance.Mongo.Database,
		)
	})

	return instance
}
