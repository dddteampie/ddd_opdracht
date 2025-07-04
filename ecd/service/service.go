package service

import (
	"context"
	"ecd/api/dto"

	"github.com/gofrs/uuid"
)

type ECDService interface {
	GetClient(ctx context.Context, id uuid.UUID) (*dto.ClientDTO, error)
	GetAllClients(ctx context.Context) ([]dto.ClientDTO, error)
	CreateClient(ctx context.Context, dto dto.ClientDTO) error

	GetZorgdossierByClientID(ctx context.Context, clientID uuid.UUID) (*dto.ZorgdossierDTO, error)
	CreateZorgdossier(ctx context.Context, dto dto.ZorgdossierDTO) error

	CreateOnderzoek(ctx context.Context, dto dto.OnderzoekDTO) error
	GetOnderzoekByID(ctx context.Context, onderzoekID uuid.UUID) (*dto.OnderzoekDTO, error)
	UpdateOnderzoek(ctx context.Context, dto dto.OnderzoekDTO) error
	AddAnamnese(ctx context.Context, onderzoekID uuid.UUID, dto dto.AnamneseDTO) error
	AddMeetresultaat(ctx context.Context, onderzoekID uuid.UUID, dto dto.MeetresultaatDTO) error
	AddDiagnose(ctx context.Context, onderzoekID uuid.UUID, dto dto.DiagnoseDTO) error
}
