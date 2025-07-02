package handlers

import (
	"bytes"
	"encoding/json"

	//"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"aanvraagverwerking/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	//"gorm.io/driver/sqlite"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Helper voor een in-memory testdatabase
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&models.Aanvraag{}, &models.Client{}, &models.Behoefte{}); err != nil {
		panic(err)
	}
	// ...debug code...
	return db
}

// --- Tests voor GetAanvraagByID ---
// TestGetAanvraagByID_Succes test of een aanvraag succesvol kan worden opgehaald
func TestGetAanvraagByID_Succes(t *testing.T) {
    db := setupTestDB()
	DB = db
    client := models.Client{ID: uuid.New()}
    behoefte := models.Behoefte{ID: uuid.New()}
    aanvraag := models.Aanvraag{
        ID:         uuid.New(),
        ClientID:   client.ID,
        BehoefteID: behoefte.ID,
        Client:     client,
        Behoefte:   behoefte,
    }
    db.Create(&client)
    db.Create(&behoefte)
    db.Create(&aanvraag)

    req := httptest.NewRequest("GET", "/aanvraag/"+aanvraag.ID.String(), nil)
    req = mux.SetURLVars(req, map[string]string{"id": aanvraag.ID.String()})
    w := httptest.NewRecorder()

    GetAanvraagByID(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("verwacht 200, kreeg %d", w.Code)
    }
    var got models.Aanvraag
    if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
        t.Fatalf("kon response niet decoden: %v", err)
    }
    if got.ID != aanvraag.ID {
        t.Errorf("Verkeerde aanvraag teruggegeven")
    }
}

// TestGetAanvraagByID_NotFound test of een aanvraag niet gevonden wordt wanneer deze niet bestaat
func TestGetAanvraagByID_NotFound(t *testing.T) {
    db := setupTestDB()
	DB = db
    req := httptest.NewRequest("GET", "/aanvraag/"+uuid.New().String(), nil)
    req = mux.SetURLVars(req, map[string]string{"id": uuid.New().String()})
    w := httptest.NewRecorder()

    GetAanvraagByID(w, req)
    if w.Code != http.StatusNotFound {
        t.Errorf("verwacht 404, kreeg %d", w.Code)
    }
}

// TestGetAanvraagByID_BadRequest test of een aanvraag niet gevonden wordt wanneer de ID ongeldig is
func TestGetAanvraagByID_BadRequest(t *testing.T) {
    db := setupTestDB()
	DB = db
    req := httptest.NewRequest("GET", "/aanvraag/", nil)
    req = mux.SetURLVars(req, map[string]string{"id": ""})
    w := httptest.NewRecorder()

    GetAanvraagByID(w, req)
    if w.Code != http.StatusBadRequest {
        t.Errorf("verwacht 400, kreeg %d", w.Code)
    }
}

// --- Tests voor GetAanvragenByClientID ---
// TestGetAanvragenByClientID_Succes test of aanvragen succesvol kunnen worden opgehaald voor een specifieke client
func TestGetAanvragenByClientID_Succes(t *testing.T) {
    db := setupTestDB()
	DB = db
    client := models.Client{ID: uuid.New()}
    behoefte := models.Behoefte{ID: uuid.New()}
    aanvraag := models.Aanvraag{
        ID:         uuid.New(),
        ClientID:   client.ID,
        BehoefteID: behoefte.ID,
        Client:     client,
        Behoefte:   behoefte,
    }
    db.Create(&client)
    db.Create(&behoefte)
    db.Create(&aanvraag)

    req := httptest.NewRequest("GET", "/aanvragen/"+client.ID.String(), nil)
    req = mux.SetURLVars(req, map[string]string{"clientId": client.ID.String()})
    w := httptest.NewRecorder()

    GetAanvragenByClientID(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("verwacht 200, kreeg %d", w.Code)
    }
    var got []models.Aanvraag
    if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
        t.Fatalf("kon response niet decoden: %v", err)
    }
    if len(got) != 1 || got[0].ID != aanvraag.ID {
        t.Errorf("Verkeerde aanvragen teruggegeven")
    }
}

// TestGetAanvragenByClientID_NotFound test of aanvragen niet gevonden worden wanneer er geen aanvragen zijn voor de client
func TestGetAanvragenByClientID_NotFound(t *testing.T) {
    db := setupTestDB()
	DB = db
    req := httptest.NewRequest("GET", "/aanvragen/"+uuid.New().String(), nil)
    req = mux.SetURLVars(req, map[string]string{"clientId": uuid.New().String()})
    w := httptest.NewRecorder()

    GetAanvragenByClientID(w, req)
    if w.Code != http.StatusNotFound {
        t.Errorf("verwacht 404, kreeg %d", w.Code)
    }
}

// TestGetAanvragenByClientID_BadRequest test of aanvragen niet gevonden worden wanneer de client ID ongeldig is
func TestGetAanvragenByClientID_BadRequest(t *testing.T) {
    db := setupTestDB()
	DB = db
    req := httptest.NewRequest("GET", "/aanvragen/", nil)
    req = mux.SetURLVars(req, map[string]string{"clientId": ""})
    w := httptest.NewRecorder()

    GetAanvragenByClientID(w, req)
    if w.Code != http.StatusBadRequest {
        t.Errorf("verwacht 400, kreeg %d", w.Code)
    }
}

// --- Tests voor aanvraag ---
// --- Tests voor DecodeAanvraagInput ---
func TestDecodeAanvraagInput_Success(t *testing.T) {
	client := models.Client{ID: uuid.New(), Naam: "Test"}
	behoefte := models.Behoefte{ID: uuid.New(), Beschrijving: "Test"}
	body := map[string]interface{}{
		"client":   client,
		"behoefte": behoefte,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/aanvraag", bytes.NewBuffer(jsonBody))

	gotClient, gotBehoefte, err := DecodeAanvraagInput(req)
	if err != nil {
		t.Fatalf("verwacht geen error, kreeg: %v", err)
	}
	if gotClient.ID != client.ID {
		t.Errorf("verwacht client ID %v, kreeg %v", client.ID, gotClient.ID)
	}
	if gotBehoefte.ID != behoefte.ID {
		t.Errorf("verwacht behoefte ID %v, kreeg %v", behoefte.ID, gotBehoefte.ID)
	}
}

func TestDecodeAanvraagInput_BadRequest(t *testing.T) {
	req, _ := http.NewRequest("POST", "/aanvraag", bytes.NewBuffer([]byte("invalid json")))
	_, _, err := DecodeAanvraagInput(req)
	if err == nil {
		t.Error("verwacht error bij ongeldige input")
	}
}

// --- Tests voor AanvraagBestaat ---
func TestAanvraagBestaat_Bestaat(t *testing.T) {
	db := setupTestDB()
	clientID := uuid.New()
	behoefteID := uuid.New()
	aanvraag := models.Aanvraag{ID: uuid.New(), ClientID: clientID, BehoefteID: behoefteID}
	if err := db.Create(&aanvraag).Error; err != nil {
		t.Fatalf("Fout bij aanmaken aanvraag: %v", err)
	}

	bestaat, err := AanvraagBestaat(db, clientID, behoefteID)
	if err != nil {
		t.Fatalf("geen error verwacht, kreeg: %v", err)
	}
	if !bestaat {
		t.Error("aanvraag zou moeten bestaan")
	}
}

func TestAanvraagBestaat_BestaatNiet(t *testing.T) {
	db := setupTestDB()
	clientID := uuid.New()
	behoefteID := uuid.New()

	bestaat, err := AanvraagBestaat(db, clientID, behoefteID)
	if err != nil {
		t.Fatalf("geen error verwacht, kreeg: %v", err)
	}
	if bestaat {
		t.Error("aanvraag zou niet moeten bestaan")
	}
}

// --- Tests voor BouwAanvraag ---
func TestBouwAanvraag_Slaagt(t *testing.T) {
	client := models.Client{ID: uuid.New(), Naam: "Test"}
	behoefte := models.Behoefte{ID: uuid.New(), Beschrijving: "Test"}
	aanvraag := BouwAanvraag(client, behoefte)

	if aanvraag.ClientID != client.ID {
		t.Errorf("ClientID niet correct")
	}
	if aanvraag.BehoefteID != behoefte.ID {
		t.Errorf("BehoefteID niet correct")
	}
	if aanvraag.Status != models.BehoefteOntvangen {
		t.Errorf("Status niet correct")
	}
	if aanvraag.Budget < 200 || aanvraag.Budget > 5000 {
		t.Errorf("Budget buiten bereik: %f", aanvraag.Budget)
	}
}

func TestBouwAanvraag_Faalt(t *testing.T) {
	// Lege client en behoefte
	client := models.Client{}
	behoefte := models.Behoefte{}
	aanvraag := BouwAanvraag(client, behoefte)

	if aanvraag.ClientID != client.ID {
		t.Log("ClientID niet correct (verwacht bij lege input)")
	}
	if aanvraag.BehoefteID != behoefte.ID {
		t.Log("BehoefteID niet correct (verwacht bij lege input)")
	}
	if aanvraag.Status != models.BehoefteOntvangen {
		t.Error("Status niet correct bij lege input")
	}
	if aanvraag.Budget < 200 || aanvraag.Budget > 5000 {
		t.Error("Budget buiten bereik bij lege input")
	}
}

// --- Tests voor SlaAanvraagOp ---
func TestSlaAanvraagOp_Slaagt(t *testing.T) {
	db := setupTestDB()
	client := models.Client{ID: uuid.New(), Naam: "Test"}
	behoefte := models.Behoefte{ID: uuid.New(), Beschrijving: "Test"}
	aanvraag := BouwAanvraag(client, behoefte)

	err := SlaAanvraagOp(db, aanvraag)
	if err != nil {
		t.Fatalf("verwacht geen error bij opslaan, kreeg: %v", err)
	}

	// Controleer of hij echt in de database staat
	var gevonden models.Aanvraag
	if err := db.First(&gevonden, "id = ?", aanvraag.ID).Error; err != nil {
		t.Errorf("aanvraag niet gevonden in database")
	}
}

// --- Tests voor CategorieAanvraag utils ---
// --- Tests voor DecodeCategorieAanvraagInput ---
func TestDecodeCategorieAanvraagInput_Success(t *testing.T) {
    dto := models.CategorieAanvraagDTO{
        ClientID:            uuid.New().String(),
        BehoefteBeschrijving: "TestBehoefte",
    }
    body, _ := json.Marshal(dto)
    req := httptest.NewRequest("POST", "/dummy", bytes.NewBuffer(body))

    got, err := DecodeCategorieAanvraagInput(req)
    if err != nil {
        t.Fatalf("verwacht geen error, kreeg: %v", err)
    }
    if got.ClientID != dto.ClientID || got.BehoefteBeschrijving != dto.BehoefteBeschrijving {
        t.Errorf("Decoded input niet correct: %+v", got)
    }
}

func TestDecodeCategorieAanvraagInput_BadRequest(t *testing.T) {
    req := httptest.NewRequest("POST", "/dummy", bytes.NewBuffer([]byte("invalid json")))
    _, err := DecodeCategorieAanvraagInput(req)
    if err == nil {
        t.Error("verwacht error bij ongeldige input")
    }
}

// --- Tests voor VindAanvraagMetBehoefte ---
func TestVindAanvraagMetBehoefte_Bestaat(t *testing.T) {
    db := setupTestDB()
    clientID := uuid.New()
    behoefte := models.Behoefte{ID: uuid.New(), Beschrijving: "TestBehoefte"}
    db.Create(&behoefte)
    aanvraag := models.Aanvraag{
        ID:         uuid.New(),
        ClientID:   clientID,
        BehoefteID: behoefte.ID,
        Behoefte:   behoefte,
    }
    db.Create(&aanvraag)

    gevonden, err := VindAanvraagMetBehoefte(db, clientID, "TestBehoefte")
    if err != nil {
        t.Fatalf("verwacht geen error, kreeg: %v", err)
    }
    if gevonden.ID != aanvraag.ID {
        t.Errorf("Verkeerde aanvraag gevonden")
    }
}

func TestVindAanvraagMetBehoefte_BestaatNiet(t *testing.T) {
    db := setupTestDB()
    clientID := uuid.New()
    _, err := VindAanvraagMetBehoefte(db, clientID, "NietBestaand")
    if err == nil {
        t.Error("verwacht error bij niet-bestaande aanvraag")
    }
}

// --- Tests voor ZetStatusWachtenOpCategorie ---
func TestZetStatusWachtenOpCategorie_Slaagt(t *testing.T) {
    db := setupTestDB()
    clientID := uuid.New()
    behoefte := models.Behoefte{ID: uuid.New(), Beschrijving: "TestBehoefte"}
    db.Create(&behoefte)
    aanvraag := models.Aanvraag{
        ID:         uuid.New(),
        ClientID:   clientID,
        BehoefteID: behoefte.ID,
        Behoefte:   behoefte,
        Status:     "OudStatus",
    }
    db.Create(&aanvraag)

    err := ZetStatusWachtenOpCategorie(db, &aanvraag)
    if err != nil {
        t.Fatalf("verwacht geen error, kreeg: %v", err)
    }
    var gevonden models.Aanvraag
    db.First(&gevonden, "id = ?", aanvraag.ID)
    if gevonden.Status != models.WachtenOpCategorieKeuze {
        t.Errorf("Status niet correct geüpdatet, kreeg: %v", gevonden.Status)
    }
}

func TestZetStatusWachtenOpCategorie_Faalt(t *testing.T) {
    db := setupTestDB()
    // Maak een aanvraag die niet in de database staat
    aanvraag := models.Aanvraag{
        ID:         uuid.New(),
        ClientID:   uuid.New(),
        BehoefteID: uuid.New(),
        Status:     "OudStatus",
    }

    err := ZetStatusWachtenOpCategorie(db, &aanvraag)
    if err == nil {
        t.Error("verwacht een error bij updaten van niet-bestaande aanvraag, maar kreeg geen error")
    }
}

// Test voor KoppelCategorieOptiesAanAanvraag (succes als aanvraag bestaat)
func TestKoppelCategorieOptiesAanAanvraag_Success(t *testing.T) {
    db := setupTestDB()
    clientID := uuid.New()
    aanvraag := models.Aanvraag{
        ID:       uuid.New(),
        ClientID: clientID,
    }
    db.Create(&aanvraag)

    lijst := models.CategorieShortListDTO{
        Categorielijst: []models.CategorieDTO{
            {ID: 1, Naam: "TestCat"},
            {ID: 2, Naam: "TestCat2"},
        },
    }

    err := KoppelCategorieOptiesAanAanvraag(db, clientID.String(), lijst)
    if err != nil {
        t.Fatalf("verwacht geen error, kreeg: %v", err)
    }

    var gevonden models.Aanvraag
    db.First(&gevonden, "id = ?", aanvraag.ID)
    if len(gevonden.CategorieOpties) != 2 || gevonden.CategorieOpties[0] != 1 {
        t.Errorf("CategorieOpties niet correct opgeslagen: %+v", gevonden.CategorieOpties)
    }
}

// Test voor KoppelCategorieOptiesAanAanvraag (faalt als aanvraag niet bestaat)
func TestKoppelCategorieOptiesAanAanvraag_Faalt(t *testing.T) {
    db := setupTestDB()
    lijst := models.CategorieShortListDTO{
        Categorielijst: []models.CategorieDTO{{ID: 1, Naam: "TestjesDaarHoudIkVan"}},
    }
    err := KoppelCategorieOptiesAanAanvraag(db, uuid.New().String(), lijst)
    if err == nil {
        t.Error("verwacht error bij niet-bestaande aanvraag")
    }
}

// --- Tests voor KiesCategorie
// Tests voor haalAanvraagOp: Als de aanvraag bestaat, moet deze worden opgehaald
func TestHaalAanvraagOp_Succes(t *testing.T) {
    db := setupTestDB()
    clientID := uuid.New()
    behoefteID := uuid.New()
    aanvraag := models.Aanvraag{
        ID:         uuid.New(),
        ClientID:   clientID,
        BehoefteID: behoefteID,
    }
    db.Create(&aanvraag)

    gevonden, err := haalAanvraagOp(db, clientID, behoefteID)
    if err != nil {
        t.Fatalf("verwacht geen error, kreeg: %v", err)
    }
    if gevonden.ID != aanvraag.ID {
        t.Errorf("Verkeerde aanvraag gevonden")
    }
}

// Tests voor haalAanvraagOp: Als de aanvraag niet bestaat, moet er een error zijn
func TestHaalAanvraagOp_Faalt(t *testing.T) {
    db := setupTestDB()
    _, err := haalAanvraagOp(db, uuid.New(), uuid.New())
    if err == nil {
        t.Error("verwacht error bij niet-bestaande aanvraag")
    }
}

// Tests voor categorieToegestaan: Controleer of een categorie in de lijst zit en dus toegestaan is
func TestCategorieToegestaan_True(t *testing.T) {
    opties := []int64{1, 2, 3}
    if !categorieToegestaan(opties, 2) {
        t.Error("verwacht true voor bestaande categorie")
    }
}

// Tests voor categorieToegestaan: Controleer of een categorie niet in de lijst zit en dus niet toegestaan is
func TestCategorieToegestaan_False(t *testing.T) {
    opties := []int64{1, 2, 3}
    if categorieToegestaan(opties, 5) {
        t.Error("verwacht false voor niet-bestaande categorie")
    }
}

// Tests voor slaGekozenCategorieOp: Sla de gekozen categorie op in de aanvraag
func TestSlaGekozenCategorieOp_Succes(t *testing.T) {
    db := setupTestDB()
    aanvraag := models.Aanvraag{
        ID:     uuid.New(),
        Status: "OudStatus",
    }
    db.Create(&aanvraag)

    err := slaGekozenCategorieOp(db, &aanvraag, 42)
    if err != nil {
        t.Fatalf("verwacht geen error, kreeg: %v", err)
    }
    var gevonden models.Aanvraag
    db.First(&gevonden, "id = ?", aanvraag.ID)
    if gevonden.GekozenCategorieID == nil || *gevonden.GekozenCategorieID != 42 {
        t.Errorf("GekozenCategorieID niet correct opgeslagen")
    }
    if gevonden.Status != models.CategorieGekozen {
        t.Errorf("Status niet correct geüpdatet")
    }
}

// --- Tests voor Startproductaanvraag utils ---
// Tests voor validateProductAanvraagInput: Controleer of de input correct gevalideerd wordt
func TestValidateProductAanvraagInput(t *testing.T) {
    validUUID := uuid.New().String()
    catID := 1

    tests := []struct {
        name    string
        input   models.ProductAanvraagDTO
        wantErr bool
    }{
        {
            name: "valid input",
            input: models.ProductAanvraagDTO{
                ClientID:             validUUID,
                BehoefteBeschrijving: "Test",
                GekozenCategorieID:   &catID,
            },
            wantErr: false,
        },
        {
            name: "empty clientID",
            input: models.ProductAanvraagDTO{
                ClientID:             "",
                BehoefteBeschrijving: "Test",
                GekozenCategorieID:   &catID,
            },
            wantErr: true,
        },
        {
            name: "invalid clientID",
            input: models.ProductAanvraagDTO{
                ClientID:             "123",
                BehoefteBeschrijving: "Test",
                GekozenCategorieID:   &catID,
            },
            wantErr: true,
        },
        {
            name: "empty behoefte",
            input: models.ProductAanvraagDTO{
                ClientID:             validUUID,
                BehoefteBeschrijving: "",
                GekozenCategorieID:   &catID,
            },
            wantErr: true,
        },
        {
            name: "nil categorie",
            input: models.ProductAanvraagDTO{
                ClientID:             validUUID,
                BehoefteBeschrijving: "Test",
                GekozenCategorieID:   nil,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateProductAanvraagInput(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateProductAanvraagInput() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

// Tests voor haalProductAanvraagOp: Haal een productaanvraag op voor een client en behoefte
func TestHaalProductAanvraagOp_Succes(t *testing.T) {
    db := setupTestDB()
    clientID := uuid.New()
    behoefte := models.Behoefte{ID: uuid.New(), Beschrijving: "TestBehoefte"}
    db.Create(&behoefte)
    aanvraag := models.Aanvraag{
        ID:         uuid.New(),
        ClientID:   clientID,
        BehoefteID: behoefte.ID,
        Behoefte:   behoefte,
    }
    db.Create(&aanvraag)

    gevonden, err := haalProductAanvraagOp(db, clientID, "TestBehoefte")
    if err != nil {
        t.Fatalf("verwacht geen error, kreeg: %v", err)
    }
    if gevonden.ID != aanvraag.ID {
        t.Errorf("Verkeerde aanvraag gevonden")
    }
}

// Tests voor haalProductAanvraagOp: Controleer of een error wordt gegeven bij een niet-bestaande aanvraag
func TestHaalProductAanvraagOp_Faalt(t *testing.T) {
    db := setupTestDB()
    clientID := uuid.New()
    _, err := haalProductAanvraagOp(db, clientID, "NietBestaand")
    if err == nil {
        t.Error("verwacht error bij niet-bestaande aanvraag")
    }
}

// --- Tests voor Vraagproductenlijstop utils ---
// Tests voor vraagProductenLijstOp: Controleer of de productenlijst correct wordt opgehaald
func TestKoppelProductOptiesAanAanvraag_Succes(t *testing.T) {
    db := setupTestDB()
    clientID := uuid.New().String()
    aanvraag := models.Aanvraag{
        ID:           uuid.New(),
        ClientID:     uuid.MustParse(clientID),
        ProductOpties: []int64{},
    }
    db.Create(&aanvraag)

    lijst := models.ProductShortListDTO{
        Productlijst: []models.ProductDTO{
            {EAN: 111},
            {EAN: 222},
        },
    }
    err := koppelProductOptiesAanAanvraag(db, clientID, lijst)
    if err != nil {
        t.Fatalf("verwacht geen error, kreeg: %v", err)
    }

    var updated models.Aanvraag
    db.First(&updated, "client_id = ?", clientID)
    if len(updated.ProductOpties) != 2 {
        t.Errorf("verwacht 2 productopties, kreeg %d", len(updated.ProductOpties))
    }
}

// Tests voor koppelProductOptiesAanAanvraag: Controleer of een error wordt gegeven bij een niet-bestaande aanvraag
func TestKoppelProductOptiesAanAanvraag_NietGevonden(t *testing.T) {
    db := setupTestDB()
    lijst := models.ProductShortListDTO{
        Productlijst: []models.ProductDTO{{EAN: 111}},
    }
    err := koppelProductOptiesAanAanvraag(db, uuid.New().String(), lijst)
    if err == nil {
        t.Error("verwacht error bij niet-bestaande aanvraag")
    }
}

// --- Tests voor KiesProduct utils ---
// Tests voor productToegestaan: Controleer of een product in de lijst zit en dus toegestaan is
func TestProductToegestaan_True(t *testing.T) {
    opties := []int64{111, 222, 333}
    if !productToegestaan(opties, 222) {
        t.Error("222 zou toegestaan moeten zijn")
    }
}

// Tests voor productToegestaan: Controleer of een product niet in de lijst zit en dus niet toegestaan is
func TestProductToegestaan_False(t *testing.T) {
    opties := []int64{111, 222, 333}
    if productToegestaan(opties, 444) {
        t.Error("444 zou NIET toegestaan moeten zijn")
    }
}

// Tests voor slaGekozenProductOp: Sla het gekozen product op in de aanvraag als het bestaat
func TestSlaGekozenProductOp_Succes(t *testing.T) {
    db := setupTestDB()
    aanvraag := models.Aanvraag{
        ID:         uuid.New(),
        ClientID:   uuid.New(),
        BehoefteID: uuid.New(),
        Status:     "Oud",
    }
    db.Create(&aanvraag)

    gekozenEAN := int64(555)
    err := slaGekozenProductOp(db, &aanvraag, gekozenEAN)
    if err != nil {
        t.Fatalf("verwacht geen error, kreeg: %v", err)
    }

    var updated models.Aanvraag
    db.First(&updated, "id = ?", aanvraag.ID)
    if updated.GekozenProductID == nil || *updated.GekozenProductID != gekozenEAN {
        t.Errorf("GekozenProductID niet goed opgeslagen")
    }
    if updated.Status != models.ProductGekozen {
        t.Errorf("Status niet goed opgeslagen, kreeg: %v", updated.Status)
    }
}