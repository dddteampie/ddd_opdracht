package api

import (
	"ecd/api/dto"
	"ecd/service"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
)

type Handler struct {
	ECD service.ECDService
}

func (h *Handler) GetClientHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
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
	if err := h.ECD.CreateClient(r.Context(), dto); err != nil {
		http.Error(w, "could not create client", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) CreateZorgdossierHandler(w http.ResponseWriter, r *http.Request) {
	var dto dto.ZorgdossierDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.ECD.CreateZorgdossier(r.Context(), dto); err != nil {
		http.Error(w, "could not create zorgdossier", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetZorgdossierByClientIDHandler(w http.ResponseWriter, r *http.Request) {
	clientIDStr := chi.URLParam(r, "clientId")
	clientID, err := uuid.FromString(clientIDStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}
	zorgdossier, err := h.ECD.GetZorgdossierByClientID(r.Context(), clientID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(zorgdossier)
}

func (h *Handler) CreateOnderzoekHandler(w http.ResponseWriter, r *http.Request) {
	var dto dto.OnderzoekDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.ECD.CreateOnderzoek(r.Context(), dto); err != nil {
		http.Error(w, "could not create onderzoek", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) AddAnamneseHandler(w http.ResponseWriter, r *http.Request) {
	onderzoekIDStr := chi.URLParam(r, "onderzoekId")
	onderzoekID, err := uuid.FromString(onderzoekIDStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}
	var dto dto.AnamneseDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.ECD.AddAnamnese(r.Context(), onderzoekID, dto); err != nil {
		http.Error(w, "could not add anamnese", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) AddMeetresultaatHandler(w http.ResponseWriter, r *http.Request) {
	onderzoekIDStr := chi.URLParam(r, "onderzoekId")
	onderzoekID, err := uuid.FromString(onderzoekIDStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}
	var dto dto.MeetresultaatDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.ECD.AddMeetresultaat(r.Context(), onderzoekID, dto); err != nil {
		http.Error(w, "could not add meetresultaat", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) AddDiagnoseHandler(w http.ResponseWriter, r *http.Request) {
	onderzoekIDStr := chi.URLParam(r, "onderzoekId")
	onderzoekID, err := uuid.FromString(onderzoekIDStr)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}
	var dto dto.DiagnoseDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.ECD.AddDiagnose(r.Context(), onderzoekID, dto); err != nil {
		http.Error(w, "could not add diagnose", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
