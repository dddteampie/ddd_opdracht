package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"recommendation/service"
	"strings"
)

type MaakOplossingenLijstRequest struct {
	ClientID    string  `json:"clientId"`
	Budget      float64 `json:"budget"`
	Behoeften   string  `json:"behoeften"`
	CategorieID *int    `json:"CategorieID,omitempty"`
}

type MaakPassendeCategorieënLijstRequest struct {
	PatientID string  `json:"patientId"`
	Budget    float64 `json:"budget"`
	Behoeften string  `json:"behoeften"`
}

type AanbevelingsHandler struct {
	svc service.IAanbevelingHelpers
}

func NewAanbevelingsHandler(svc service.IAanbevelingHelpers) *AanbevelingsHandler {
	return &AanbevelingsHandler{svc: svc}
}

func (h *AanbevelingsHandler) MaakPassendeCategorieënLijstHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("MaakPassendeCategorieënLijstHandler: Ontvangen PUT request.")
	if r.Method != http.MethodPut {
		http.Error(w, "Methode niet toegestaan", http.StatusMethodNotAllowed)
		return
	}

	var reqBody MaakPassendeCategorieënLijstRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Printf("MaakPassendeCategorieënLijstHandler: Ongeldige request body: %v", err)
		http.Error(w, "Ongeldige request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if reqBody.PatientID == "" || reqBody.Budget <= 0 || reqBody.Behoeften == "" {
		log.Printf("MaakPassendeCategorieënLijstHandler: Ontbrekende of ongeldige parameters. PatientID: '%s', Budget: %.2f, Behoeften: '%s'", reqBody.PatientID, reqBody.Budget, reqBody.Behoeften)
		http.Error(w, "Ontbrekende of ongeldige parameters: patientId, budget, of behoeften", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	passendeLijst, err := h.svc.MaakPassendeCategorieënLijst(ctx, reqBody.PatientID, reqBody.Budget, reqBody.Behoeften)
	if err != nil {
		log.Printf("MaakPassendeCategorieënLijstHandler: Fout bij het maken van passende categorieënlijst via service: %v", err)
		http.Error(w, "Interne serverfout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(passendeLijst)
	log.Println("MaakPassendeCategorieënLijstHandler: Succesvol verwerkt.")
}

func (h *AanbevelingsHandler) MaakOplossingenLijstHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("MaakOplossingenLijstHandler: Ontvangen PUT request.")
	if r.Method != http.MethodPut {
		http.Error(w, "Methode niet toegestaan", http.StatusMethodNotAllowed)
		return
	}

	var reqBody MaakOplossingenLijstRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		log.Printf("MaakOplossingenLijstHandler: Ongeldige request body: %v", err)
		http.Error(w, "Ongeldige request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if reqBody.ClientID == "" || reqBody.Budget <= 0 || reqBody.Behoeften == "" {
		log.Printf("MaakOplossingenLijstHandler: Ontbrekende of ongeldige parameters. ClientID: '%s', Budget: %.2f, Behoeften: '%s'", reqBody.ClientID, reqBody.Budget, reqBody.Behoeften)
		http.Error(w, "Ontbrekende of ongeldige parameters: clientId, budget, of behoeften", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	oplossingenLijst, err := h.svc.MaakOplossingenLijst(ctx, reqBody.ClientID, reqBody.Budget, reqBody.Behoeften, reqBody.CategorieID)
	if err != nil {
		log.Printf("MaakOplossingenLijstHandler: Fout bij het maken van oplossingenlijst via service: %v", err)
		http.Error(w, "Interne serverfout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(oplossingenLijst)
	log.Println("MaakOplossingenLijstHandler: Succesvol verwerkt.")
}

func (h *AanbevelingsHandler) HaalPassendeCategorieënLijstOpHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HaalPassendeCategorieënLijstOpHandler: Ontvangen GET request.")
	patientID := r.URL.Query().Get("patientId")
	if patientID == "" {
		http.Error(w, "Ontbrekende parameter: patientId", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	passendeLijst, err := h.svc.HaalPassendeCategorieënLijstOp(ctx, patientID)
	if err != nil {
		if strings.Contains(err.Error(), "geen passende categorieënlijst gevonden") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			log.Printf("HaalPassendeCategorieënLijstOpHandler: Fout bij het ophalen van passende categorieënlijst: %v", err)
			http.Error(w, "Interne serverfout", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(passendeLijst)
	log.Println("MaakPassendeCategorieënLijstHandler: Succesvol verwerkt.")
}

func (h *AanbevelingsHandler) HaalOplossingenLijstOpHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HaalOplossingenLijstOpHandler: Ontvangen GET request.")
	clientID := r.URL.Query().Get("clientId")
	if clientID == "" {
		http.Error(w, "Ontbrekende parameter: clientId", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	oplossingenLijst, err := h.svc.HaalOplossingenLijstOp(ctx, clientID)
	if err != nil {
		if strings.Contains(err.Error(), "geen oplossingenlijst gevonden") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			log.Printf("HaalOplossingenLijstOpHandler: Fout bij het ophalen van oplossingenlijst: %v", err)
			http.Error(w, "Interne serverfout", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oplossingenLijst)
	log.Println("HaalOplossingenLijstOpHandler: Succesvol verwerkt.")
}

func (h *AanbevelingsHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Methode niet toegestaan", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	log.Println("HealthCheckHandler: Succesvol verwerkt.")
}
