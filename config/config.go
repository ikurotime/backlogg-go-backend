package config

import (
	"ikurotime/backlog-go-backend/pkg/root"
	"ikurotime/backlog-go-backend/pkg/yamlx"
	"log"
	"os"
)

type ClerkConfig struct {
	ApiKey string `yaml:"apiKey"`
}

type MongoDBConfig struct {
	Protocol string `yaml:"protocol"`
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	DB       int    `yaml:"db"`
}

type Config struct {
	MongoDBConfig MongoDBConfig `yaml:"mongodb"`
	ClerkConfig   ClerkConfig   `yaml:"clerk"`
}

func LoadConfig() (*Config, error) {
	var err error
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	cfg := &Config{}

	err = yamlx.ReadFile(root.GetRootDir()+"/config/.env."+env, cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	return cfg, err
}
