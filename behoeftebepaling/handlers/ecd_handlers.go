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

//Lokale ECD voor testdoeleinden
// In een productieomgeving zou dit de URL van het echte ECD zijn die uit .env variabelen of een configuratiebestand zou komen
var ecdURL = "http://host.docker.internal:8090"

// func KoppelAnamneseHandler(w http.ResponseWriter, r *http.Request) {
//     var anamnese models.AnamneseDTO
//     if err := json.NewDecoder(r.Body).Decode(&anamnese); err != nil {
//         http.Error(w, "Ongeldige input", http.StatusBadRequest)
//         return
//     }

// 	vars := mux.Vars(r)
//     onderzoekId := vars["onderzoekId"]

//     err := service.AddAnamneseToECD(ecdURL, onderzoekId, anamnese)
//     if err != nil {
//         http.Error(w, err.Error(), http.StatusBadGateway)
//         return
//     }

//     w.WriteHeader(http.StatusCreated)
//     w.Write([]byte("Anamnese succesvol opgeslagen in ECD"))
// }

func KoppelAnamneseHandler(w http.ResponseWriter, r *http.Request) {
    var anamnese models.AnamneseDTO
    if err := json.NewDecoder(r.Body).Decode(&anamnese); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    onderzoekId := vars["onderzoekId"]

    // Gebruik de service-functie (met retry)
    ecdURL := "http://ecd-service:8082/api"
    err := service.AddAnamneseToECD(ecdURL, onderzoekId, anamnese)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Write([]byte("Anamnese succesvol opgeslagen in ECD"))
}

// func KoppelAnamneseHandler(w http.ResponseWriter, r *http.Request) {
//     var anamnese models.AnamneseDTO
//     if err := json.NewDecoder(r.Body).Decode(&anamnese); err != nil {
//         http.Error(w, "Ongeldige input", http.StatusBadRequest)
//         return
//     }

//     vars := mux.Vars(r)
//     onderzoekId := vars["onderzoekId"]

//     // Maak de URL naar de ECD-service
//     ecdURL := "http://ecd-service:8082/api/onderzoek/" + onderzoekId + "/anamnese"

//     // Encode de anamnese als JSON
//     body, err := json.Marshal(anamnese)
//     if err != nil {
//         http.Error(w, "Fout bij encoden van data", http.StatusInternalServerError)
//         return
//     }

//     // Doe de POST request naar de ECD-service
//     resp, err := http.Post(ecdURL, "application/json", bytes.NewBuffer(body))
//     if err != nil {
//         http.Error(w, "ECD-service niet bereikbaar: "+err.Error(), http.StatusBadGateway)
//         return
//     }
//     defer resp.Body.Close()

//     // Geef de response van de ECD-service door
//     w.WriteHeader(resp.StatusCode)
//     io.Copy(w, resp.Body)
// }

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


// func GetOnderzoekenVanPatiëntHandler(w http.ResponseWriter, r *http.Request) {
//     vars := mux.Vars(r)
//     patientId := vars["patientId"]

//     onderzoeken, err := service.GetOnderzoekenVanCliënt(ecdURL, patientId)
//     if err != nil {
//         http.Error(w, err.Error(), http.StatusBadGateway)
//         return
//     }

//     w.Header().Set("Content-Type", "application/json")
//     json.NewEncoder(w).Encode(onderzoeken)
// }