package handlers

import (
	"behoeftebepaling/models"
	"behoeftebepaling/service"
	"encoding/json"
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
	w.Write([]byte("Behoeftebepaling-service is gezond"))
}

// var behoeften []models.Behoefte
func CreateBehoefte(w http.ResponseWriter, r *http.Request) {
	var behoefte models.Behoefte
	if err := json.NewDecoder(r.Body).Decode(&behoefte); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	if behoefte.OnderzoekID == uuid.Nil {
		http.Error(w, "OnderzoekID is verplicht", http.StatusBadRequest)
		return
	}
	var onderzoek models.Onderzoek
	if err := DB.First(&onderzoek, "id = ?", behoefte.OnderzoekID).Error; err != nil {
		http.Error(w, "Onderzoek bestaat niet", http.StatusBadRequest)
		return
	}
	if behoefte.ClientID == uuid.Nil {
		http.Error(w, "ClientID is verplicht", http.StatusBadRequest)
		return
	}
	var client models.Client
	if err := DB.First(&client, "id = ?", behoefte.ClientID).Error; err != nil {
		http.Error(w, "Client bestaat niet", http.StatusBadRequest)
		return
	}

	// 1. Check of client bestaat in ECD
	exists, err := service.ClientExistsInECD(ecdURL, behoefte.ClientID.String())
	if err != nil {
		http.Error(w, "Fout bij controleren client in ECD", http.StatusBadGateway)
		return
	}
	if !exists {
		http.Error(w, "Client bestaat niet in ECD", http.StatusBadRequest)
		return
	}

	// // 3. Check of onderzoek bestaat
	exists, err = service.OnderzoekExists(ecdURL, behoefte.OnderzoekID.String())
	if err != nil {
		http.Error(w, "Fout bij controleren onderzoek in ECD", http.StatusBadGateway)
		return
	}
	if !exists {
		http.Error(w, "Onderzoek bestaat niet in ECD", http.StatusBadRequest)
		return
	}

	// // 4. Check of diagnose bestaat voor onderzoek
	exists, err = service.DiagnoseExistsForOnderzoek(ecdURL, behoefte.OnderzoekID.String())
	if err != nil {
		http.Error(w, "Fout bij controleren diagnose in ECD", http.StatusBadGateway)
		return
	}
	if !exists {
		http.Error(w, "Onderzoek heeft nog geen diagnose in ECD, behoefte kan nog niet worden gemaakt", http.StatusBadRequest)
		return
	}

	// Alles klopt, sla behoefte op
	behoefte.ID = uuid.New()
	behoefte.Datum = time.Now()
	behoefte.Status = models.BehoefteNogNietDoorgestuurd

	if err := DB.Create(&behoefte).Error; err != nil {
		http.Error(w, "Fout bij opslaan in database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(behoefte)
}

func GetBehoefteByOnderzoekID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	onderzoekIDStr := vars["onderzoekId"]
	onderzoekID, err := uuid.Parse(onderzoekIDStr)
	if err != nil {
		http.Error(w, "Ongeldig OnderzoekID", http.StatusBadRequest)
		return
	}

	var gevondenBehoeften []models.Behoefte
	// Haal uit de database en preload de relaties
	if err := DB.Preload("Onderzoek").Preload("Client").
		Where("onderzoek_id = ?", onderzoekID).
		Find(&gevondenBehoeften).Error; err != nil {
		http.Error(w, "Fout bij ophalen uit database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(gevondenBehoeften) == 0 {
		http.Error(w, "Geen behoeften gevonden voor dit OnderzoekID", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gevondenBehoeften)
}

func GetBehoefteByClientNameAndBirthdate(w http.ResponseWriter, r *http.Request) {
	var behoefte models.Behoefte
	if err := json.NewDecoder(r.Body).Decode(&behoefte); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	if behoefte.Client.Naam == "" || behoefte.Client.Geboortedatum.IsZero() {
		http.Error(w, "ClientName en ClientBirthdate zijn verplicht", http.StatusBadRequest)
		return
	}

	var gevondenBehoeften []models.Behoefte
	if err := DB.Preload("Onderzoek").Preload("Client").
		Joins("JOIN clients ON clients.id = behoeftes.client_id").
		Where("clients.naam = ? AND clients.geboortedatum = ?", behoefte.Client.Naam, behoefte.Client.Geboortedatum).
		Find(&gevondenBehoeften).Error; err != nil {
		http.Error(w, "Fout bij ophalen uit database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(gevondenBehoeften) == 0 {
		http.Error(w, "Geen behoeften gevonden voor deze client", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gevondenBehoeften)
}

func GetBehoefteByClientID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientIDStr := vars["clientId"]
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		http.Error(w, "Ongeldig ClientID", http.StatusBadRequest)
		return
	}

	var gevondenBehoeften []models.Behoefte
	if err := DB.Preload("Onderzoek").Preload("Client").
		Where("client_id = ?", clientID).
		Find(&gevondenBehoeften).Error; err != nil {
		http.Error(w, "Fout bij ophalen uit database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(gevondenBehoeften) == 0 {
		http.Error(w, "Geen behoeften gevonden voor deze ClientID", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gevondenBehoeften)
}
