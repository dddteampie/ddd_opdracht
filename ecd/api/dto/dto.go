package dto

import (
	"time"

	"github.com/gofrs/uuid"
)

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

type DiagnoseDTO struct {
	ID           uuid.UUID `json:"id"`
	OnderzoekID  uuid.UUID `json:"onderzoek_id"`
	Diagnosecode string    `json:"diagnosecode"`
	Naam         string    `json:"naam"`
	Toelichting  string    `json:"toelichting"`
	Datum        time.Time `json:"datum"`
	Status       string    `json:"status"`
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
	ID             uuid.UUID `json:"id"`
	OnderzoekID    uuid.UUID `json:"onderzoek_id"`
	InstrumentNaam string    `json:"instrument_naam"`
	Meetwaarde     string    `json:"meetwaarde"`
	Datum          time.Time `json:"datum"`
	UitgevoerdDoor string    `json:"uitgevoerd_door"`
}
