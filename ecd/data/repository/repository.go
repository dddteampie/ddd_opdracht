package repository

import (
	"ecd/api/dto"

	"github.com/gofrs/uuid"
)

type Repository interface {
	GetClientByID(id uuid.UUID) (*dto.ClientDTO, error)
	GetAllClients() ([]dto.ClientDTO, error)
	SaveClient(dto dto.ClientDTO) error
	GetZorgdossierByClientID(clientID uuid.UUID) (*dto.ZorgdossierDTO, error)
	SaveZorgdossier(dto dto.ZorgdossierDTO) error
	CreateOnderzoek(dto dto.OnderzoekDTO) error
	AddAnamnese(onderzoekID uuid.UUID, dto dto.AnamneseDTO) error
	AddMeetresultaat(onderzoekID uuid.UUID, dto dto.MeetresultaatDTO) error
	AddDiagnose(onderzoekID uuid.UUID, dto dto.DiagnoseDTO) error
	GetOnderzoekByID(id uuid.UUID) (*dto.OnderzoekDTO, error)
	GetOnderzoekenByZorgdossierID(zorgdossierID uuid.UUID) ([]dto.OnderzoekDTO, error)
	UpdateOnderzoek(dto dto.OnderzoekDTO) error
}
