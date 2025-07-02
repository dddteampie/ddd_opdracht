package handlers

import (
	"aanvraagverwerking/helper"
	"aanvraagverwerking/models"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitHandlers(db *gorm.DB) {
	DB = db
}

func GetAanvraagByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aanvraagID := vars["id"]
	if aanvraagID == "" {
		http.Error(w, "Aanvraag ID is vereist", http.StatusBadRequest)
		return
	}

	var aanvraag models.Aanvraag
	if err := DB.Preload("Client").Preload("Behoefte").First(&aanvraag, "id = ?", aanvraagID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		} else {
			http.Error(w, "Fout bij ophalen aanvraag: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aanvraag)
}

func GetAanvragenByClientID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["clientId"]
	if clientID == "" {
		http.Error(w, "Client ID is vereist", http.StatusBadRequest)
		return
	}

	var aanvragen []models.Aanvraag
	if err := DB.Preload("Client").Preload("Behoefte").Where("client_id = ?", clientID).Find(&aanvragen).Error; err != nil {
		http.Error(w, "Fout bij ophalen aanvragen: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(aanvragen) == 0 {
		http.Error(w, "Geen aanvragen gevonden voor deze client", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(aanvragen)
}

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

	//Maak een random budget aan tussen 200 en 5000
	budget := helper.RandomFloat64Between(200, 5000)

	aanvraag := models.Aanvraag{
		ID:         uuid.New(),
		ClientID:   input.Client.ID,
		BehoefteID: input.Behoefte.ID,
		Client:     input.Client,
		Behoefte:   input.Behoefte,
		Status:     models.BehoefteOntvangen,
		Budget:     budget,
	}

	if err := DB.Create(&aanvraag).Error; err != nil {
		http.Error(w, "Fout bij opslaan in database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(aanvraag)
}

func StartCategorieAanvraag(w http.ResponseWriter, r *http.Request) {
	var input models.CategorieAanvraagDTO
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// Haal de aanvraag op, inclusief de behoefte
	var aanvraag models.Aanvraag
	// Zoek de aanvraag op basis van clientID en behoefteBeschrijving (aangenomen dat BehoefteBeschrijving uniek is voor de aanvraag)
	if err := DB.Preload("Behoefte").Joins("JOIN behoeftes ON behoeftes.id = aanvraags.behoefte_id").Where("aanvraags.client_id = ? AND behoeftes.beschrijving = ?", input.ClientID, input.BehoefteBeschrijving).First(&aanvraag).Error; err != nil {
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	jsonPayload, err := json.Marshal(input)
	if err != nil {
		http.Error(w, "Fout bij serialiseren JSON", http.StatusInternalServerError)
		return
	}
	url := "http://recommendation-service:8084/recommend/categorie/"
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		http.Error(w, "Fout bij bouwen request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	recResp, err := client.Do(req)
	if err != nil {
		log.Printf("Fout bij aanroepen recommendation service: %v", err)
		http.Error(w, "Fout bij aanroepen recommendation service", http.StatusBadGateway)
		return
	}
	defer recResp.Body.Close()

	// Zet de status op WachtenOpCategorieKeuze
	aanvraag.Status = models.WachtenOpCategorieKeuze
	if err := DB.Save(&aanvraag).Error; err != nil {
		log.Printf("Fout bij updaten aanvraagstatus: %v", err)
		http.Error(w, "Fout bij updaten aanvraagstatus", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Categorie-aanvraag gestart"))
}

// HaalPassendeCategorieenLijstOp haalt de lijst op voor een gegeven patientID
func HaalPassendeCategorieenLijstOp(w http.ResponseWriter, r *http.Request) {
	recommendationServiceURL := "http://recommendation-service:8084" 
	patientID := r.URL.Query().Get("patientId")
	if patientID == "" {
		http.Error(w, "patientId is verplicht", http.StatusBadRequest)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/recommend/categorie/?patientId=%s", recommendationServiceURL, patientID)
	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("kon geen request doen: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		http.Error(w, fmt.Sprintf("geen passende categorieënlijst gevonden voor patient %s", patientID), http.StatusNotFound)
		return
	}
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("onverwachte status van recommendation service: %s", resp.Status), http.StatusBadGateway)
		return
	}

	var lijst models.CategorieShortListDTO
	if err := json.NewDecoder(resp.Body).Decode(&lijst); err != nil {
		http.Error(w, fmt.Sprintf("kon response niet decoden: %v", err), http.StatusInternalServerError)
		return
	}

    var aanvraag models.Aanvraag
    if err := DB.Where("client_id = ?", patientID).First(&aanvraag).Error; err != nil {
        log.Printf("Aanvraag niet gevonden voor client_id %s: %v", patientID, err)
        http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
        return
    } else {
        var ids []int64
        for _, cat := range lijst.Categorielijst {
            ids = append(ids, int64(cat.ID))
        }
        aanvraag.CategorieOpties = ids
        if err := DB.Save(&aanvraag).Error; err != nil {
            log.Printf("Fout bij opslaan categorie-opties: %v", err)
        }
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lijst)
}

func KiesCategorie(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ClientID    uuid.UUID `json:"client_id"`
		BehoefteID  uuid.UUID `json:"behoefte_id"`
		CategorieID int       `json:"categorie"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Ongeldige input", http.StatusBadRequest)
		return
	}

	// Haal de aanvraag op
	var aanvraag models.Aanvraag
	if err := DB.Where("client_id = ? AND behoefte_id = ?", input.ClientID, input.BehoefteID).First(&aanvraag).Error; err != nil {
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

    // Controleer of de gekozen categorie in de opgeslagen opties zit
    gevonden := false
    for _, id := range aanvraag.CategorieOpties {
        if id == int64(input.CategorieID) {
            gevonden = true
            break
        }
    }
    if !gevonden {
        http.Error(w, "Gekozen categorie is niet toegestaan", http.StatusBadRequest)
        return
    }

	// Sla de gekozen categorie op
	aanvraag.GekozenCategorieID = &input.CategorieID
	aanvraag.Status = models.CategorieGekozen 

	if err := DB.Save(&aanvraag).Error; err != nil {
		log.Printf("Fout bij opslaan gekozen categorie: %v", err)
		http.Error(w, "Fout bij opslaan gekozen categorie", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Categorie succesvol gekozen"))
}

func StartProductAanvraag(w http.ResponseWriter, r *http.Request) {
    var input models.ProductAanvraagDTO
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }

    // Haal de aanvraag op
    var aanvraag models.Aanvraag
    // Zoek de aanvraag op basis van clientID en behoefteBeschrijving (aangenomen dat BehoefteBeschrijving uniek is voor de aanvraag)
	if err := DB.Preload("Behoefte").Joins("JOIN behoeftes ON behoeftes.id = aanvraags.behoefte_id").Where("aanvraags.client_id = ? AND behoeftes.beschrijving = ?", input.ClientID, input.BehoefteBeschrijving).First(&aanvraag).Error; err != nil {
		http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
		return
	}

	if aanvraag.GekozenCategorieID == nil || *aanvraag.GekozenCategorieID != *input.GekozenCategorieID {
		http.Error(w, "Categorie moet eerst gekozen zijn", http.StatusBadRequest)
		return
	}

    jsonPayload, err := json.Marshal(input)
    if err != nil {
        http.Error(w, "Fout bij serialiseren JSON", http.StatusInternalServerError)
        return
    }
    url := "http://recommendation-service:8084/recommend/oplossing/"
    client := &http.Client{Timeout: 10 * time.Second}
    req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        http.Error(w, "Fout bij bouwen request", http.StatusInternalServerError)
        return
    }
    req.Header.Set("Content-Type", "application/json")
    recResp, err := client.Do(req)
    if err != nil {
        log.Printf("Fout bij aanroepen recommendation service: %v", err)
        http.Error(w, "Fout bij aanroepen recommendation service", http.StatusBadGateway)
        return
    }
    defer recResp.Body.Close()

    if recResp.StatusCode != http.StatusOK && recResp.StatusCode != http.StatusCreated {
    http.Error(w, "Fout bij ophalen product aanbevelingen", recResp.StatusCode)
    return
}

    // Zet de status op WachtenOpProductKeuze
    aanvraag.Status = models.WachtenOpProductKeuze
    if err := DB.Save(&aanvraag).Error; err != nil {
        log.Printf("Fout bij updaten aanvraagstatus: %v", err)
    }

    w.WriteHeader(http.StatusAccepted)
    w.Write([]byte("Product-aanvraag gestart"))
}

func HaalPassendeProductenLijstOp(w http.ResponseWriter, r *http.Request) {
    recommendationServiceURL := "http://recommendation-service:8084"
    ClientID := r.URL.Query().Get("clientId")
    if ClientID == "" {
        http.Error(w, "ClientId is verplicht", http.StatusBadRequest)
        return
    }

    client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/recommend/oplossing/?clientId=%s", recommendationServiceURL, ClientID)
    resp, err := client.Get(url)
    if err != nil {
        http.Error(w, fmt.Sprintf("kon geen request doen: %v", err), http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        http.Error(w, fmt.Sprintf("geen passende productlijst gevonden voor patient %s", ClientID), http.StatusNotFound)
        return
    }
    if resp.StatusCode != http.StatusOK {
        http.Error(w, fmt.Sprintf("onverwachte status van recommendation service: %s", resp.Status), http.StatusBadGateway)
        return
    }

    var lijst models.ProductShortListDTO
    if err := json.NewDecoder(resp.Body).Decode(&lijst); err != nil {
        http.Error(w, fmt.Sprintf("kon response niet decoden: %v", err), http.StatusInternalServerError)
        return
    }

    var aanvraag models.Aanvraag
    if err := DB.Where("client_id = ?", ClientID).First(&aanvraag).Error; err != nil {
        log.Printf("Aanvraag niet gevonden voor client_id %s: %v", ClientID, err)
        http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
        return
    } else {
        var ids []int64
        for _, cat := range lijst.Productlijst {
            ids = append(ids, int64(cat.EAN))
        }
        aanvraag.ProductOpties = ids
        if err := DB.Save(&aanvraag).Error; err != nil {
            log.Printf("Fout bij opslaan producten-opties: %v", err)
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(lijst)
}

func KiesProduct(w http.ResponseWriter, r *http.Request) {
    var input struct {
        ClientID   uuid.UUID `json:"client_id"`
        BehoefteID uuid.UUID `json:"behoefte_id"`
        ProductEAN int64     `json:"product_ean"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Ongeldige input", http.StatusBadRequest)
        return
    }

    // Haal de aanvraag op
    var aanvraag models.Aanvraag
    if err := DB.Where("client_id = ? AND behoefte_id = ?", input.ClientID, input.BehoefteID).First(&aanvraag).Error; err != nil {
        http.Error(w, "Aanvraag niet gevonden", http.StatusNotFound)
        return
    }

    // Controleer of het gekozen product in de opgeslagen opties zit
    gevonden := false
    for _, ean := range aanvraag.ProductOpties {
        if ean == input.ProductEAN {
            gevonden = true
            break
        }
    }
    if !gevonden {
        http.Error(w, "Gekozen product is niet toegestaan", http.StatusBadRequest)
        return
    }

    // Sla het gekozen product op
    aanvraag.GekozenProductID = &input.ProductEAN
    aanvraag.Status = models.ProductGekozen 

    if err := DB.Save(&aanvraag).Error; err != nil {
        log.Printf("Fout bij opslaan gekozen product: %v", err)
        http.Error(w, "Fout bij opslaan gekozen product", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Product succesvol gekozen"))
}