package models

import (
	"time"

	"github.com/google/uuid"
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
	GekozenCategorieID *int           `gorm:"index" json:"gekozen_categorie_id,omitempty"` // Nullable foreign key
	GekozenProductID   *int           `gorm:"index" json:"gekozen_product_id,omitempty"`   // Nullable foreign key

	Client           Client     `gorm:"foreignKey:ClientID;references:ID"`
	Behoefte         Behoefte   `gorm:"foreignKey:BehoefteID;references:ID"`
	GekozenCategorie *Categorie `gorm:"foreignKey:GekozenCategorieID;references:ID"`
	GekozenProduct   *Product   `gorm:"foreignKey:GekozenProductID;references:EAN"`
}

type Client struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Naam          string    `json:"naam"`
	Geboortedatum time.Time `json:"geboortedatum"`
	// Afhankelijk van wat de Aanvraag Verwerking Service nodig heeft van de client.
	// Dit kan een subset zijn van de Client in het ECD.
}

type Behoefte struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Beschrijving string    `json:"beschrijving"`
	// Deze Behoefte is waarschijnlijk een snapshot of een referentie naar de Behoefte in de Behoefte Bepaling Service.
	// Het zou hier niet alle details van de Behoefte in de Behoefte Bepaling Service hoeven te bevatten.
}

type Product struct {
	EAN         int       `gorm:"primaryKey;autoIncrement:false" json:"ean"`
	Naam        string    `json:"naam"`
	CategorieID int       `gorm:"index" json:"categorie_id"`
	Categorie   Categorie `gorm:"foreignKey:CategorieID;references:ID"`
}

type Categorie struct {
	ID   int    `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Naam string `json:"naam"`
}

// // --- DTO's voor communicatie met externe services (Recommendation Service) ---

// type CategorieAanvraagDTO struct {
// 	AanvraagID          uuid.UUID `json:"aanvraag_id"`
// 	BehoefteBeschrijving string    `json:"behoefte_beschrijving"`
// }

// type CategorieShortListDTO struct {
// 	Categorielijst []CategorieDTO `json:"categorie_lijst"`
// }

// type CategorieDTO struct {
// 	Naam         string    `json:"naam"`
// 	Prijsindicatie int       `json:"prijsindicatie"`
// }

// type ProductAanvraagDTO struct {
// 	AanvraagID          uuid.UUID `json:"aanvraag_id"`
// 	BehoefteBeschrijving string    `json:"behoefte_beschrijving"`
// 	GekozenCategorieID uuid.UUID `json:"gekozen_categorie_id"`
// }

// type ProductShortListDTO struct {
// 	Productlijst []ProductDTO `json:"product_lijst"`
// }

// type ProductDTO struct {
// 	EAN          int `json:"ean"`
// 	Naam        string `json:"naam"`
// 	Omschrijving string `json:"omschrijving"`
// 	Categorie   string `json:"categorie"` // Categorie naam of ID? Volgens UML String
// }

// // --- DTO voor communicatie met externe services (Technologie Bestel Service) ---

// type AanvraagBestelDTO struct {
// 	ClientID        uuid.UUID `json:"client_id"`
// 	GekozenProductID string    `json:"gekozen_product_id"` // EAN is een string
// }
