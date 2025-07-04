package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseDSN  string `json:"database_dsn"`
	ServerPort   string `json:"server_port"`
	AuthzDevMode bool   `json:"AuthzDevMode"`
	CorsOrigin   string `json:"CorsOrigin"`
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
	if AuthzDevMode := os.Getenv("AuthzDevMode"); AuthzDevMode != "" {
		cfg.AuthzDevMode, _ = strconv.ParseBool(AuthzDevMode)
	}
	if CorsOrigin := os.Getenv("CorsOrigin"); CorsOrigin != "" {
		cfg.CorsOrigin = CorsOrigin
	}

	return cfg, nil
}

// LoadConfigFromFile reads the configuration from the specified JSON file path used local runtime ONLY
func LoadConfigFromFile(filePath string) (*Config, error) {
	err := godotenv.Load(filePath)
	if err != nil {
		log.Printf("Error loading .env file from %s: %v", filePath, err)
		return nil, err
	}
	log.Printf(".env file loaded successfully from %s.", filePath)

	cfg := &Config{}
	cfg.DatabaseDSN = os.Getenv("DATABASE_DSN")
	cfg.ServerPort = os.Getenv("SERVER_PORT")
	cfg.AuthzDevMode, _ = strconv.ParseBool(os.Getenv("AuthzDevMode"))
	cfg.CorsOrigin = os.Getenv("CorsOrigin")

	if cfg.DatabaseDSN == "" {
		log.Println("Warning: DATABASE_DSN is not set in the .env file.")
	}
	if cfg.ServerPort == "" {
		log.Println("Warning: SERVER_PORT is not set in the .env file.")
	}
	if cfg.CorsOrigin == "" {
		log.Println("Warning: CorsOrigin is not set in the .env file.")
	}

	return cfg, nil
}
