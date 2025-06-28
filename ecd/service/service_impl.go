package service

import (
	"context"
	"ecd/api/dto"
	"ecd/data/repository"

	"github.com/gofrs/uuid"
)

type ECDServiceImpl struct {
	repo repository.Repository
}

func NewECDService(repo repository.Repository) *ECDServiceImpl {
	return &ECDServiceImpl{repo: repo}
}

func (s *ECDServiceImpl) GetClient(ctx context.Context, id uuid.UUID) (*dto.ClientDTO, error) {
	return s.repo.GetClientByID(id)
}

func (s *ECDServiceImpl) CreateClient(ctx context.Context, dto dto.ClientDTO) error {
	return s.repo.SaveClient(dto)
}

func (s *ECDServiceImpl) GetZorgdossierByClientID(ctx context.Context, clientID uuid.UUID) (*dto.ZorgdossierDTO, error) {
	return s.repo.GetZorgdossierByClientID(clientID)
}

func (s *ECDServiceImpl) CreateZorgdossier(ctx context.Context, dto dto.ZorgdossierDTO) error {
	return s.repo.SaveZorgdossier(dto)
}

func (s *ECDServiceImpl) CreateOnderzoek(ctx context.Context, dto dto.OnderzoekDTO) error {
	return s.repo.CreateOnderzoek(dto)
}

func (s *ECDServiceImpl) AddAnamnese(ctx context.Context, onderzoekId uuid.UUID, anamnese dto.AnamneseDTO) error {
	return s.repo.AddAnamnese(onderzoekId, anamnese)
}

func (s *ECDServiceImpl) AddDiagnose(ctx context.Context, onderzoekId uuid.UUID, diagnose dto.DiagnoseDTO) error {
	return s.repo.AddDiagnose(onderzoekId, diagnose)
}

func (s *ECDServiceImpl) AddMeetresultaat(ctx context.Context, onderzoekID uuid.UUID, meetresultaat dto.MeetresultaatDTO) error {
	return s.repo.AddMeetresultaat(onderzoekID, meetresultaat)
}

func (s *ECDServiceImpl) GetOnderzoekByID(ctx context.Context, onderzoekID uuid.UUID) (*dto.OnderzoekDTO, error) {
	return s.repo.GetOnderzoekByID(onderzoekID)
}

func (s *ECDServiceImpl) GetOnderzoekenByZorgdossierID(ctx context.Context, zorgdossierID uuid.UUID) ([]dto.OnderzoekDTO, error) {
	return s.repo.GetOnderzoekenByZorgdossierID(zorgdossierID)
}

func (s *ECDServiceImpl) UpdateOnderzoek(ctx context.Context, dto dto.OnderzoekDTO) error {
	return s.repo.UpdateOnderzoek(dto)
}
