package behoefte_repo

import (
	"behoeftebepaling/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbInstance *gorm.DB

// InitDB initializes the database connection and performs auto-migrations.
func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database!")

	// Auto-Migrate models
	err = db.AutoMigrate(
		&models.Behoefte{},
		&models.Client{},
		&models.Onderzoek{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}
	log.Println("Database auto-migration complete.")

	return db, nil
}

func DBConnection() *gorm.DB {
	return dbInstance
}