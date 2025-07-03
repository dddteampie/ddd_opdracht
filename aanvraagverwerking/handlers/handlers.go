package handlers

import (
	"aanvraagverwerking/models"
	"encoding/json"
	"fmt"
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

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Aanvraagverwerking-service is gezond"))
}

func GetAanvraagByID(w http.ResponseWriter, r *http.Request) {
	// Haal de aanvraag ID uit de URL-variabelen
	vars := mux.Vars(r)
	aanvraagID := vars["id"]
	if aanvraagID == "" {
		http.Error(w, "Aanvraag ID is vereist", http.StatusBadRequest)
		return
	}

	// check of aanvraag bestaat 
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
	// Haal de clientId uit de URL-variabelen
	vars := mux.Vars(r)
	clientID := vars["clientId"]
	if clientID == "" {
		http.Error(w, "Client ID is vereist", http.StatusBadRequest)
		return
	}

	// Check of aanvraag bestaat voor deze client
	var aanvragen []models.Aanvraag
	if err := DB.Preload("Client").Preload("Behoefte").Where("client_id = ?", clientID).Find(&aanvragen).Error; err != nil {
		http.Error(w, "Fout bij ophalen aanvragen: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Controleer of er aanvragen zijn gevonden
	if len(aanvragen) == 0 {
		http.Error(w, "Geen aanvragen gevonden voor deze client", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aanvragen)
}

func StartAanvraag(w http.ResponseWriter, r *http.Request) {
	// 1. Decodeer de input
	client, behoefte, err := DecodeAanvraagInput(r)
	if err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// 2. Controleer of de client bestaat in de ECD
	bestaat, err := AanvraagBestaat(DB, client.ID, behoefte.ID)
	if err != nil {
		http.Error(w, "Databasefout", http.StatusInternalServerError)
		return
	}
	if bestaat {
		http.Error(w, "Er bestaat al een aanvraag voor deze client en behoefte", http.StatusConflict)
		return
	}

	// 3. Controleer of de client bestaat in de ECD
	aanvraag := BouwAanvraag(client, behoefte)
	if err := SlaAanvraagOp(DB, aanvraag); err != nil {
		http.Error(w, "Fout bij opslaan in database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(aanvraag)
}

func StartCategorieAanvraag(w http.ResponseWriter, r *http.Request) {
	// 1. Decodeer de input
	input, err := DecodeCategorieAanvraagInput(r)
	if err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// 2. Haal de aanvraag op op basis van clientID en behoefteBeschrijving
	clientUUID, err := uuid.Parse(input.ClientID)
	if err != nil {
		http.Error(w, "Ongeldige clientID", http.StatusBadRequest)
		return
	}
	aanvraag, err := VindAanvraagMetBehoefte(DB, clientUUID, input.BehoefteBeschrijving)
	if err != nil {
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	// 3. Stuur de categorie-aanvraag door naar de recommendation-service
	if err := StuurCategorieAanvraagNaarRecommendation(input); err != nil {
		log.Printf("Fout bij aanroepen recommendation service: %v", err)
		http.Error(w, "Fout bij aanroepen recommendation service", http.StatusBadGateway)
		return
	}

	// 4. Zet de status van de aanvraag op 'WachtenOpCategorieKeuze'
	if err := ZetStatusWachtenOpCategorie(DB, &aanvraag); err != nil {
		log.Printf("Fout bij updaten aanvraagstatus: %v", err)
		http.Error(w, "Fout bij updaten aanvraagstatus", http.StatusInternalServerError)
		return
	}

	// 5. Geef een succesvolle response terug
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Categorie-aanvraag gestart"))
}

func HaalPassendeCategorieenLijstOp(w http.ResponseWriter, r *http.Request) {
	// 1. Haal en valideer patientId uit de query
	patientID := r.URL.Query().Get("patientId")
	if patientID == "" {
		http.Error(w, "patientId is verplicht", http.StatusBadRequest)
		return
	}

	// 2. Haal de categorieënlijst op bij de recommendation-service
	lijst, status, err := VraagCategorieenLijstOp(patientID)
	if err != nil {
		if status == http.StatusNotFound {
			http.Error(w, fmt.Sprintf("geen passende categorieënlijst gevonden voor patient %s", patientID), http.StatusNotFound)
		} else if status != 0 {
			http.Error(w, fmt.Sprintf("onverwachte status van recommendation service: %d", status), http.StatusBadGateway)
		} else {
			http.Error(w, fmt.Sprintf("kon geen request doen: %v", err), http.StatusBadGateway)
		}
		return
	}

	// 3. Koppel de categorie-opties aan de aanvraag in de database
	if err := KoppelCategorieOptiesAanAanvraag(DB, patientID, lijst); err != nil {
		log.Printf("Fout bij koppelen categorie-opties: %v", err)
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	// 4. Geef de lijst terug als JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lijst)
}

func KiesCategorie(w http.ResponseWriter, r *http.Request) {
	// 1. Decode input
	var input struct {
		ClientID    uuid.UUID `json:"client_id"`
		BehoefteID  uuid.UUID `json:"behoefte_id"`
		CategorieID int       `json:"categorie"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// 2. Haal aanvraag op
	aanvraag, err := haalAanvraagOp(DB, input.ClientID, input.BehoefteID)
	if err != nil {
		log.Printf("Aanvraag niet gevonden voor client %s en behoefte %s: %v", input.ClientID, input.BehoefteID, err)
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	// 3. Controleer of gekozen categorie toegestaan is
	if !categorieToegestaan(aanvraag.CategorieOpties, input.CategorieID) {
		http.Error(w, "Gekozen categorie is niet toegestaan", http.StatusBadRequest)
		return
	}

	// 4. Sla de gekozen categorie op
	if err := slaGekozenCategorieOp(DB, &aanvraag, input.CategorieID); err != nil {
		log.Printf("Fout bij opslaan gekozen categorie: %v", err)
		http.Error(w, "Fout bij opslaan gekozen categorie", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Categorie succesvol gekozen"))
}

func StartProductAanvraag(w http.ResponseWriter, r *http.Request) {
	var input models.ProductAanvraagDTO
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	if err := validateProductAanvraagInput(input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	clientUUID, err := uuid.Parse(input.ClientID)
	if err != nil {
		http.Error(w, "Ongeldige clientID", http.StatusBadRequest)
		return
	}

	aanvraag, err := haalProductAanvraagOp(DB, clientUUID, input.BehoefteBeschrijving)
	if err != nil {
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	if aanvraag.GekozenCategorieID == nil || input.GekozenCategorieID == nil || *aanvraag.GekozenCategorieID != *input.GekozenCategorieID {
		http.Error(w, "Geef juiste categorie mee", http.StatusBadRequest)
		return
	}

	if err := stuurProductAanvraagNaarRecommendation(input); err != nil {
		log.Printf("Fout bij aanroepen recommendation service: %v", err)
		http.Error(w, "Fout bij aanroepen recommendation service", http.StatusBadGateway)
		return
	}

	aanvraag.Status = models.WachtenOpProductKeuze
	if err := DB.Save(&aanvraag).Error; err != nil {
		log.Printf("Fout bij updaten aanvraagstatus: %v", err)
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Product-aanvraag gestart"))
}

func HaalPassendeProductenLijstOp(w http.ResponseWriter, r *http.Request) {
	// 1. Haal en valideer clientId uit de query
	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		http.Error(w, "clientId is verplicht", http.StatusBadRequest)
		return
	}

	// 2. Haal de productenlijst op bij de recommendation-service
	lijst, status, err := vraagProductenLijstOp(clientID)
	if err != nil {
		if status == http.StatusNotFound {
			http.Error(w, fmt.Sprintf("geen passende productlijst gevonden voor client %s", clientID), http.StatusNotFound)
		} else if status != 0 {
			http.Error(w, fmt.Sprintf("onverwachte status van recommendation service: %d", status), http.StatusBadGateway)
		} else {
			http.Error(w, fmt.Sprintf("kon geen request doen: %v", err), http.StatusBadGateway)
		}
		return
	}

	// 3. Koppel de product-opties aan de aanvraag in de database
	if err := koppelProductOptiesAanAanvraag(DB, clientID, lijst); err != nil {
		log.Printf("Fout bij koppelen product-opties: %v", err)
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	// 4. Geef de lijst terug als JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lijst)
}

func KiesProduct(w http.ResponseWriter, r *http.Request) {
	// 1. Decode input
	var input struct {
		ClientID   uuid.UUID `json:"client_id"`
		BehoefteID uuid.UUID `json:"behoefte_id"`
		ProductEAN int64     `json:"product_ean"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// 2. Haal aanvraag op
	aanvraag, err := haalAanvraagOp(DB, input.ClientID, input.BehoefteID)
	if err != nil {
		log.Printf("Aanvraag niet gevonden voor client %s en behoefte %s: %v", input.ClientID, input.BehoefteID, err)
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	// 3. Controleer of gekozen product toegestaan is
	if !productToegestaan(aanvraag.ProductOpties, input.ProductEAN) {
		http.Error(w, "Gekozen product is niet toegestaan", http.StatusBadRequest)
		return
	}

	// 4. Sla het gekozen product op
	if err := slaGekozenProductOp(DB, &aanvraag, input.ProductEAN); err != nil {
		log.Printf("Fout bij opslaan gekozen product: %v", err)
		http.Error(w, "Fout bij opslaan gekozen product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Product succesvol gekozen"))
}
