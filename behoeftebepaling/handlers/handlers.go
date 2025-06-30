package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"behoeftebepaling/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var behoeften []models.Behoefte

func CreateBehoefte(w http.ResponseWriter, r *http.Request) {
    var behoefte models.Behoefte
    if err := json.NewDecoder(r.Body).Decode(&behoefte); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }

    // Controleer of OnderzoekID is meegegeven
    if behoefte.OnderzoekID == uuid.Nil {
        http.Error(w, "OnderzoekID is verplicht", http.StatusBadRequest)
        return
    }

    behoefte.ID = uuid.New()
    behoefte.Datum = time.Now()
    behoeften = append(behoeften, behoefte)

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
    for _, b := range behoeften {
        if b.OnderzoekID == onderzoekID {
            gevondenBehoeften = append(gevondenBehoeften, b)
        }
    }

    if len(gevondenBehoeften) == 0 {
        http.Error(w, "Geen behoeften gevonden voor dit OnderzoekID", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(gevondenBehoeften)
}

