package models

import (
	"time"

	"github.com/google/uuid"
)

type VerwerkingStatus string

const (
	BehoefteDoorgestuurd        VerwerkingStatus = "BehoefteDoorgestuurdNaarVerwerking"
	BehoefteNogNietDoorgestuurd VerwerkingStatus = "BehoefteNogNietDoorgestuurdNaarVerwerking"
)

type Behoefte struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	OnderzoekID  uuid.UUID        `gorm:"type:uuid" json:"onderzoek_id"`
	ClientID     uuid.UUID        `gorm:"type:uuid" json:"client_id"`
	Titel        string           `json:"titel"`
	Beschrijving string           `json:"beschrijving"`
	Urgentie     string           `json:"urgentie"`
	Datum        time.Time        `json:"datum"`
	Status       VerwerkingStatus `gorm:"type:string" json:"status"`
	Onderzoek    Onderzoek        `gorm:"foreignKey:OnderzoekID;references:ID"`
	Client       Client           `gorm:"foreignKey:ClientID;references:ID"`
}

type Client struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Naam          string    `json:"naam"`
	Geboortedatum time.Time `json:"geboortedatum"`
}

type Onderzoek struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Situatie   string    `json:"situatie"`
	BeginDatum time.Time `json:"begin_datum"`
	EindDatum  time.Time `json:"eind_datum"`
}

type ClientDTO struct {
	ID            uuid.UUID `json:"id"`
	Naam          string    `json:"naam"`
	Adres         string    `json:"adres"`
	Geboortedatum time.Time `json:"geboortedatum"`
}

type ZorgdossierDTO struct {
	ID       uuid.UUID `json:"id"`
	ClientID uuid.UUID `json:"client_id"`
	Situatie string    `json:"situatie"`
}

type OnderzoekDTO struct {
	ID            uuid.UUID          `json:"id"`
	ZorgdossierID uuid.UUID          `json:"zorgdossier_id"`
	Situatie      string             `json:"situatie"`
	BeginDatum    time.Time          `json:"begin_datum"`
	EindDatum     time.Time          `json:"eind_datum"`
	Diagnose      []DiagnoseDTO      `json:"diagnose"`
	Anamnese      []AnamneseDTO      `json:"anamnese"`
	Meetresultaat []MeetresultaatDTO `json:"meetresultaat"`
}

type AnamneseDTO struct {
	ID               uuid.UUID `json:"id"`
	OnderzoekID      uuid.UUID `json:"onderzoek_id"`
	Klachten         string    `json:"klachten"`
	DuurKlachten     string    `json:"duur_klachten"`
	Medicatiegebruik string    `json:"medicatiegebruik"`
	Allergieën       string    `json:"allergieën"`
	Leefstijl        string    `json:"leefstijl"`
	Datum            time.Time `json:"datum"`
}

type MeetresultaatDTO struct {
	InstrumentNaam string    `json:"instrument_naam"`
	Meetwaarde     string    `json:"meetwaarde"`
	Beschrijving   string    `json:"beschrijving"`
	Datum          time.Time `json:"datum"`
}

type DiagnoseDTO struct {
	Naam         string    `json:"naam"`
	Beschrijving string    `json:"beschrijving"`
	Datum        time.Time `json:"datum"`
}
