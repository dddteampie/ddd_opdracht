package api

import (
	"ecd/api/dto"
	"ecd/service"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
)

type Handler struct {
	ECD service.ECDService
}

func (h *Handler) GetClientHandler(w http.ResponseWriter, r *http.Request) {
	var id, err = GetUUIDFromRequest(r, "id")
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	client, err := h.ECD.GetClient(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(client)
}

func (h *Handler) CreateClientHandler(w http.ResponseWriter, r *http.Request) {
	var dto dto.ClientDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	dto.ID = uuid.Must(uuid.NewV4()) // Generate a new UUID for the client
	if dto.Naam == "" || dto.Adres == "" || dto.Geboortedatum.IsZero() {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	if err := h.ECD.CreateClient(r.Context(), dto); err != nil {
		http.Error(w, "could not create client", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clientId": dto.ID,
	})
}

func (h *Handler) CreateZorgdossierHandler(w http.ResponseWriter, r *http.Request) {
	var dto dto.ZorgdossierDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if dto.ClientID == uuid.Nil {
		http.Error(w, "client ID is required", http.StatusBadRequest)
		return
	}
	dto.ID = uuid.Must(uuid.NewV4())
	if err := h.ECD.CreateZorgdossier(r.Context(), dto); err != nil {
		http.Error(w, "could not create zorgdossier", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"zorgdossierId": dto.ID,
	})
}

func (h *Handler) GetZorgdossierByClientIDHandler(w http.ResponseWriter, r *http.Request) {
	var clientID, err = GetUUIDFromRequest(r, "clientId")
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	zorgdossier, err := h.ECD.GetZorgdossierByClientID(r.Context(), clientID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(zorgdossier)
}

func (h *Handler) CreateOnderzoekHandler(w http.ResponseWriter, r *http.Request) {
	var dto dto.OnderzoekDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		fmt.Print("Error decoding request body:", err)
		return
	}
	if dto.ZorgdossierID == uuid.Nil {
		http.Error(w, "zorgdossier ID is required", http.StatusBadRequest)
		return
	}
	dto.ID = uuid.Must(uuid.NewV4())
	if err := h.ECD.CreateOnderzoek(r.Context(), dto); err != nil {
		http.Error(w, "could not create onderzoek", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"onderzoekId": dto.ID,
	})
}

func (h *Handler) GetOnderzoekByIDHandler(w http.ResponseWriter, r *http.Request) {
	var onderzoekID, err = GetUUIDFromRequest(r, "onderzoekId")
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	onderzoek, err := h.ECD.GetOnderzoekByID(r.Context(), onderzoekID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(onderzoek)
}

func (h *Handler) UpdateOnderzoekHandler(w http.ResponseWriter, r *http.Request) {
	var onderzoekID, err = GetUUIDFromRequest(r, "onderzoekId")
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var dto dto.OnderzoekDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	dto.ID = onderzoekID // Ensure the ID is set to the correct onderzoek ID
	if err := h.ECD.UpdateOnderzoek(r.Context(), dto); err != nil {
		http.Error(w, "could not update onderzoek", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AddAnamneseHandler(w http.ResponseWriter, r *http.Request) {
	var onderzoekID, err = GetUUIDFromRequest(r, "onderzoekId")
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var dto dto.AnamneseDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	dto.ID = uuid.Must(uuid.NewV4())
	if err := h.ECD.AddAnamnese(r.Context(), onderzoekID, dto); err != nil {
		http.Error(w, "could not add anamnese", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) AddMeetresultaatHandler(w http.ResponseWriter, r *http.Request) {
	var onderzoekID, err = GetUUIDFromRequest(r, "onderzoekId")
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var dto dto.MeetresultaatDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	dto.ID = uuid.Must(uuid.NewV4())
	if err := h.ECD.AddMeetresultaat(r.Context(), onderzoekID, dto); err != nil {
		http.Error(w, "could not add meetresultaat", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) AddDiagnoseHandler(w http.ResponseWriter, r *http.Request) {
	var onderzoekID, err = GetUUIDFromRequest(r, "onderzoekId")
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var dto dto.DiagnoseDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	dto.ID = uuid.Must(uuid.NewV4())
	if err := h.ECD.AddDiagnose(r.Context(), onderzoekID, dto); err != nil {
		http.Error(w, "could not add diagnose", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetUUIDFromRequest(r *http.Request, id string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, id)
	if idStr == "" {
		return uuid.Nil, fmt.Errorf("id parameter is required")
	}
	uuID, err := uuid.FromString(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid format: %w", err)
	}
	return uuID, nil
}
