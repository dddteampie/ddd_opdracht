package service

import (
	"context"
	"ecd/api/dto"
	"ecd/data/repository"

	"github.com/gofrs/uuid"
)

type ECDServiceImpl struct {
	repo *repository.GormRepository
}

func NewECDService(repo *repository.GormRepository) *ECDServiceImpl {
	return &ECDServiceImpl{repo: repo}
}

func (s *ECDServiceImpl) GetClient(ctx context.Context, id uuid.UUID) (*dto.ClientDTO, error) {
	return s.repo.GetClientByID(id)
}

func (s *ECDServiceImpl) CreateClient(ctx context.Context, dto dto.ClientDTO) error {
	return s.repo.SaveClient(dto)
}

// Same pattern for other methods...
func (s *ECDServiceImpl) GetZorgdossierByClientID(ctx context.Context, clientID uuid.UUID) (*dto.ZorgdossierDTO, error) {
	return s.repo.GetZorgdossierByClientID(clientID)
}
