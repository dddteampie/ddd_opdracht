package data_objects

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Product struct {
	EAN          int            `gorm:"primaryKey;autoIncrement:false" json:"ean"`
	SKU          string         `json:"sku"`
	Naam         string         `json:"naam"`
	Omschrijving string         `json:"omschrijving"`
	Merk         string         `json:"merk"`
	Afbeeldingen pq.StringArray `gorm:"type:text[]" json:"afbeeldingen"`
	Gewicht      float64        `json:"gewicht"`

	ProductTypeID uint        `gorm:"index" json:"productTypeID"`
	Type          ProductType `gorm:"foreignKey:ProductTypeID" json:"type"`

	Categorieen   []Categorie     `gorm:"many2many:product_categories;" json:"categorieen"`
	Specificaties []Specificatie  `gorm:"foreignKey:ProductEAN;references:EAN" json:"specificaties"`
	Reviews       []Review        `gorm:"foreignKey:ProductEAN;references:EAN" json:"reviews"`
	Tags          []Tag           `gorm:"many2many:product_tags;" json:"tags"`
	ProductAanbod []ProductAanbod `gorm:"foreignKey:ProductEAN;references:EAN" json:"productAanbod"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Naam string `gorm:"uniqueIndex" json:"naam"`
}

type Categorie struct {
	gorm.Model
	Naam              string     `json:"naam"`
	PriceRange        int        `json:"priceRange"`
	ParentCategorieID *uint      `gorm:"column:parent_categorie_id" json:"parentCategorieID,omitempty"`
	Products          []*Product `gorm:"many2many:product_categories;" json:"-"`
}

type Specificatie struct {
	gorm.Model
	ProductEAN int    `gorm:"index" json:"productEAN"`
	Naam       string `json:"naam"`
	Waarde     string `json:"waarde"`
}

type Review struct {
	gorm.Model
	ProductEAN int    `gorm:"index" json:"productEAN"`
	Naam       string `json:"naam"`
	Score      int    `json:"score"`
	Titel      string `json:"titel"`
	Inhoud     string `json:"inhoud"`
}

type ProductAanbod struct {
	gorm.Model
	ProductEAN    int      `gorm:"index" json:"productEAN"`
	Prijs         int      `json:"prijs"`
	Voorraad      int      `json:"voorraad"`
	LeverancierID uint     `gorm:"index" json:"leverancierID"`
	Supplier      Supplier `gorm:"foreignKey:LeverancierID;references:ID" json:"supplier"`
}

type ProductType struct {
	gorm.Model
	Naam         string `json:"naam"`
	Omschrijving string `json:"omschrijving"`
}

type Supplier struct {
	gorm.Model
	Name string `gorm:"column:name" json:"name"`
}
