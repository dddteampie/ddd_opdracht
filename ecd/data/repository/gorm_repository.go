package repository

import (
	dto "ecd/api/dto"
	model "ecd/data/models"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetClientByID(id uuid.UUID) (*dto.ClientDTO, error) {
	var model model.Client
	if err := r.db.First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	dto := ToClientDTO(model)
	return &dto, nil
}

func (r *GormRepository) SaveClient(dto dto.ClientDTO) error {
	model := ToClientModel(dto)
	return r.db.Create(&model).Error
}

func (r *GormRepository) GetZorgdossierByClientID(clientID uuid.UUID) (*dto.ZorgdossierDTO, error) {
	var model model.Zorgdossier
	if err := r.db.First(&model, "client_id = ?", clientID).Error; err != nil {
		return nil, err
	}
	dto := ToZorgdossierDTO(model)
	return &dto, nil
}

func (r *GormRepository) SaveZorgdossier(dto dto.ZorgdossierDTO) error {
	model := ToZorgdossierModel(dto)
	return r.db.Create(&model).Error
}

func (r *GormRepository) CreateOnderzoek(dto dto.OnderzoekDTO) error {
	model := ToOnderzoekModel(dto)
	return r.db.Create(&model).Error
}

func (r *GormRepository) AddAnamnese(anamnese dto.AnamneseDTO) error {
	model := ToAnamneseModel(anamnese)
	return r.db.Create(&model).Error
}

func (r *GormRepository) AddDiagnose(diagnose dto.DiagnoseDTO) error {
	model := ToDiagnoseModel(diagnose)
	return r.db.Create(&model).Error
}

func (r *GormRepository) AddMeetresultaat(meetresultaat dto.MeetresultaatDTO) error {
	model := ToMeetresultaatModel(meetresultaat)
	return r.db.Create(&model).Error
}
