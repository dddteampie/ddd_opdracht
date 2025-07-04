package handlers

import (
	"behoeftebepaling/helper"
	"behoeftebepaling/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func StuurBehoefteNaarAanvraagverwerking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	behoefteID := vars["behoefteId"]

	// Haal behoefte op
	var behoefte models.Behoefte
	if err := DB.First(&behoefte, "id = ?", behoefteID).Error; err != nil {
		http.Error(w, "Behoefte niet gevonden", http.StatusNotFound)
		return
	}

	// Haal client op
	var client models.Client
	if err := DB.First(&client, "id = ?", behoefte.ClientID).Error; err != nil {
		http.Error(w, "Client niet gevonden", http.StatusNotFound)
		return
	}

	// Roep de helper aan
	aanvraagID, err := helper.NotifyAanvraagverwerking(behoefte, client)
	if err != nil {
		http.Error(w, "Fout bij doorsturen naar aanvraagverwerking: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Update status van de behoefte
	behoefte.Status = models.BehoefteDoorgestuurd
	if err := DB.Save(&behoefte).Error; err != nil {
		http.Error(w, "Behoefte doorgestuurd, maar status kon niet worden bijgewerkt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      aanvraagID,
		"message": "Behoefte doorgestuurd naar aanvraagverwerking",
	})
}