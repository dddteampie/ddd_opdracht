package database

import (
	"ecd/data/models"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dsn string) (*gorm.DB, error) {
	log.Println("Connecting to PostgreSQL database with", dsn)
	conn, err := pgx.ParseConfig(dsn)
	log.Println("Parsed DSN:", conn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database!")

	// Auto-Migrate models
	err = db.AutoMigrate(
		&models.Client{}, &models.Zorgdossier{}, &models.Onderzoek{},
		&models.Anamnese{}, &models.Diagnose{}, &models.Meetresultaat{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}
	log.Println("Database auto-migration complete.")
	return db, nil
}
