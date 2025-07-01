package handlers

import (
	"aanvraagverwerking/helper"
	"aanvraagverwerking/models"
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitHandlers(db *gorm.DB) {
	DB = db
}

func GetAanvraagByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aanvraagID := vars["id"]
	if aanvraagID == "" {
		http.Error(w, "Aanvraag ID is vereist", http.StatusBadRequest)
		return
	}

	var aanvraag models.Aanvraag
	if err := DB.Preload("Client").Preload("Behoefte").First(&aanvraag, "id = ?", aanvraagID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		} else {
			http.Error(w, "Fout bij ophalen aanvraag: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aanvraag)
}

func GetAanvragenByClientID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    clientID := vars["clientId"]
    if clientID == "" {
        http.Error(w, "Client ID is vereist", http.StatusBadRequest)
        return
    }

    var aanvragen []models.Aanvraag
    if err := DB.Preload("Client").Preload("Behoefte").Where("client_id = ?", clientID).Find(&aanvragen).Error; err != nil {
        http.Error(w, "Fout bij ophalen aanvragen: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if len(aanvragen) == 0 {
        http.Error(w, "Geen aanvragen gevonden voor deze client", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(aanvragen)
}

func StartAanvraag(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Client   models.Client   `json:"client"`
		Behoefte models.Behoefte `json:"behoefte"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// Controleer of er al een aanvraag bestaat met dezelfde ClientID en BehoefteID
	var bestaandeAanvraag models.Aanvraag
	if err := DB.Where("client_id = ? AND behoefte_id = ?", input.Client.ID, input.Behoefte.ID).First(&bestaandeAanvraag).Error; err == nil {
		http.Error(w, "Er bestaat al een aanvraag voor deze client en behoefte", http.StatusConflict)
		return
	}

	//Maak een random budget aan tussen 200 en 5000
	budget := helper.RandomFloat64Between(200, 5000)

	aanvraag := models.Aanvraag{
		ID:         uuid.New(),
		ClientID:   input.Client.ID,
		BehoefteID: input.Behoefte.ID,
		Client:     input.Client,
		Behoefte:   input.Behoefte,
		Status:     models.BehoefteOntvangen,
		Budget:     budget,
	}

	if err := DB.Create(&aanvraag).Error; err != nil {
		http.Error(w, "Fout bij opslaan in database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(aanvraag)
}

func StartCategorieAanvraag(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ClientID   uuid.UUID `json:"client_id"`
		BehoefteID uuid.UUID `json:"behoefte_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// Haal de aanvraag op, inclusief de behoefte
	var aanvraag models.Aanvraag
	if err := DB.Preload("Behoefte").Where("client_id = ? AND behoefte_id = ?", input.ClientID, input.BehoefteID).First(&aanvraag).Error; err != nil {
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	// --- Bouw de DTO voor de recommendation service ---
	categorieAanvraag := models.CategorieAanvraagDTO{
		ClientID:             input.ClientID,
		BehoefteBeschrijving: aanvraag.Behoefte.Beschrijving,
		Budget:               aanvraag.Budget,
	}
	jsonPayload, err := json.Marshal(categorieAanvraag)
	if err != nil {
		http.Error(w, "Fout bij serialiseren JSON", http.StatusInternalServerError)
		return
	}
	url := "http://recommendation-service:8080/recommend/categorie"
	recResp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Fout bij aanroepen recommendation service: %v", err)
		http.Error(w, "Fout bij aanroepen recommendation service", http.StatusBadGateway)
		return
	}
	defer recResp.Body.Close()

	// Zet de status op WachtenOpCategorieKeuze
	aanvraag.Status = models.WachtenOpCategorieKeuze
	if err := DB.Save(&aanvraag).Error; err != nil {
		log.Printf("Fout bij updaten aanvraagstatus: %v", err)
	}

	// OF NOG BETER: ONTVANG GELIJK een response met lijst van categorieën

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Categorie-aanvraag gestart"))
}

func KiesCategorie(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ClientID    uuid.UUID `json:"client_id"`
		BehoefteID  uuid.UUID `json:"behoefte_id"`
		CategorieID int       `json:"categorie"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// Haal de aanvraag op
	var aanvraag models.Aanvraag
	if err := DB.Where("client_id = ? AND behoefte_id = ?", input.ClientID, input.BehoefteID).First(&aanvraag).Error; err != nil {
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	// Sla de gekozen categorie op
	aanvraag.GekozenCategorieID = &input.CategorieID
	aanvraag.Status = models.CategorieGekozen // Voeg deze status toe aan je enum/consts

	if err := DB.Save(&aanvraag).Error; err != nil {
		log.Printf("Fout bij opslaan gekozen categorie: %v", err)
		http.Error(w, "Fout bij opslaan gekozen categorie", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Categorie succesvol gekozen"))
}

func StartProductAanvraag(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ClientID    uuid.UUID `json:"client_id"`
		BehoefteID  uuid.UUID `json:"behoefte_id"`
		CategorieID int       `json:"categorie_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// Haal de aanvraag op
	var aanvraag models.Aanvraag
	if err := DB.Where("client_id = ? AND behoefte_id = ?", input.ClientID, input.BehoefteID).First(&aanvraag).Error; err != nil {
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	if aanvraag.GekozenCategorieID == nil || *aanvraag.GekozenCategorieID != input.CategorieID {
		http.Error(w, "Categorie moet eerst gekozen zijn", http.StatusBadRequest)
		return
	}

	// --- Bouw de DTO voor de recommendation service ---
	productAanvraag := models.ProductAanvraagDTO{
		ClientID:             input.ClientID,
		BehoefteBeschrijving: aanvraag.Behoefte.Beschrijving,
		Budget:               aanvraag.Budget,
		GekozenCategorieID:   input.CategorieID,
	}
	jsonPayload, err := json.Marshal(productAanvraag)
	if err != nil {
		http.Error(w, "Fout bij serialiseren JSON", http.StatusInternalServerError)
		return
	}
	url := "http://recommendation-service:8080/recommend/product"
	recResp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Fout bij aanroepen recommendation service: %v", err)
		http.Error(w, "Fout bij aanroepen recommendation service", http.StatusBadGateway)
		return
	}
	defer recResp.Body.Close()

	if recResp.StatusCode != http.StatusOK {
		http.Error(w, "Fout bij ophalen product aanbevelingen", recResp.StatusCode)
		return
	}

	// Zet de status op WachtenOpProductKeuze
	aanvraag.Status = models.WachtenOpProductKeuze
	if err := DB.Save(&aanvraag).Error; err != nil {
		log.Printf("Fout bij updaten aanvraagstatus: %v", err)
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Product-aanvraag gestart"))
}
