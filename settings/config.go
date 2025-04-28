package setting

import (
	"os"

	"gopkg.in/yaml.v3" // Updated to v3
)

type Config struct {
	Database      DatabaseConfig `yaml:"database"`
	Server        ServerConfig   `yaml:"server"`
	JWTSecret     string         `yaml:"jwt-secret"`
	JWTExpiration string         `yaml:"jwt-expiration"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

// Remove the AuthConfig struct since we're not using it anymore

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
