package repository

import (
	"context"
	"fmt"
	"log"
	models "recommendation/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database!")

	err = db.AutoMigrate(
		&models.Aanbeveling{},
		&models.PassendeCategorieënLijst{},
		&models.OplossingenLijst{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}
	log.Println("Database auto-migration complete.")

	return db, nil
}

type AanbevelingsOpslag struct {
	db *gorm.DB
}

func NewAanbevelingsOpslag(db *gorm.DB) *AanbevelingsOpslag {
	return &AanbevelingsOpslag{db: db}
}

func (r *AanbevelingsOpslag) SlaAanbevelingOp(ctx context.Context, rec *models.Aanbeveling) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *AanbevelingsOpslag) HaalAanbevelingOpMetCliëntID(ctx context.Context, clientID string) (*models.Aanbeveling, error) {
	var rec models.Aanbeveling
	err := r.db.WithContext(ctx).Where("client_id = ?", clientID).
		Order("versie DESC").
		Preload(clause.Associations).
		First(&rec).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &rec, err
}

func (r *AanbevelingsOpslag) WerkAanbevelingBij(ctx context.Context, rec *models.Aanbeveling) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *AanbevelingsOpslag) MaakPassendeCategorieënLijstDB(ctx context.Context, lijst *models.PassendeCategorieënLijst) error {
	log.Printf("Repository: Opslaan PassendeCategorieënLijst met Categorie ID's: %v", lijst.CategoryIDs)
	return r.db.WithContext(ctx).Create(lijst).Error
}

func (r *AanbevelingsOpslag) HaalPassendeCategorieënLijstOpMetID(ctx context.Context, id uint) (*models.PassendeCategorieënLijst, error) {
	var lijst models.PassendeCategorieënLijst
	err := r.db.WithContext(ctx).First(&lijst, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &lijst, err
}

func (r *AanbevelingsOpslag) MaakOplossingenLijstDB(ctx context.Context, lijst *models.OplossingenLijst) error {
	log.Printf("Repository: Opslaan OplossingenLijst met Product EAN's: %v", lijst.ProductEANs)
	return r.db.WithContext(ctx).Create(lijst).Error
}

func (r *AanbevelingsOpslag) HaalOplossingenLijstOpMetID(ctx context.Context, id uint) (*models.OplossingenLijst, error) {
	var lijst models.OplossingenLijst
	err := r.db.WithContext(ctx).First(&lijst, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &lijst, err
}

func (r *AanbevelingsOpslag) WerkPassendeCategorieënLijstBijDB(ctx context.Context, lijst *models.PassendeCategorieënLijst) error {
	log.Printf("Repository: Bijwerken PassendeCategorieënLijst met ID %d en Categorie ID's: %v", lijst.ID, lijst.CategoryIDs)
	return r.db.WithContext(ctx).Save(lijst).Error
}

func (r *AanbevelingsOpslag) WerkOplossingenLijstBijDB(ctx context.Context, lijst *models.OplossingenLijst) error {
	log.Printf("Repository: Bijwerken OplossingenLijst met ID %d en Product EAN's: %v", lijst.ID, lijst.ProductEANs)
	return r.db.WithContext(ctx).Save(lijst).Error
}
