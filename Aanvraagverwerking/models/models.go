package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AanvraagStatus string

const (
	BehoefteOntvangen       AanvraagStatus = "BehoefteOntvangen"
	WachtenOpCategorieKeuze AanvraagStatus = "WachtenOpCategorieKeuze"
	CategorieGekozen        AanvraagStatus = "CategorieGekozen"
	WachtenOpProductKeuze   AanvraagStatus = "WachtenOpProductKeuze"
	ProductGekozen          AanvraagStatus = "ProductGekozen"
)

type Aanvraag struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ClientID           uuid.UUID      `gorm:"type:uuid;index" json:"client_id"`
	BehoefteID         uuid.UUID      `gorm:"type:uuid;index" json:"behoefte_id"` // Referentie naar de Behoefte in de Behoefte Bepaling Service
	Status             AanvraagStatus `gorm:"type:string" json:"status"`
	Budget             float64        `json:"budget"`                                                // Budget voor de aanvraag
	GekozenCategorieID *int           `gorm:"index" json:"gekozen_categorie_id,omitempty"`           // Nullable foreign key
	GekozenProductID   *int64         `gorm:"type:bigint;index" json:"gekozen_product_id,omitempty"` // Nullable foreign key
	CategorieOpties    pq.Int64Array  `gorm:"type:integer[]" json:"categorie_opties,omitempty"`      // Optionele categorieën voor de aanvraag
	ProductOpties      pq.Int64Array  `gorm:"type:bigint[]" json:"product_opties,omitempty"`         // Optionele producten voor de aanvraag

	Client   Client   `gorm:"foreignKey:ClientID;references:ID"`
	Behoefte Behoefte `gorm:"foreignKey:BehoefteID;references:ID"`
	//GekozenCategorie *Categorie `gorm:"foreignKey:GekozenCategorieID;references:ID"`
	//GekozenProduct   *Product   `gorm:"foreignKey:GekozenProductID;references:EAN"`
}

type Client struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Naam          string    `json:"naam"`
	Geboortedatum time.Time `json:"geboortedatum"`
}

type Behoefte struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Beschrijving string    `json:"beschrijving"`
}

// type Product struct {
// 	EAN         int    `gorm:"primaryKey;autoIncrement:false" json:"ean"`
// 	Naam        string `json:"naam"`
// 	CategorieID int    `gorm:"index" json:"categorie_id"`
// 	Categorie   Categorie `gorm:"foreignKey:CategorieID;references:ID"`
// }

// type Categorie struct {
// 	ID   int    `gorm:"primaryKey;autoIncrement:false" json:"id"`
// 	Naam string `json:"naam"`
// }

// // --- DTO's voor communicatie met externe services (Recommendation Service) ---

type CategorieAanvraagDTO struct {
	ClientID             string  `json:"patientId"`
	Budget               float64 `json:"budget"`
	BehoefteBeschrijving string  `json:"behoeften"`
}

type CategorieShortListDTO struct {
	CategoryIDs    pq.Int64Array  `gorm:"type:integer[]" json:"-"`
	Categorielijst []CategorieDTO `json:"categories"`
}

type CategorieDTO struct {
	ID   int    `json:"id"`
	Naam string `json:"naam"`
}

type ProductAanvraagDTO struct {
	ClientID             string  `json:"clientId"`
	Budget               float64 `json:"budget"`
	BehoefteBeschrijving string  `json:"behoeften"`
	GekozenCategorieID   *int    `json:"CategorieID,omitempty"`
}

type ProductShortListDTO struct {
	ProductEANs  pq.Int64Array `gorm:"type:bigint[]" json:"-"`
	Productlijst []ProductDTO  `json:"products"`
}

type ProductDTO struct {
	EAN          int64  `json:"ean"`
	Naam         string `json:"naam"`
	Omschrijving string `json:"omschrijving"`
}

// // --- DTO voor communicatie met externe services (Technologie Bestel Service) ---
// type AanvraagBestelDTO struct {
// 	ClientID        uuid.UUID `json:"client_id"`
// 	GekozenProductID string    `json:"gekozen_product_id"` // EAN is een string
// }