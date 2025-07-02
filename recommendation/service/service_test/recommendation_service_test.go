package service_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"recommendation/model"
	"recommendation/service"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockAanbevelingsOpslag struct {
	mock.Mock
}

func (m *MockAanbevelingsOpslag) SlaAanbevelingOp(ctx context.Context, rec *model.Aanbeveling) error {
	args := m.Called(ctx, rec)
	if rec.ID == 0 {
		rec.ID = 1
	}
	return args.Error(0)
}

func (m *MockAanbevelingsOpslag) HaalAanbevelingOpMetCliëntID(ctx context.Context, clientID string) (*model.Aanbeveling, error) {
	args := m.Called(ctx, clientID)
	rec, ok := args.Get(0).(*model.Aanbeveling)
	if !ok && args.Get(0) != nil {
		return nil, args.Error(1)
	}
	return rec, args.Error(1)
}

func (m *MockAanbevelingsOpslag) WerkAanbevelingBij(ctx context.Context, rec *model.Aanbeveling) error {
	args := m.Called(ctx, rec)
	return args.Error(0)
}

func (m *MockAanbevelingsOpslag) MaakPassendeCategorieënLijstDB(ctx context.Context, lijst *model.PassendeCategorieënLijst) error {
	args := m.Called(ctx, lijst)
	if lijst.ID == 0 {
		lijst.ID = 100
	}
	return args.Error(0)
}

func (m *MockAanbevelingsOpslag) WerkPassendeCategorieënLijstBijDB(ctx context.Context, lijst *model.PassendeCategorieënLijst) error {
	args := m.Called(ctx, lijst)
	return args.Error(0)
}

func (m *MockAanbevelingsOpslag) HaalPassendeCategorieënLijstOpMetID(ctx context.Context, id uint) (*model.PassendeCategorieënLijst, error) {
	args := m.Called(ctx, id)
	lijst, ok := args.Get(0).(*model.PassendeCategorieënLijst)
	if !ok && args.Get(0) != nil {
		return nil, args.Error(1)
	}
	return lijst, args.Error(1)
}

func (m *MockAanbevelingsOpslag) MaakOplossingenLijstDB(ctx context.Context, lijst *model.OplossingenLijst) error {
	args := m.Called(ctx, lijst)
	if lijst.ID == 0 {
		lijst.ID = 200
	}
	return args.Error(0)
}

func (m *MockAanbevelingsOpslag) WerkOplossingenLijstBijDB(ctx context.Context, lijst *model.OplossingenLijst) error {
	args := m.Called(ctx, lijst)
	return args.Error(0)
}

func (m *MockAanbevelingsOpslag) HaalOplossingenLijstOpMetID(ctx context.Context, id uint) (*model.OplossingenLijst, error) {
	args := m.Called(ctx, id)
	lijst, ok := args.Get(0).(*model.OplossingenLijst)
	if !ok && args.Get(0) != nil {
		return nil, args.Error(1)
	}
	return lijst, args.Error(1)
}

type MockAICategorieenLijstMaker struct {
	mock.Mock
}

func (m *MockAICategorieenLijstMaker) MaakPassendeCategorieënLijst(ctx context.Context, behoeften string, availableCategories []model.Category) ([]int, error) {
	args := m.Called(ctx, behoeften, availableCategories)
	return args.Get(0).([]int), args.Error(1)
}

type MockAIOplossingenLijstMaker struct {
	mock.Mock
}

func (m *MockAIOplossingenLijstMaker) MaakRelevanteTags(ctx context.Context, behoeften string, allAvailableTags []string) ([]string, error) {
	args := m.Called(ctx, behoeften, allAvailableTags)
	return args.Get(0).([]string), args.Error(1)
}

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func createHttpResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

func TestMaakPassendeCategorieënLijst_NewRecommendation(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}

	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	patientID := "test-patient-1"
	budget := 100.0
	behoeften := "Ik zoek technologie voor mijn slaapproblemen."

	availableCategories := []model.Category{
		{ID: 1, Naam: "Slaaptechnologie"},
		{ID: 2, Naam: "Mobiliteit"},
	}
	selectedCategoryIDs := []int{1}

	mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"Slaaptechnologie","price_range":100},{"id":2,"naam":"Mobiliteit","price_range":200}]`), nil,
	).Once()

	mockCategoryAI.On("MaakPassendeCategorieënLijst", ctx, behoeften, availableCategories).Return(selectedCategoryIDs, nil).Once()

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, patientID).Return(nil, gorm.ErrRecordNotFound).Once()

	mockRepo.On("MaakPassendeCategorieënLijstDB", ctx, mock.AnythingOfType("*model.PassendeCategorieënLijst")).Return(nil).Once()

	mockRepo.On("SlaAanbevelingOp", ctx, mock.AnythingOfType("*model.Aanbeveling")).Return(nil).Once()

	mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"Slaaptechnologie"}]`), nil,
	).Once()

	result, err := svc.MaakPassendeCategorieënLijst(ctx, patientID, budget, behoeften)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Categories, 1)
	assert.Equal(t, "Slaaptechnologie", result.Categories[0].Naam)
	assert.Equal(t, model.ConvertIntSliceToPQInt64Array(selectedCategoryIDs), result.CategoryIDs)

	mockRepo.AssertExpectations(t)
	mockCategoryAI.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestMaakPassendeCategorieënLijst_UpdateExistingRecommendation(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	patientID := "test-patient-2"
	budget := 150.0
	behoeften := "Ik wil mijn huis slimmer maken."

	existingCategoryID := uint(50)
	existingRec := &model.Aanbeveling{
		Model:                 gorm.Model{ID: 10, CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now().Add(-24 * time.Hour)},
		ClientID:              patientID,
		Versie:                1,
		AanmaakDatum:          time.Now().Add(-24 * time.Hour),
		PassendeCategorieënID: &existingCategoryID,
		OplossingenLijstID:    nil,
	}
	existingCategoryList := &model.PassendeCategorieënLijst{
		Model:       gorm.Model{ID: existingCategoryID},
		CategoryIDs: model.ConvertIntSliceToPQInt64Array([]int{3, 4}),
	}

	availableCategories := []model.Category{
		{ID: 3, Naam: "Smart Home"},
		{ID: 4, Naam: "Energiebeheer"},
		{ID: 5, Naam: "Veiligheid"},
	}
	selectedCategoryIDs := []int{3, 5}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://product-service-url/categorieen?budget=150.00"
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":3,"naam":"Smart Home","price_range":100},{"id":4,"naam":"Energiebeheer","price_range":150},{"id":5,"naam":"Veiligheid","price_range":200}]`), nil,
	).Once()

	mockCategoryAI.On("MaakPassendeCategorieënLijst", ctx, behoeften, availableCategories).Return(selectedCategoryIDs, nil).Once()

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, patientID).Return(existingRec, nil).Once()

	mockRepo.On("HaalPassendeCategorieënLijstOpMetID", ctx, existingCategoryID).Return(existingCategoryList, nil).Once()

	mockRepo.On("WerkPassendeCategorieënLijstBijDB", ctx, mock.AnythingOfType("*model.PassendeCategorieënLijst")).Return(nil).Once()

	mockRepo.On("WerkAanbevelingBij", ctx, mock.AnythingOfType("*model.Aanbeveling")).Return(nil).Once()

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "categorieen?ids=") &&
			strings.Contains(req.URL.Query().Get("ids"), strconv.Itoa(selectedCategoryIDs[0])) &&
			strings.Contains(req.URL.Query().Get("ids"), strconv.Itoa(selectedCategoryIDs[1]))
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":3,"naam":"Smart Home"},{"id":5,"naam":"Veiligheid"}]`), nil,
	).Once()

	result, err := svc.MaakPassendeCategorieënLijst(ctx, patientID, budget, behoeften)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Categories, 2)
	assert.Equal(t, "Smart Home", result.Categories[0].Naam)
	assert.Equal(t, "Veiligheid", result.Categories[1].Naam)
	assert.Equal(t, model.ConvertIntSliceToPQInt64Array(selectedCategoryIDs), result.CategoryIDs)
	assert.Equal(t, existingCategoryID, result.ID)

	mockRepo.AssertExpectations(t)
	mockCategoryAI.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestMaakPassendeCategorieënLijst_ExistingRecMissingCategoryList(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	patientID := "test-patient-3"
	budget := 200.0
	behoeften := "Ik wil een gezonder leven leiden."

	existingCategoryID := uint(99)
	existingRec := &model.Aanbeveling{
		Model:                 gorm.Model{ID: 11, CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now().Add(-48 * time.Hour)},
		ClientID:              patientID,
		Versie:                1,
		AanmaakDatum:          time.Now().Add(-48 * time.Hour),
		PassendeCategorieënID: &existingCategoryID,
		OplossingenLijstID:    nil,
	}

	availableCategories := []model.Category{
		{ID: 6, Naam: "Voeding"},
		{ID: 7, Naam: "Fitness"},
	}
	selectedCategoryIDs := []int{6}

	mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(
		createHttpResponse(http.StatusOK, `[{"id":6,"naam":"Voeding","price_range":100},{"id":7,"naam":"Fitness","price_range":200}]`), nil,
	).Once()

	mockCategoryAI.On("MaakPassendeCategorieënLijst", ctx, behoeften, availableCategories).Return(selectedCategoryIDs, nil).Once()

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, patientID).Return(existingRec, nil).Once()

	mockRepo.On("HaalPassendeCategorieënLijstOpMetID", ctx, existingCategoryID).Return(nil, gorm.ErrRecordNotFound).Once()

	mockRepo.On("MaakPassendeCategorieënLijstDB", ctx, mock.AnythingOfType("*model.PassendeCategorieënLijst")).Return(nil).Once()

	mockRepo.On("WerkAanbevelingBij", ctx, mock.AnythingOfType("*model.Aanbeveling")).Return(nil).Once()

	mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(
		createHttpResponse(http.StatusOK, `[{"id":6,"naam":"Voeding"}]`), nil,
	).Once()

	result, err := svc.MaakPassendeCategorieënLijst(ctx, patientID, budget, behoeften)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Categories, 1)
	assert.Equal(t, "Voeding", result.Categories[0].Naam)
	assert.Equal(t, model.ConvertIntSliceToPQInt64Array(selectedCategoryIDs), result.CategoryIDs)
	assert.NotEqual(t, existingCategoryID, result.ID)

	mockRepo.AssertExpectations(t)
	mockCategoryAI.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestMaakOplossingenLijst_NewRecommendation(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "test-client-1"
	budget := 500.0
	behoeften := "Ik heb een nieuwe telefoon nodig."
	categoryID := 1

	allAvailableTags := []string{"smartphone", "tablet", "laptop"}
	selectedTags := []string{"smartphone"}
	productEANs := []int64{1234567890123}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://product-service-url/categorieen/tags?categorieID=1"
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"smartphone"},{"id":2,"naam":"tablet"},{"id":3,"naam":"laptop"}]`), nil,
	).Once()

	mockSolutionsAI.On("MaakRelevanteTags", ctx, behoeften, allAvailableTags).Return(selectedTags, nil).Once()

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, clientID).Return(nil, gorm.ErrRecordNotFound).Once()

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "product?tags=") &&
			strings.Contains(req.URL.Query().Get("tags"), selectedTags[0]) &&
			strings.Contains(req.URL.Query().Get("budget"), strconv.FormatFloat(budget, 'f', 2, 64)) &&
			strings.Contains(req.URL.Query().Get("categorieen"), strconv.Itoa(categoryID))
	})).Return(
		createHttpResponse(http.StatusOK, `[{"ean":1234567890123,"naam":"SuperPhone X","omschrijving":"De nieuwste smartphone","price":499.99}]`), nil,
	).Once()

	mockRepo.On("MaakOplossingenLijstDB", ctx, mock.AnythingOfType("*model.OplossingenLijst")).Return(nil).Once()

	mockRepo.On("SlaAanbevelingOp", ctx, mock.AnythingOfType("*model.Aanbeveling")).Return(nil).Once()

	result, err := svc.MaakOplossingenLijst(ctx, clientID, budget, behoeften, &categoryID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Products, 1)
	assert.Equal(t, "SuperPhone X", result.Products[0].Naam)
	assert.Equal(t, model.ConvertInt64SliceToPQInt64Array(productEANs), result.ProductEANs)

	mockRepo.AssertExpectations(t)
	mockSolutionsAI.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestMaakOplossingenLijst_UpdateExistingRecommendation(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "test-client-2"
	budget := 750.0
	behoeften := "Ik wil een laptop voor werk en studie."
	categoryListID := uint(101)
	chosenCategoryId := 2

	existingSolutionsListID := uint(20)
	existingRec := &model.Aanbeveling{
		Model:                 gorm.Model{ID: 12, CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now().Add(-24 * time.Hour)},
		ClientID:              clientID,
		Versie:                1,
		AanmaakDatum:          time.Now().Add(-24 * time.Hour),
		PassendeCategorieënID: &categoryListID,
		OplossingenLijstID:    &existingSolutionsListID,
	}
	existingSolutionsList := &model.OplossingenLijst{
		Model:       gorm.Model{ID: existingSolutionsListID},
		ProductEANs: model.ConvertInt64SliceToPQInt64Array([]int64{9876543210987}),
	}

	allAvailableTags := []string{"laptop", "desktop", "monitor"}
	selectedTags := []string{"laptop"}

	productEANs := []int64{1122334455667}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://product-service-url/categorieen/tags?categorieID=2"
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":4,"naam":"laptop"},{"id":5,"naam":"desktop"},{"id":6,"naam":"monitor"}]`), nil,
	).Once()

	mockSolutionsAI.On("MaakRelevanteTags", ctx, behoeften, allAvailableTags).Return(selectedTags, nil).Once()

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, clientID).Return(existingRec, nil).Once()

	mockRepo.On("HaalOplossingenLijstOpMetID", ctx, existingSolutionsListID).Return(existingSolutionsList, nil).Once()

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "product?tags=") &&
			strings.Contains(req.URL.Query().Get("tags"), selectedTags[0]) &&
			strings.Contains(req.URL.Query().Get("budget"), strconv.FormatFloat(budget, 'f', 2, 64)) &&
			strings.Contains(req.URL.Query().Get("categorieen"), strconv.Itoa(chosenCategoryId))
	})).Return(
		createHttpResponse(http.StatusOK, `[{"ean":1122334455667,"naam":"UltraBook Pro","omschrijving":"Lichtgewicht en krachtig","price":749.99}]`), nil,
	).Once()

	mockRepo.On("WerkOplossingenLijstBijDB", ctx, mock.AnythingOfType("*model.OplossingenLijst")).Return(nil).Once()

	mockRepo.On("WerkAanbevelingBij", ctx, mock.AnythingOfType("*model.Aanbeveling")).Return(nil).Once()

	result, err := svc.MaakOplossingenLijst(ctx, clientID, budget, behoeften, &chosenCategoryId)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Products, 1)
	assert.Equal(t, "UltraBook Pro", result.Products[0].Naam)
	assert.Equal(t, model.ConvertInt64SliceToPQInt64Array(productEANs), result.ProductEANs)
	assert.Equal(t, existingSolutionsListID, result.ID)

	mockRepo.AssertExpectations(t)
	mockSolutionsAI.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestMaakOplossingenLijst_ExistingRecMissingSolutionsList(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "test-client-3"
	budget := 300.0
	behoeften := "Ik zoek een smart TV."
	categoryID := uint(102)
	chosenCategoryId := 5

	existingSolutionsListID := uint(999)
	existingRec := &model.Aanbeveling{
		Model:                 gorm.Model{ID: 13, CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now().Add(-48 * time.Hour)},
		ClientID:              clientID,
		Versie:                1,
		AanmaakDatum:          time.Now().Add(-48 * time.Hour),
		PassendeCategorieënID: &categoryID,
		OplossingenLijstID:    &existingSolutionsListID,
	}

	allAvailableTags := []string{"tv", "audio", "gaming"}
	selectedTags := []string{"tv"}

	productEANs := []int64{9988776655443}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://product-service-url/categorieen/tags?categorieID=5"
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":7,"naam":"tv"},{"id":8,"naam":"audio"},{"id":9,"naam":"gaming"}]`), nil,
	).Once()

	mockSolutionsAI.On("MaakRelevanteTags", ctx, behoeften, allAvailableTags).Return(selectedTags, nil).Once()

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, clientID).Return(existingRec, nil).Once()

	mockRepo.On("HaalOplossingenLijstOpMetID", ctx, existingSolutionsListID).Return(nil, gorm.ErrRecordNotFound).Once()

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "product?tags=") &&
			strings.Contains(req.URL.Query().Get("tags"), selectedTags[0]) &&
			strings.Contains(req.URL.Query().Get("budget"), strconv.FormatFloat(budget, 'f', 2, 64)) &&
			strings.Contains(req.URL.Query().Get("categorieen"), strconv.Itoa(chosenCategoryId))
	})).Return(
		createHttpResponse(http.StatusOK, `[{"ean":9988776655443,"naam":"Smart TV X","omschrijving":"Groot scherm, veel apps","price":299.99}]`), nil,
	).Once()

	mockRepo.On("MaakOplossingenLijstDB", ctx, mock.AnythingOfType("*model.OplossingenLijst")).Return(nil).Once()

	mockRepo.On("WerkAanbevelingBij", ctx, mock.AnythingOfType("*model.Aanbeveling")).Return(nil).Once()

	result, err := svc.MaakOplossingenLijst(ctx, clientID, budget, behoeften, &chosenCategoryId)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Products, 1)
	assert.Equal(t, "Smart TV X", result.Products[0].Naam)
	assert.Equal(t, model.ConvertInt64SliceToPQInt64Array(productEANs), result.ProductEANs)
	assert.NotEqual(t, existingSolutionsListID, result.ID)

	mockRepo.AssertExpectations(t)
	mockSolutionsAI.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestHaalPassendeCategorieënLijstOp_Success(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	patientID := "get-patient-1"
	categoryListID := uint(300)
	expectedCategoryIDs := []int{10, 11}

	mockRec := &model.Aanbeveling{
		Model:                 gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		ClientID:              patientID,
		Versie:                1,
		AanmaakDatum:          time.Now(),
		PassendeCategorieënID: &categoryListID,
	}
	mockCategoryList := &model.PassendeCategorieënLijst{
		Model:       gorm.Model{ID: categoryListID},
		CategoryIDs: model.ConvertIntSliceToPQInt64Array(expectedCategoryIDs),
	}

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, patientID).Return(mockRec, nil).Once()
	mockRepo.On("HaalPassendeCategorieënLijstOpMetID", ctx, categoryListID).Return(mockCategoryList, nil).Once()
	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "categorieen?ids=10%2C11")
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":10,"naam":"Gezondheid"},{"id":11,"naam":"Welzijn"}]`), nil,
	).Once()

	result, err := svc.HaalPassendeCategorieënLijstOp(ctx, patientID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, categoryListID, result.ID)
	assert.Equal(t, model.ConvertIntSliceToPQInt64Array(expectedCategoryIDs), result.CategoryIDs)
	assert.Len(t, result.Categories, 2)
	assert.Equal(t, "Gezondheid", result.Categories[0].Naam)
	assert.Equal(t, "Welzijn", result.Categories[1].Naam)

	mockRepo.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestHaalPassendeCategorieënLijstOp_NoRecommendationFound(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	patientID := "non-existent-patient"

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, patientID).Return(nil, gorm.ErrRecordNotFound).Once()

	result, err := svc.HaalPassendeCategorieënLijstOp(ctx, patientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fout bij ophalen aanbeveling voor cliënt")
	assert.Contains(t, err.Error(), "record not found")
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockRT.AssertNotCalled(t, "RoundTrip", mock.Anything)
}

func TestHaalOplossingenLijstOp_Success(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "get-client-1"
	solutionsListID := uint(400)
	expectedEANs := []int64{1000000000001, 1000000000002}

	mockRec := &model.Aanbeveling{
		Model:              gorm.Model{ID: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		ClientID:           clientID,
		Versie:             1,
		AanmaakDatum:       time.Now(),
		OplossingenLijstID: &solutionsListID,
	}
	mockSolutionsList := &model.OplossingenLijst{
		Model:       gorm.Model{ID: solutionsListID},
		ProductEANs: model.ConvertInt64SliceToPQInt64Array(expectedEANs),
	}

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, clientID).Return(mockRec, nil).Once()
	mockRepo.On("HaalOplossingenLijstOpMetID", ctx, solutionsListID).Return(mockSolutionsList, nil).Once()

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "product?eans=") &&
			strings.Contains(req.URL.Query().Get("eans"), strconv.FormatInt(expectedEANs[0], 10)) &&
			strings.Contains(req.URL.Query().Get("eans"), strconv.FormatInt(expectedEANs[1], 10))
	})).Return(
		createHttpResponse(http.StatusOK, `[{"ean":1000000000001,"naam":"Product A","omschrijving":"Beschrijving A","price":10.0},{"ean":1000000000002,"naam":"Product B","omschrijving":"Beschrijving B","price":20.0}]`), nil,
	).Once()

	result, err := svc.HaalOplossingenLijstOp(ctx, clientID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, solutionsListID, result.ID)
	assert.Equal(t, model.ConvertInt64SliceToPQInt64Array(expectedEANs), result.ProductEANs)
	assert.Len(t, result.Products, 2)
	assert.Equal(t, "Product A", result.Products[0].Naam)
	assert.Equal(t, "Product B", result.Products[1].Naam)

	mockRepo.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestHaalOplossingenLijstOp_NoRecommendationFound(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "non-existent-client"

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, clientID).Return(nil, gorm.ErrRecordNotFound).Once()

	result, err := svc.HaalOplossingenLijstOp(ctx, clientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fout bij ophalen aanbeveling voor cliënt")
	assert.Contains(t, err.Error(), "record not found")
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockRT.AssertNotCalled(t, "RoundTrip", mock.Anything)
}

func TestHaalOplossingenLijstOp_RecFoundNoSolutionsListID(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "client-no-solutions-id"

	mockRec := &model.Aanbeveling{
		Model:              gorm.Model{ID: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		ClientID:           clientID,
		Versie:             1,
		AanmaakDatum:       time.Now(),
		OplossingenLijstID: nil,
	}

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, clientID).Return(mockRec, nil).Once()

	result, err := svc.HaalOplossingenLijstOp(ctx, clientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "geen oplossingenlijst gevonden voor cliënt")
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockRT.AssertNotCalled(t, "RoundTrip", mock.Anything)
}

func TestHaalOplossingenLijstOp_SolutionsListNotFoundInDB(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "client-missing-solutions-list"
	solutionsListID := uint(500)

	mockRec := &model.Aanbeveling{
		Model:              gorm.Model{ID: 4, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		ClientID:           clientID,
		Versie:             1,
		AanmaakDatum:       time.Now(),
		OplossingenLijstID: &solutionsListID,
	}

	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, clientID).Return(mockRec, nil).Once()
	mockRepo.On("HaalOplossingenLijstOpMetID", ctx, solutionsListID).Return(nil, gorm.ErrRecordNotFound).Once()

	result, err := svc.HaalOplossingenLijstOp(ctx, clientID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oplossingenlijst met ID 500 niet gevonden")
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockRT.AssertNotCalled(t, "RoundTrip", mock.Anything)
}

func TestHaalAlleTagsOp_Success(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	categoryID := 123
	expectedTags := []string{"tag1", "tag2"}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://product-service-url/categorieen/tags?categorieID=123"
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"tag1"},{"id":2,"naam":"tag2"}]`), nil,
	).Once()

	tags, err := svc.HaalAlleTagsOp(ctx, &categoryID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTags, tags)
	mockRT.AssertExpectations(t)
}

func TestHaalAlleTagsOp_ProductServiceError(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()

	mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(
		createHttpResponse(http.StatusInternalServerError, `{"error":"internal server error"}`), nil,
	).Once()

	tags, err := svc.HaalAlleTagsOp(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product service retourneerde foutstatus 500")
	assert.Nil(t, tags)
	mockRT.AssertExpectations(t)
}

func TestHaalCategorieënOp_Success(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	budget := 500.0
	expectedCategories := []model.Category{
		{ID: 1, Naam: "Elektronica"},
		{ID: 2, Naam: "Huishoudelijk"},
	}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://product-service-url/categorieen?budget=500.00"
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"Elektronica","price_range":500},{"id":2,"naam":"Huishoudelijk","price_range":300}]`), nil,
	).Once()

	categories, err := svc.HaalCategorieënOp(ctx, budget)

	assert.NoError(t, err)
	assert.Equal(t, expectedCategories, categories)
	mockRT.AssertExpectations(t)
}

func TestHaalProductenOp_Success(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	tags := []string{"gaming", "high-performance"}
	budget := 1500.0
	categorieen := []int{1, 2}
	expectedProducts := []model.Product{
		{EAN: 1000000000005, Naam: "Gaming Laptop", Omschrijving: "Krachtige gaming laptop"},
	}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "product?tags=gaming%2Chigh-performance&budget=1500.00&categorieen=1%2C2")
	})).Return(
		createHttpResponse(http.StatusOK, `[{"ean":1000000000005,"naam":"Gaming Laptop","omschrijving":"Krachtige gaming laptop"}]`), nil,
	).Once()

	products, err := svc.HaalProductenOp(ctx, tags, budget, categorieen)

	assert.NoError(t, err)
	assert.Equal(t, expectedProducts, products)
	mockRT.AssertExpectations(t)
}

func TestHaalProductenOpMetEANs_Success(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	eans := []int{1000000000001, 1000000000002}
	expectedEANs := []int64{1000000000001, 1000000000002}
	expectedProducts := []model.Product{
		{EAN: 1000000000001, Naam: "Product A", Omschrijving: "Beschrijving A"},
		{EAN: 1000000000002, Naam: "Product B", Omschrijving: "Beschrijving B"},
	}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "product?eans=") &&
			strings.Contains(req.URL.Query().Get("eans"), strconv.FormatInt(expectedEANs[0], 10)) &&
			strings.Contains(req.URL.Query().Get("eans"), strconv.FormatInt(expectedEANs[1], 10))
	})).Return(
		createHttpResponse(http.StatusOK, `[{"ean":1000000000001,"naam":"Product A","omschrijving":"Beschrijving A","price":10.0},{"ean":1000000000002,"naam":"Product B","omschrijving":"Beschrijving B","price":20.0}]`), nil,
	).Once()

	products, err := svc.HaalProductenOpMetEANs(ctx, eans)

	assert.NoError(t, err)
	assert.Equal(t, expectedProducts, products)
	mockRT.AssertExpectations(t)
}

func TestHaalProductenOpMetEANs_EmptyEANs(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	eans := []int{}

	products, err := svc.HaalProductenOpMetEANs(ctx, eans)

	assert.NoError(t, err)
	assert.Empty(t, products)
	mockRT.AssertNotCalled(t, "RoundTrip", mock.Anything)
}

func TestHaalCategorieenOpMetIDs_Success(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	ids := []int{1, 2}
	expectedCategories := []model.Category{
		{ID: 1, Naam: "Categorie A"},
		{ID: 2, Naam: "Categorie B"},
	}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "categorieen?ids=1%2C2")
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"Categorie A"},{"id":2,"naam":"Categorie B"}]`), nil,
	).Once()

	categories, err := svc.HaalCategorieenOpMetIDs(ctx, ids)

	assert.NoError(t, err)
	assert.Equal(t, expectedCategories, categories)
	mockRT.AssertExpectations(t)
}

func TestHaalCategorieenOpMetIDs_EmptyIDs(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	ids := []int{}

	categories, err := svc.HaalCategorieenOpMetIDs(ctx, ids)

	assert.NoError(t, err)
	assert.Empty(t, categories)
	mockRT.AssertNotCalled(t, "RoundTrip", mock.Anything)
}

func TestMaakPassendeCategorieënLijst_ProductServiceError(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	patientID := "test-patient-error"
	budget := 100.0
	behoeften := "Test behoeften."

	mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(
		createHttpResponse(http.StatusInternalServerError, `{"error":"internal server error"}`), nil,
	).Once()

	result, err := svc.MaakPassendeCategorieënLijst(ctx, patientID, budget, behoeften)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fout bij ophalen budget-passende categorieën van product service")
	assert.Nil(t, result)
	mockRT.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "HaalAanbevelingOpMetCliëntID", mock.Anything, mock.Anything)
	mockCategoryAI.AssertNotCalled(t, "MaakPassendeCategorieënLijst", mock.Anything, mock.Anything, mock.Anything)
}

func TestMaakPassendeCategorieënLijst_AIGenerationError(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	patientID := "test-patient-ai-error"
	budget := 100.0
	behoeften := "Test behoeften."

	availableCategories := []model.Category{
		{ID: 1, Naam: "Slaaptechnologie"},
	}

	mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"Slaaptechnologie","price_range":100}]`), nil,
	).Once()
	mockCategoryAI.On("MaakPassendeCategorieënLijst", ctx, behoeften, availableCategories).Return([]int{}, errors.New("AI error")).Once()

	result, err := svc.MaakPassendeCategorieënLijst(ctx, patientID, budget, behoeften)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fout bij het maken van passende categorieënlijst door AI")
	assert.Nil(t, result)
	mockRT.AssertExpectations(t)
	mockCategoryAI.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "HaalAanbevelingOpMetCliëntID", mock.Anything, mock.Anything)
}

func TestMaakOplossingenLijst_AIGenerationError(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "test-client-ai-error-sol"
	budget := 500.0
	behoeften := "Test behoeften."
	categoryID := 1

	allAvailableTags := []string{"tag1", "tag2"}

	mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"tag1"},{"id":2,"naam":"tag2"}]`), nil,
	).Once()
	mockSolutionsAI.On("MaakRelevanteTags", ctx, behoeften, allAvailableTags).Return([]string{}, errors.New("AI solutions error")).Once()

	result, err := svc.MaakOplossingenLijst(ctx, clientID, budget, behoeften, &categoryID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fout bij het maken van relevante tags door AI")
	assert.Nil(t, result)
	mockRT.AssertExpectations(t)
	mockSolutionsAI.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "HaalAanbevelingOpMetCliëntID", mock.Anything, mock.Anything)
}

func TestMaakOplossingenLijst_NoExistingCategoryList(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "client-no-cat-list"
	budget := 100.0
	behoeften := "Test behoeften."

	mockRT.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"tag1"}]`), nil,
	).Once()
	mockSolutionsAI.On("MaakRelevanteTags", ctx, behoeften, mock.Anything).Return([]string{"tag1"}, nil).Once()
	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, clientID).Return(&model.Aanbeveling{ClientID: clientID, PassendeCategorieënID: nil}, nil).Once()

	result, err := svc.MaakOplossingenLijst(ctx, clientID, budget, behoeften, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "geen bestaande passende categorieënlijst gevonden voor cliënt")
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
	mockSolutionsAI.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestMaakOplossingenLijst_RepositoryErrorOnSave(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	clientID := "client-repo-error"
	budget := 200.0
	behoeften := "Test behoeften."
	categoryID := uint(103)
	chosenCategoryId := 9

	allAvailableTags := []string{"tag1"}
	selectedTags := []string{"tag1"}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://product-service-url/categorieen/tags?categorieID=9"
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":1,"naam":"tag1"}]`), nil,
	).Once()
	mockSolutionsAI.On("MaakRelevanteTags", ctx, behoeften, allAvailableTags).Return(selectedTags, nil).Once()
	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, clientID).Return(&model.Aanbeveling{ClientID: clientID, PassendeCategorieënID: &categoryID}, nil).Once()
	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return strings.Contains(req.URL.String(), "product?tags=") &&
			strings.Contains(req.URL.Query().Get("tags"), selectedTags[0]) &&
			strings.Contains(req.URL.Query().Get("budget"), strconv.FormatFloat(budget, 'f', 2, 64)) &&
			strings.Contains(req.URL.Query().Get("categorieen"), strconv.Itoa(chosenCategoryId))
	})).Return(
		createHttpResponse(http.StatusOK, `[{"ean":1,"naam":"Test Product","omschrijving":"Desc","price":10.0}]`), nil,
	).Once()
	mockRepo.On("MaakOplossingenLijstDB", ctx, mock.AnythingOfType("*model.OplossingenLijst")).Return(errors.New("DB save error")).Once()

	result, err := svc.MaakOplossingenLijst(ctx, clientID, budget, behoeften, &chosenCategoryId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fout bij opslaan nieuwe oplossingenlijst")
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
	mockSolutionsAI.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}

func TestMaakPassendeCategorieënLijst_RepositoryErrorOnUpdate(t *testing.T) {
	mockRepo := new(MockAanbevelingsOpslag)
	mockCategoryAI := new(MockAICategorieenLijstMaker)
	mockSolutionsAI := new(MockAIOplossingenLijstMaker)
	mockRT := new(MockRoundTripper)

	testClient := &http.Client{Transport: mockRT}
	svc := service.NewAanbevelingHelpers(mockRepo, mockCategoryAI, mockSolutionsAI, "http://product-service-url")
	svc.(*service.AanbevelingHelpers).SetHTTPClient(testClient)

	ctx := context.Background()
	patientID := "test-patient-update-error"
	budget := 100.0
	behoeften := "Test behoeften."

	existingCategoryID := uint(50)
	existingRec := &model.Aanbeveling{
		Model:                 gorm.Model{ID: 10, CreatedAt: time.Now().Add(-24 * time.Hour), UpdatedAt: time.Now().Add(-24 * time.Hour)},
		ClientID:              patientID,
		Versie:                1,
		AanmaakDatum:          time.Now().Add(-24 * time.Hour),
		PassendeCategorieënID: &existingCategoryID,
		OplossingenLijstID:    nil,
	}
	existingCategoryList := &model.PassendeCategorieënLijst{
		Model:       gorm.Model{ID: existingCategoryID},
		CategoryIDs: model.ConvertIntSliceToPQInt64Array([]int{3, 4}),
	}

	availableCategories := []model.Category{
		{ID: 3, Naam: "Smart Home"},
	}

	mockRT.On("RoundTrip", mock.MatchedBy(func(req *http.Request) bool {
		return req.URL.String() == "http://product-service-url/categorieen?budget=100.00"
	})).Return(
		createHttpResponse(http.StatusOK, `[{"id":3,"naam":"Smart Home","price_range":100}]`), nil,
	).Once()
	mockCategoryAI.On("MaakPassendeCategorieënLijst", ctx, behoeften, availableCategories).Return([]int{3}, nil).Once()
	mockRepo.On("HaalAanbevelingOpMetCliëntID", ctx, patientID).Return(existingRec, nil).Once()
	mockRepo.On("HaalPassendeCategorieënLijstOpMetID", ctx, existingCategoryID).Return(existingCategoryList, nil).Once()
	mockRepo.On("WerkPassendeCategorieënLijstBijDB", ctx, mock.AnythingOfType("*model.PassendeCategorieënLijst")).Return(nil).Once()
	mockRepo.On("WerkAanbevelingBij", ctx, mock.AnythingOfType("*model.Aanbeveling")).Return(errors.New("DB update error")).Once()

	result, err := svc.MaakPassendeCategorieënLijst(ctx, patientID, budget, behoeften)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fout bij bijwerken aanbeveling voor cliënt")
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
	mockCategoryAI.AssertExpectations(t)
	mockRT.AssertExpectations(t)
}
