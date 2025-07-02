package handlers

import (
	"behoeftebepaling/models"
	"behoeftebepaling/service"
	"encoding/json"
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

    // Maak client aan in ECD en ontvang het ECD-ID
    ecdID, err := service.CreateClientInECDAndReturnID(ecdURL, client)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    // Sla client op in eigen database met ECD-ID als ID
    newClient := models.Client{
        ID:            ecdID,
        Naam:          client.Naam,
        Adres:         client.Adres,
        Geboortedatum: client.Geboortedatum,
    }
    if err := DB.Create(&newClient).Error; err != nil {
        http.Error(w, "Fout bij opslaan client in eigen database: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "clientId": ecdID,
    })
}

func GetClientHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    clientId := vars["clientId"]

    client, err := service.GetClientFromECD(ecdURL, clientId)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(client)
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

func GetZorgdossierByClientIDHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    client_id := vars["clientId"]

    zorgdossier, err := service.GetZorgdossierFromECD(ecdURL, client_id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(zorgdossier)
}

func KoppelOnderzoekHandler(w http.ResponseWriter, r *http.Request) {
    var onderzoek models.OnderzoekDTO
    if err := json.NewDecoder(r.Body).Decode(&onderzoek); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }

    // Maak onderzoek aan in ECD en ontvang het ECD-onderzoekID
    ecdID, err := service.CreateOnderzoekInECDAndReturnID(ecdURL, onderzoek)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    // Sla onderzoek op in eigen database met ECD-ID als ID
    newOnderzoek := models.Onderzoek{
        ID:            ecdID,
        ZorgdossierId: onderzoek.ZorgdossierID,
        BeginDatum:    onderzoek.BeginDatum,
        EindDatum:     onderzoek.EindDatum,
    }
    if err := DB.Create(&newOnderzoek).Error; err != nil {
        http.Error(w, "Fout bij opslaan onderzoek in eigen database: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "onderzoekId": ecdID,
    })
}

func GetOnderzoekByIdHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    onderzoekId := vars["onderzoekId"]

    onderzoek, err := service.GetOnderzoekByID(ecdURL, onderzoekId)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(onderzoek)
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