package product_service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"product/handlers"
	models "product/model"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Minute),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("Failed to terminate PostgreSQL container: %v", err)
		}
	}()

	host, err := pgContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}
	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("Failed to get container mapped port: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpassword dbname=testdb sslmode=disable", host, port.Port())
	log.Printf("Connecting to PostgreSQL test database at DSN: %s", dsn)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}

	sqlDB, err := testDB.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying SQL DB: %v", err)
	}
	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}

	log.Println("PostgreSQL test database connection established.")

	exitCode := m.Run()

	if err := sqlDB.Close(); err != nil {
		log.Printf("Error closing PostgreSQL database connection: %v", err)
	}

	os.Exit(exitCode)
}

func setupTestDB(t *testing.T) *gorm.DB {
	err := testDB.Migrator().DropTable(
		&models.Product{},
		&models.Categorie{},
		&models.Specificatie{},
		&models.Review{},
		&models.ProductAanbod{},
		&models.ProductType{},
		&models.Supplier{},
		&models.Tag{},
		"product_categories",
		"product_tags",
	)
	if err != nil {
		t.Fatalf("Failed to drop tables: %v", err)
	}

	err = testDB.AutoMigrate(
		&models.Product{},
		&models.Categorie{},
		&models.Specificatie{},
		&models.Review{},
		&models.ProductAanbod{},
		&models.ProductType{},
		&models.Supplier{},
		&models.Tag{},
	)
	if err != nil {
		t.Fatalf("Failed to auto-migrate models: %v", err)
	}

	handlers.InitHandlers(testDB)

	return testDB
}

func seedTestData(db *gorm.DB) {
	pt1 := models.ProductType{Naam: "Mobiliteitshulpmiddel", Omschrijving: "Hulp bij bewegen"}
	pt2 := models.ProductType{Naam: "Revalidatiehulpmiddel", Omschrijving: "Hulp bij herstel"}
	db.Create(&pt1)
	db.Create(&pt2)

	s1 := models.Supplier{Name: "Leverancier A"}
	s2 := models.Supplier{Name: "Leverancier B"}
	db.Create(&s1)
	db.Create(&s2)

	cat1 := models.Categorie{Naam: "Loopondersteuning", PriceRange: 1000}
	cat2 := models.Categorie{Naam: "Wielvoertuigen", PriceRange: 5000}
	cat3 := models.Categorie{Naam: "Badkamerhulpmiddelen", PriceRange: 800}
	db.Create(&cat1)
	db.Create(&cat2)
	db.Create(&cat3)

	tag1 := models.Tag{Naam: "lichtgewicht"}
	tag2 := models.Tag{Naam: "opvouwbaar"}
	tag3 := models.Tag{Naam: "comfort"}
	tag4 := models.Tag{Naam: "elektrisch"}
	db.Create(&tag1)
	db.Create(&tag2)
	db.Create(&tag3)
	db.Create(&tag4)

	p1 := models.Product{
		EAN:           1000000000000,
		SKU:           "SKU001",
		Naam:          "Rollator Lichtgewicht",
		Omschrijving:  "Lichtgewicht opvouwbare rollator.",
		Merk:          "MerkX",
		Afbeeldingen:  []string{"img1.jpg"},
		Gewicht:       7.5,
		ProductTypeID: pt1.ID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	p2 := models.Product{
		EAN:           1000000000001,
		SKU:           "SKU002",
		Naam:          "Scootmobiel Comfort",
		Omschrijving:  "Comfortabele elektrische scootmobiel.",
		Merk:          "MerkY",
		Afbeeldingen:  []string{"img2.jpg"},
		Gewicht:       50.0,
		ProductTypeID: pt1.ID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	p3 := models.Product{
		EAN:           1000000000002,
		SKU:           "SKU003",
		Naam:          "Douchekruk Compact",
		Omschrijving:  "Compacte douchekruk voor kleine badkamers.",
		Merk:          "MerkZ",
		Afbeeldingen:  []string{"img3.jpg"},
		Gewicht:       2.0,
		ProductTypeID: pt2.ID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	db.Create(&p1)
	db.Create(&p2)
	db.Create(&p3)

	db.Model(&p1).Association("Categories").Append(&cat1, &cat2)
	db.Model(&p2).Association("Categories").Append(&cat2)
	db.Model(&p3).Association("Categories").Append(&cat3)

	db.Model(&p1).Association("Tags").Append(&tag1, &tag2)
	db.Model(&p2).Association("Tags").Append(&tag3, &tag4)
	db.Model(&p3).Association("Tags").Append(&tag1)

	pa1 := models.ProductAanbod{ProductEAN: p1.EAN, Prijs: 399, Voorraad: 10, LeverancierID: s1.ID}
	pa2 := models.ProductAanbod{ProductEAN: p2.EAN, Prijs: 2500, Voorraad: 5, LeverancierID: s2.ID}
	pa3 := models.ProductAanbod{ProductEAN: p3.EAN, Prijs: 79, Voorraad: 20, LeverancierID: s1.ID}
	pa4 := models.ProductAanbod{ProductEAN: p1.EAN, Prijs: 420, Voorraad: 5, LeverancierID: s2.ID}
	db.Create(&pa1)
	db.Create(&pa2)
	db.Create(&pa3)
	db.Create(&pa4)
}

func TestHaalProductLeveraarsOp(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(db)

	tests := []struct {
		name           string
		ean            string
		expectedStatus int
		expectedCount  int
		expectedNames  []string
	}{
		{
			name:           "Succesvol ophalen leveranciers voor EAN 1000000000000",
			ean:            "1000000000000",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedNames:  []string{"Leverancier A", "Leverancier B"},
		},
		{
			name:           "Geen leveranciers voor onbestaande EAN",
			ean:            "9999999999999",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedNames:  []string{},
		},
		{
			name:           "Ongeldige EAN parameter",
			ean:            "abc",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedNames:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/product/suppliers?ean="+tt.ean, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handlers.HaalProductLeveraarsOp(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if tt.expectedStatus != http.StatusOK {
				return
			}

			var suppliers []models.Supplier
			err = json.Unmarshal(rr.Body.Bytes(), &suppliers)
			if err != nil {
				t.Errorf("Could not unmarshal response: %v", err)
			}

			if len(suppliers) != tt.expectedCount {
				t.Errorf("handler returned unexpected number of suppliers: got %v want %v", len(suppliers), tt.expectedCount)
			}

			foundNames := make(map[string]bool)
			for _, s := range suppliers {
				foundNames[s.Name] = true
			}
			for _, name := range tt.expectedNames {
				if !foundNames[name] {
					t.Errorf("Expected supplier %s not found in response", name)
				}
			}
		})
	}
}

func TestHaalProductenOp(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(db)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		expectedEANs   []int
	}{
		{
			name:           "Alle producten ophalen (geen filters)",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			expectedEANs:   []int{1000000000000, 1000000000001, 1000000000002},
		},
		{
			name:           "Filter op budget (prijs <= 100)",
			queryParams:    "budget=100",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedEANs:   []int{1000000000002},
		},
		{
			name:           "Filter op bestaande tag (lichtgewicht)",
			queryParams:    "tags=lichtgewicht",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedEANs:   []int{1000000000000, 1000000000002},
		},
		{
			name:           "Filter op meerdere bestaande tags (lichtgewicht,opvouwbaar)",
			queryParams:    "tags=lichtgewicht,opvouwbaar",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedEANs:   []int{1000000000000},
		},
		{
			name:           "Filter op niet-bestaande tag",
			queryParams:    "tags=onbestaand",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedEANs:   []int{},
		},
		{
			name:           "Filter op EANs",
			queryParams:    "eans=1000000000000,1000000000002",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedEANs:   []int{1000000000000, 1000000000002},
		},
		{
			name:           "Filter op niet-bestaande EAN",
			queryParams:    "eans=9999999999999",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedEANs:   []int{},
		},
		{
			name:           "Combinatie filter: budget en tags",
			queryParams:    "budget=2000&tags=elektrisch",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedEANs:   []int{},
		},
		{
			name:           "Combinatie filter: budget en tags (matching)",
			queryParams:    "budget=3000&tags=elektrisch,comfort",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedEANs:   []int{1000000000001},
		},
		{
			name:           "Ongeldige budget-parameter",
			queryParams:    "budget=abc",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedEANs:   []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/product?"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handlers.HaalProductenOp(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if tt.expectedStatus != http.StatusOK {
				return
			}

			var products []models.Product
			err = json.Unmarshal(rr.Body.Bytes(), &products)
			if err != nil {
				t.Errorf("Could not unmarshal response: %v", err)
			}

			if len(products) != tt.expectedCount {

				t.Errorf("handler returned unexpected number of products: got %v want %v", products, tt.expectedCount)
				t.Errorf("\n\n\n\n\n handler returned unexpected number of products: got %v want %v", len(products), tt.expectedCount)
			}

			foundEANs := make(map[int]bool)
			for _, p := range products {
				foundEANs[p.EAN] = true
			}
			for _, ean := range tt.expectedEANs {
				if !foundEANs[ean] {
					t.Errorf("Expected EAN %d not found in response", ean)
				}
			}
		})
	}
}

func TestHaalCategorieenOp(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(db)

	getCategoryID := func(name string) uint {
		var cat models.Categorie
		if err := db.Where("naam = ?", name).First(&cat).Error; err != nil {
			t.Fatalf("Failed to find category %s for test setup: %v", name, err)
		}
		return cat.ID
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		expectedNames  []string
	}{
		{
			name:           "Alle categorieën ophalen (geen filters)",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			expectedNames:  []string{"Loopondersteuning", "Wielvoertuigen", "Badkamerhulpmiddelen"},
		},
		{
			name:           "Filter op budget (price_range <= 1000)",
			queryParams:    "budget=1000",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedNames:  []string{"Loopondersteuning", "Badkamerhulpmiddelen"},
		},
		{
			name:           "Filter op specifieke ID",
			queryParams:    "ids=" + strconv.Itoa(int(getCategoryID("Wielvoertuigen"))),
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedNames:  []string{"Wielvoertuigen"},
		},
		{
			name:           "Filter op meerdere ID's",
			queryParams:    fmt.Sprintf("ids=%d,%d", getCategoryID("Loopondersteuning"), getCategoryID("Badkamerhulpmiddelen")),
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedNames:  []string{"Loopondersteuning", "Badkamerhulpmiddelen"},
		},
		{
			name:           "Filter op niet-bestaande ID",
			queryParams:    "ids=999",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedNames:  []string{},
		},
		{
			name:           "Ongeldige budget-parameter",
			queryParams:    "budget=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedNames:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/categorieen?"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handlers.HaalCategorieenOp(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if tt.expectedStatus != http.StatusOK {
				return
			}

			var categories []models.Categorie
			err = json.Unmarshal(rr.Body.Bytes(), &categories)
			if err != nil {
				t.Errorf("Could not unmarshal response: %v", err)
			}

			if len(categories) != tt.expectedCount {
				t.Errorf("handler returned unexpected number of categories: got %v want %v", len(categories), tt.expectedCount)
			}

			foundNames := make(map[string]bool)
			for _, c := range categories {
				foundNames[c.Naam] = true
			}
			for _, name := range tt.expectedNames {
				if !foundNames[name] {
					t.Errorf("Expected category %s not found in response", name)
				}
			}
		})
	}
}

func TestPlaatsReview(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(db)

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Succesvol plaatsen van een review",
			method:         http.MethodPost,
			requestBody:    models.Review{ProductEAN: 1000000000000, Naam: "Test Reviewer", Score: 5, Titel: "Geweldig!", Inhoud: "Dit product is fantastisch."},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Ongeldige methode (GET)",
			method:         http.MethodGet,
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Methode niet toegestaan\n",
		},
		{
			name:           "Ongeldige JSON data",
			method:         http.MethodPost,
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Ongeldige review data:",
		},
		{
			name:           "Ontbrekende EAN in review",
			method:         http.MethodPost,
			requestBody:    models.Review{Naam: "Test", Score: 3},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Ontbrekende of ongeldige verplichte reviewvelden (EAN, Naam, Score)",
		},
		{
			name:           "Ongeldige score (buiten bereik)",
			method:         http.MethodPost,
			requestBody:    models.Review{ProductEAN: 1000000000000, Naam: "Test", Score: 0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Ontbrekende of ongeldige verplichte reviewvelden (EAN, Naam, Score)",
		},
		{
			name:           "Product niet gevonden voor review",
			method:         http.MethodPost,
			requestBody:    models.Review{ProductEAN: 9999999999999, Naam: "Test Reviewer", Score: 5, Titel: "Geweldig!", Inhoud: "Dit product is fantastisch."},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Product met opgegeven EAN niet gevonden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.requestBody != nil {
				if s, ok := tt.requestBody.(string); ok {
					reqBody = bytes.NewBufferString(s)
				} else {
					jsonBody, err := json.Marshal(tt.requestBody)
					if err != nil {
						t.Fatalf("Failed to marshal request body: %v", err)
					}
					reqBody = bytes.NewBuffer(jsonBody)
				}
			}

			req, err := http.NewRequest(tt.method, "/review", reqBody)
			if err != nil {
				t.Fatal(err)
			}
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}
			rr := httptest.NewRecorder()

			handlers.PlaatsReview(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if tt.expectedError != "" {
				if !strings.Contains(rr.Body.String(), tt.expectedError) {
					t.Errorf("handler returned unexpected error message: got %q want %q to contain %q", rr.Body.String(), rr.Body.String(), tt.expectedError)
				}
			}

			if tt.expectedStatus == http.StatusCreated {
				var createdReview models.Review
				err := json.Unmarshal(rr.Body.Bytes(), &createdReview)
				if err != nil {
					t.Errorf("Could not unmarshal response for created review: %v", err)
				}
				if createdReview.ID == 0 {
					t.Errorf("Expected created review to have an ID, got 0")
				}
				if createdReview.ProductEAN != tt.requestBody.(models.Review).ProductEAN {
					t.Errorf("Expected created review EAN %d, got %d", tt.requestBody.(models.Review).ProductEAN, createdReview.ProductEAN)
				}
			}
		})
	}
}

func TestVoegNieuwProductToe(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(db)

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Succesvol toevoegen van een nieuw product",
			method:         http.MethodPost,
			requestBody:    models.Product{EAN: 1000000000060, Naam: "Nieuw Product", SKU: "NP001", ProductTypeID: 1},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Ongeldige methode (GET)",
			method:         http.MethodGet,
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Methode niet toegestaan\n",
		},
		{
			name:           "Ongeldige JSON data",
			method:         http.MethodPost,
			requestBody:    "not json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Ongeldige product data:",
		},
		{
			name:           "Ontbrekende EAN",
			method:         http.MethodPost,
			requestBody:    models.Product{Naam: "Product zonder EAN"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "EAN en Naam zijn verplichte velden",
		},
		{
			name:           "Product met bestaande EAN",
			method:         http.MethodPost,
			requestBody:    models.Product{EAN: 1000000000000, Naam: "Bestaand Product", SKU: "DUP001", ProductTypeID: 1},
			expectedStatus: http.StatusConflict,
			expectedError:  "Product met dit EAN bestaat al",
		},
		{
			name:           "Product met niet-bestaande ProductTypeID",
			method:         http.MethodPost,
			requestBody:    models.Product{EAN: 1000000000061, Naam: "Product met onbekend type", SKU: "NP002", ProductTypeID: 999},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ProductType met opgegeven ID niet gevonden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db = setupTestDB(t)
			seedTestData(db)

			var reqBody io.Reader
			if tt.requestBody != nil {
				if s, ok := tt.requestBody.(string); ok {
					reqBody = bytes.NewBufferString(s)
				} else {
					jsonBody, err := json.Marshal(tt.requestBody)
					if err != nil {
						t.Fatalf("Failed to marshal request body: %v", err)
					}
					reqBody = bytes.NewBuffer(jsonBody)
				}
			}

			req, err := http.NewRequest(tt.method, "/product/add", reqBody)
			if err != nil {
				t.Fatal(err)
			}
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}
			rr := httptest.NewRecorder()

			handlers.VoegNieuwProductToe(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if tt.expectedError != "" {
				if !strings.Contains(rr.Body.String(), tt.expectedError) {
					t.Errorf("handler returned unexpected error message: got %q want %q to contain %q", rr.Body.String(), rr.Body.String(), tt.expectedError)
				}
			}

			if tt.expectedStatus == http.StatusCreated {
				var createdProduct models.Product
				err := json.Unmarshal(rr.Body.Bytes(), &createdProduct)
				if err != nil {
					t.Errorf("Could not unmarshal response for created product: %v", err)
				}
				if createdProduct.EAN != tt.requestBody.(models.Product).EAN {
					t.Errorf("Expected created product EAN %d, got %d", tt.requestBody.(models.Product).EAN, createdProduct.EAN)
				}
			}
		})
	}
}

func TestVoegProductAanbodToe(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(db)

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Succesvol toevoegen van een productaanbod",
			method:         http.MethodPost,
			requestBody:    models.ProductAanbod{ProductEAN: 1000000000000, Prijs: 450, Voorraad: 10, LeverancierID: 1},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Ongeldige methode (GET)",
			method:         http.MethodGet,
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Methode niet toegestaan\n",
		},
		{
			name:           "Ongeldige JSON data",
			method:         http.MethodPost,
			requestBody:    "bad json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Ongeldige aanbod data:",
		},
		{
			name:           "Ontbrekende ProductEAN",
			method:         http.MethodPost,
			requestBody:    models.ProductAanbod{Prijs: 100, Voorraad: 5, LeverancierID: 1},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ProductEAN, Prijs, Voorraad en LeverancierID zijn verplichte velden",
		},
		{
			name:           "Product niet gevonden",
			method:         http.MethodPost,
			requestBody:    models.ProductAanbod{ProductEAN: 9999999999999, Prijs: 100, Voorraad: 5, LeverancierID: 1},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Product met opgegeven EAN niet gevonden",
		},
		{
			name:           "Ontbrekende LeverancierID",
			method:         http.MethodPost,
			requestBody:    models.ProductAanbod{ProductEAN: 1000000000000, Prijs: 100, Voorraad: 5, LeverancierID: 0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ProductEAN, Prijs, Voorraad en LeverancierID zijn verplichte velden",
		},
		{
			name:           "Leverancier niet gevonden",
			method:         http.MethodPost,
			requestBody:    models.ProductAanbod{ProductEAN: 1000000000000, Prijs: 100, Voorraad: 5, LeverancierID: 999},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Leverancier met opgegeven ID niet gevonden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db = setupTestDB(t)
			seedTestData(db)

			var reqBody io.Reader
			if tt.requestBody != nil {
				if s, ok := tt.requestBody.(string); ok {
					reqBody = bytes.NewBufferString(s)
				} else {
					jsonBody, err := json.Marshal(tt.requestBody)
					if err != nil {
						t.Fatalf("Failed to marshal request body: %v", err)
					}
					reqBody = bytes.NewBuffer(jsonBody)
				}
			}

			req, err := http.NewRequest(tt.method, "/product/offer", reqBody)
			if err != nil {
				t.Fatal(err)
			}
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}
			rr := httptest.NewRecorder()

			handlers.VoegProductAanbodToe(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if tt.expectedError != "" {
				if !strings.Contains(rr.Body.String(), tt.expectedError) {
					t.Errorf("handler returned unexpected error message: got %q want %q to contain %q", rr.Body.String(), rr.Body.String(), tt.expectedError)
				}
			}

			if tt.expectedStatus == http.StatusCreated {
				var createdOffer models.ProductAanbod
				err := json.Unmarshal(rr.Body.Bytes(), &createdOffer)
				if err != nil {
					t.Errorf("Could not unmarshal response for created offer: %v", err)
				}
				if createdOffer.ID == 0 {
					t.Errorf("Expected created offer to have an ID, got 0")
				}
				if createdOffer.ProductEAN != tt.requestBody.(models.ProductAanbod).ProductEAN {
					t.Errorf("Expected created offer ProductEAN %d, got %d", tt.requestBody.(models.ProductAanbod).ProductEAN, createdOffer.ProductEAN)
				}
			}
		})
	}
}

func TestHaalTagsOp(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(db)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		expectedNames  []string
	}{
		{
			name:           "Alle tags ophalen (geen filters)",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			expectedNames:  []string{"lichtgewicht", "opvouwbaar", "comfort", "elektrisch"},
		},
		{
			name:           "Filter op niet-bestaande categorieID",
			queryParams:    "categorieID=999",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedNames:  []string{},
		},
		{
			name:           "Ongeldige categorieID-parameter",
			queryParams:    "categorieID=abc",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedNames:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/tags?"+tt.queryParams, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handlers.HaalTagsOp(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, tt.expectedStatus, rr.Body.String())
			}

			if tt.expectedStatus != http.StatusOK {
				return
			}

			var tags []models.Tag
			err = json.Unmarshal(rr.Body.Bytes(), &tags)
			if err != nil {
				t.Errorf("Could not unmarshal response: %v", err)
			}

			if len(tags) != tt.expectedCount {
				t.Errorf("handler returned unexpected number of tags: got %v want %v", len(tags), tt.expectedCount)
			}

			foundNames := make(map[string]bool)
			for _, t := range tags {
				foundNames[t.Naam] = true
			}
			for _, name := range tt.expectedNames {
				if !foundNames[name] {
					t.Errorf("Expected tag %s not found in response", name)
				}
			}
		})
	}
}
