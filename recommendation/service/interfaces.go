package service

import (
	"context"
	"net/http"
	models "recommendation/model"
)

type IAanbevelingsOpslag interface {
	SlaAanbevelingOp(ctx context.Context, rec *models.Aanbeveling) error
	HaalAanbevelingOpMetCliëntID(ctx context.Context, clientID string) (*models.Aanbeveling, error)
	WerkAanbevelingBij(ctx context.Context, rec *models.Aanbeveling) error
	MaakPassendeCategorieënLijstDB(ctx context.Context, lijst *models.PassendeCategorieënLijst) error
	HaalPassendeCategorieënLijstOpMetID(ctx context.Context, id uint) (*models.PassendeCategorieënLijst, error)
	MaakOplossingenLijstDB(ctx context.Context, lijst *models.OplossingenLijst) error
	HaalOplossingenLijstOpMetID(ctx context.Context, id uint) (*models.OplossingenLijst, error)
	WerkOplossingenLijstBijDB(ctx context.Context, lijst *models.OplossingenLijst) error
	WerkPassendeCategorieënLijstBijDB(ctx context.Context, lijst *models.PassendeCategorieënLijst) error
}

type IAanbevelingHelpers interface {
	MaakPassendeCategorieënLijst(ctx context.Context, patientID string, budget float64, behoeften string) (*models.PassendeCategorieënLijst, error)
	MaakOplossingenLijst(ctx context.Context, clientID string, budget float64, behoeften string, categoryID *int) (*models.OplossingenLijst, error)
	HaalPassendeCategorieënLijstOp(ctx context.Context, patientID string) (*models.PassendeCategorieënLijst, error)
	HaalOplossingenLijstOp(ctx context.Context, clientID string) (*models.OplossingenLijst, error)
	HaalAlleTagsOp(ctx context.Context, categoryID *int) ([]string, error)
	HaalCategorieënOp(ctx context.Context, budget float64) ([]models.Category, error)
	HaalProductenOp(ctx context.Context, tags []string, budget float64, categorieën []int) ([]models.Product, error)
	HaalCategorieenOpMetIDs(ctx context.Context, ids []int) ([]models.Category, error)
	HaalProductenOpMetEANs(ctx context.Context, eans []int) ([]models.Product, error)
	SetHTTPClient(client *http.Client)
}

type ICategorieenAILijstMaker interface {
	MaakPassendeCategorieënLijst(ctx context.Context, behoeften string, availableCategories []models.Category) ([]int, error)
}

type IOplossingenAILijstMaker interface {
	MaakRelevanteTags(ctx context.Context, behoeften string, allAvailableTags []string) ([]string, error)
}
