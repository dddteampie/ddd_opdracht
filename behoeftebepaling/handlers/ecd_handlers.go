package handlers

import (
	"behoeftebepaling/models"
	"behoeftebepaling/service"
	//"bytes"
	"encoding/json"
	//"io"
	"net/http"

	"github.com/gorilla/mux"
)

var ecdURL string

func SetECDURL(url string) {
    ecdURL = url
}

func KoppelClientHandler(w http.ResponseWriter, r *http.Request) {
    var client models.ClientDTO
    if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }
    err := service.CreateClientInECD(ecdURL, client)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Client succesvol aangemaakt in ECD"))
}

func KoppelZorgdossierHandler(w http.ResponseWriter, r *http.Request) {
    var zorgdossier models.ZorgdossierDTO
    if err := json.NewDecoder(r.Body).Decode(&zorgdossier); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }
    err := service.CreateZorgdossierInECD(ecdURL, zorgdossier)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Zorgdossier succesvol aangemaakt in ECD"))
}

func KoppelOnderzoekHandler(w http.ResponseWriter, r *http.Request) {
    var onderzoek models.OnderzoekDTO
    if err := json.NewDecoder(r.Body).Decode(&onderzoek); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }
    err := service.CreateOnderzoekInECD(ecdURL, onderzoek)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Onderzoek succesvol aangemaakt in ECD"))
}

func KoppelAnamneseHandler(w http.ResponseWriter, r *http.Request) {
    var anamnese models.AnamneseDTO
    if err := json.NewDecoder(r.Body).Decode(&anamnese); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    onderzoekId := vars["onderzoekId"]

    // Gebruik de service-functie (met retry)
    err := service.AddAnamneseToECD(ecdURL, onderzoekId, anamnese)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Anamnese succesvol opgeslagen in ECD"))
}

func KoppelMeetresultaatHandler(w http.ResponseWriter, r *http.Request) {
    var meetresultaat models.MeetresultaatDTO
    if err := json.NewDecoder(r.Body).Decode(&meetresultaat); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    onderzoekId := vars["onderzoekId"]

    err := service.AddMeetresultaatToECD(ecdURL, onderzoekId, meetresultaat)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Meetresultaat succesvol opgeslagen in ECD"))
}

func KoppelDiagnoseHandler(w http.ResponseWriter, r *http.Request) {
    var diagnose models.DiagnoseDTO
    if err := json.NewDecoder(r.Body).Decode(&diagnose); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    onderzoekId := vars["onderzoekId"]

    err := service.AddDiagnoseToECD(ecdURL, onderzoekId, diagnose)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Diagnose succesvol opgeslagen in ECD"))
}