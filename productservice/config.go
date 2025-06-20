package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	DatabaseDSN string `json:"database_dsn"`
	ServerPort  string `json:"server_port"`
}

func LoadConfig(filePath string) (*Config, error) {
	if _, exists := os.LookupEnv("DOCKERIZED_ENV"); !exists {
		cfg, err := LoadConfigFromFile(filePath)
		if err != nil {
			log.Printf("Warning: Could not load config file. Shutting down! Err: %v.", err)
			return nil, err
		}
		log.Println(".env file loaded successfully for local dev.")
		return cfg, nil
	} else {
		cfg, err := LoadConfigFromEnv()
		if err != nil {
			log.Printf("Warning: Could not load config from environment. Err: %v.", err)
			return nil, err
		}
		return cfg, nil
	}
}

// LoadConfigFromEnv reads the configuration from env used in prod
func LoadConfigFromEnv() (*Config, error) {
	cfg := &Config{}

	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		log.Print(dsn)
		cfg.DatabaseDSN = dsn
	}
	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		cfg.ServerPort = serverPort
	}
	return cfg, nil
}

// LoadConfigFromFile reads the configuration from the specified JSON file path used local runtime ONLY
func LoadConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
