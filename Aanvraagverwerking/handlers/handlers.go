package handlers

import (
	"encoding/json"
	"net/http"
	//"time"

	"aanvraagverwerking/models"
	//"aanvraagverwerking/service"

	"github.com/google/uuid"
	//"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitHandlers(db *gorm.DB) {
	DB = db
}

// func StartAanvraag(w http.ResponseWriter, r *http.Request) {
//     var aanvraag models.Aanvraag
//     if err := json.NewDecoder(r.Body).Decode(&aanvraag); err != nil {
//         http.Error(w, "Ongeldige input", http.StatusBadRequest)
//         return
//     }

//     aanvraag.ID = uuid.New() // Genereer een random UUID

//     // Sla de aanvraag op in de database (optioneel)
//     if err := DB.Create(&aanvraag).Error; err != nil {
//         http.Error(w, "Fout bij opslaan in database: "+err.Error(), http.StatusInternalServerError)
//         return
//     }

//     w.WriteHeader(http.StatusCreated)
//     json.NewEncoder(w).Encode(aanvraag)
// }

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

    aanvraag := models.Aanvraag{
        ID:         uuid.New(),
        ClientID:   input.Client.ID,
        BehoefteID: input.Behoefte.ID,
        Client:     input.Client,
        Behoefte:   input.Behoefte,
        Status:     models.BehoefteOntvangen,
    }

    if err := DB.Create(&aanvraag).Error; err != nil {
        http.Error(w, "Fout bij opslaan in database: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(aanvraag)
}