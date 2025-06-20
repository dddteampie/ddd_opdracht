package data_handling

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"productservice/data_access/data_objects"
)

// InitDB initializes the database connection and performs auto-migrations.
func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database!")

	// Auto-Migrate models
	err = db.AutoMigrate(
		&data_objects.Product{},
		&data_objects.Categorie{},
		&data_objects.Specificatie{},
		&data_objects.Review{},
		&data_objects.ProductAanbod{},
		&data_objects.ProductType{},
		&data_objects.Supplier{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}
	log.Println("Database auto-migration complete.")

	return db, nil
}
