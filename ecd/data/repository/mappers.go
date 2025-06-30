package repository

import (
	"ecd/api/dto"
	"ecd/data/models"
)

func ToClientModel(dto dto.ClientDTO) models.Client {
	return models.Client{
		ID:            dto.ID,
		Naam:          dto.Naam,
		Adres:         dto.Adres,
		Geboortedatum: dto.Geboortedatum,
	}
}

func ToClientDTO(m models.Client) dto.ClientDTO {
	return dto.ClientDTO{
		ID:            m.ID,
		Naam:          m.Naam,
		Adres:         m.Adres,
		Geboortedatum: m.Geboortedatum,
	}
}

func ToZorgdossierModel(dto dto.ZorgdossierDTO) models.Zorgdossier {
	return models.Zorgdossier{
		ID:       dto.ID,
		ClientID: dto.ClientID,
		Situatie: dto.Situatie,
	}
}

func ToZorgdossierDTO(m models.Zorgdossier) dto.ZorgdossierDTO {
	return dto.ZorgdossierDTO{
		ID:       m.ID,
		ClientID: m.ClientID,
		Situatie: m.Situatie,
	}
}

func ToOnderzoekModel(dto dto.OnderzoekDTO) models.Onderzoek {
	diagnoses := make([]models.Diagnose, len(dto.Diagnose))
	for i, d := range dto.Diagnose {
		diagnoses[i] = ToDiagnoseModel(d)
	}
	anamneses := make([]models.Anamnese, len(dto.Anamnese))
	for i, a := range dto.Anamnese {
		anamneses[i] = ToAnamneseModel(a)
	}
	meetresultaten := make([]models.Meetresultaat, len(dto.Meetresultaat))
	for i, m := range dto.Meetresultaat {
		meetresultaten[i] = ToMeetresultaatModel(m)
	}
	return models.Onderzoek{
		ID:            dto.ID,
		ZorgdossierID: dto.ZorgdossierID,
		BeginDatum:    dto.BeginDatum,
		EindDatum:     dto.EindDatum,
		Diagnose:      diagnoses,
		Anamnese:      anamneses,
		Meetresultaat: meetresultaten,
	}
}

func ToOnderzoekDTO(m models.Onderzoek) dto.OnderzoekDTO {
	diagnoses := make([]dto.DiagnoseDTO, len(m.Diagnose))
	for i, d := range m.Diagnose {
		diagnoses[i] = ToDiagnoseDTO(d)
	}
	anamneses := make([]dto.AnamneseDTO, len(m.Anamnese))
	for i, a := range m.Anamnese {
		anamneses[i] = ToAnamneseDTO(a)
	}
	meetresultaten := make([]dto.MeetresultaatDTO, len(m.Meetresultaat))
	for i, mr := range m.Meetresultaat {
		meetresultaten[i] = ToMeetresultaatDTO(mr)
	}
	return dto.OnderzoekDTO{
		ID:            m.ID,
		ZorgdossierID: m.ZorgdossierID,
		BeginDatum:    m.BeginDatum,
		EindDatum:     m.EindDatum,
		Diagnose:      diagnoses,
		Anamnese:      anamneses,
		Meetresultaat: meetresultaten,
	}
}

func ToAnamneseModel(dto dto.AnamneseDTO) models.Anamnese {
	return models.Anamnese{
		ID:               dto.ID,
		OnderzoekID:      dto.OnderzoekID,
		Klachten:         dto.Klachten,
		DuurKlachten:     dto.DuurKlachten,
		Medicatiegebruik: dto.Medicatiegebruik,
		Allergieën:       dto.Allergieën,
		Leefstijl:        dto.Leefstijl,
		Datum:            dto.Datum,
	}
}

func ToAnamneseDTO(m models.Anamnese) dto.AnamneseDTO {
	return dto.AnamneseDTO{
		ID:               m.ID,
		OnderzoekID:      m.OnderzoekID,
		Klachten:         m.Klachten,
		DuurKlachten:     m.DuurKlachten,
		Medicatiegebruik: m.Medicatiegebruik,
		Allergieën:       m.Allergieën,
		Leefstijl:        m.Leefstijl,
		Datum:            m.Datum,
	}
}

func ToDiagnoseModel(dto dto.DiagnoseDTO) models.Diagnose {
	return models.Diagnose{
		ID:           dto.ID,
		OnderzoekID:  dto.OnderzoekID,
		Diagnosecode: dto.Diagnosecode,
		Naam:         dto.Naam,
		Toelichting:  dto.Toelichting,
		Datum:        dto.Datum,
		Status:       dto.Status,
	}
}

func ToDiagnoseDTO(m models.Diagnose) dto.DiagnoseDTO {
	return dto.DiagnoseDTO{
		ID:           m.ID,
		OnderzoekID:  m.OnderzoekID,
		Diagnosecode: m.Diagnosecode,
		Naam:         m.Naam,
		Toelichting:  m.Toelichting,
		Datum:        m.Datum,
		Status:       m.Status,
	}
}

func ToMeetresultaatModel(dto dto.MeetresultaatDTO) models.Meetresultaat {
	return models.Meetresultaat{
		ID:             dto.ID,
		OnderzoekID:    dto.OnderzoekID,
		InstrumentNaam: dto.InstrumentNaam,
		Meetwaarde:     dto.Meetwaarde,
		Datum:          dto.Datum,
		UitgevoerdDoor: dto.UitgevoerdDoor,
	}
}

func ToMeetresultaatDTO(m models.Meetresultaat) dto.MeetresultaatDTO {
	return dto.MeetresultaatDTO{
		ID:             m.ID,
		OnderzoekID:    m.OnderzoekID,
		InstrumentNaam: m.InstrumentNaam,
		Meetwaarde:     m.Meetwaarde,
		Datum:          m.Datum,
		UitgevoerdDoor: m.UitgevoerdDoor,
	}
}
