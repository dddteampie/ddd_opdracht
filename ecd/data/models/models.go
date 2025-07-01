package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Client struct {
	gorm.Model
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Naam          string    `gorm:"naam"`
	Adres         string    `gorm:"adres"`
	Geboortedatum time.Time `gorm:"type:date"`
}

type Zorgdossier struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	ClientID uuid.UUID `gorm:"type:uuid"`
	Situatie string    `gorm:"situatie"`
}

type Onderzoek struct {
	gorm.Model
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"`
	ZorgdossierID uuid.UUID       `gorm:"type:uuid;foreignKey:ZorgdossierID;references:ID"`
	BeginDatum    time.Time       `gorm:"type:date"`
	EindDatum     time.Time       `gorm:"type:date"`
	Diagnose      []Diagnose      `gorm:"foreignKey:OnderzoekID;references:ID"`
	Anamnese      []Anamnese      `gorm:"foreignKey:OnderzoekID;references:ID"`
	Meetresultaat []Meetresultaat `gorm:"foreignKey:OnderzoekID;references:ID"`
}

type Anamnese struct {
	gorm.Model
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	OnderzoekID      uuid.UUID `gorm:"type:uuid"`
	Klachten         string    `gorm:"klachten"`
	DuurKlachten     string    `gorm:"duur_klachten"`
	Medicatiegebruik string    `gorm:"medicatiegebruik"`
	Allergieën       string    `gorm:"allergieën"`
	Leefstijl        string    `gorm:"leefstijl"`
	Datum            time.Time `gorm:"type:date"`
}

type Diagnose struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	OnderzoekID  uuid.UUID `gorm:"type:uuid"`
	Diagnosecode string    `gorm:"diagnosecode"`
	Naam         string    `gorm:"naam"`
	Toelichting  string    `gorm:"toelichting"`
	Datum        time.Time `gorm:"type:date"`
	Status       string    `gorm:"status"`
}
type Meetresultaat struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	OnderzoekID    uuid.UUID `gorm:"type:uuid"`
	InstrumentNaam string    `gorm:"instrument_naam"`
	Meetwaarde     string    `gorm:"meetwaarde"`
	Datum          time.Time `gorm:"type:date"`
	UitgevoerdDoor string    `gorm:"uitgevoerd_door"`
}
