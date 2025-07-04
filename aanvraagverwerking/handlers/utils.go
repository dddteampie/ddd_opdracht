package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	//"strings"

	"aanvraagverwerking/helper"
	"aanvraagverwerking/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var recURL string

func SetRECURL(url string) {
	recURL = url
}

// startaanvraag utils
func DecodeAanvraagInput(r *http.Request) (models.Client, models.Behoefte, error) {
	var input struct {
		Client   models.Client   `json:"client"`
		Behoefte models.Behoefte `json:"behoefte"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return models.Client{}, models.Behoefte{}, err
	}
	return input.Client, input.Behoefte, nil
}

func AanvraagBestaat(db *gorm.DB, clientID, behoefteID uuid.UUID) (bool, error) {
	var bestaandeAanvraag models.Aanvraag
	err := db.Where("client_id = ? AND behoefte_id = ?", clientID, behoefteID).First(&bestaandeAanvraag).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func BouwAanvraag(client models.Client, behoefte models.Behoefte) models.Aanvraag {
	return models.Aanvraag{
		ID:         uuid.New(),
		ClientID:   client.ID,
		BehoefteID: behoefte.ID,
		Client:     client,
		Behoefte:   behoefte,
		Status:     models.BehoefteOntvangen,
		Budget:     helper.RandomFloat64Between(200, 5000),
	}
}

func SlaAanvraagOp(db *gorm.DB, aanvraag models.Aanvraag) error {
	return db.Create(&aanvraag).Error
}

// CategorieAanvraag utils
func DecodeCategorieAanvraagInput(r *http.Request) (models.CategorieAanvraagDTO, error) {
	var input models.CategorieAanvraagDTO
	err := json.NewDecoder(r.Body).Decode(&input)
	return input, err
}

func VindAanvraagMetBehoefte(db *gorm.DB, clientID uuid.UUID, behoefteBeschrijving string) (models.Aanvraag, error) {
	var aanvraag models.Aanvraag
	err := db.Preload("Behoefte").
		Joins("JOIN behoeftes ON behoeftes.id = aanvraags.behoefte_id").
		Where("aanvraags.client_id = ? AND behoeftes.beschrijving = ?", clientID, behoefteBeschrijving).
		First(&aanvraag).Error
	return aanvraag, err
}

func StuurCategorieAanvraagNaarRecommendation(input models.CategorieAanvraagDTO) error {
	jsonPayload, err := json.Marshal(input)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/recommend/categorie/", recURL)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func ZetStatusWachtenOpCategorie(db *gorm.DB, aanvraag *models.Aanvraag) error {
	// Controleer of de aanvraag bestaat
	var bestaand models.Aanvraag
	if err := db.First(&bestaand, "id = ?", aanvraag.ID).Error; err != nil {
		return err // Geeft ErrRecordNotFound terug als hij niet bestaat
	}
	aanvraag.Status = models.WachtenOpCategorieKeuze
	return db.Save(aanvraag).Error
}

func VraagCategorieenLijstOp(patientID string) (models.CategorieShortListDTO, int, error) {
	url := fmt.Sprintf("%s/recommend/categorie/?patientId=%s", recURL, patientID)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return models.CategorieShortListDTO{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return models.CategorieShortListDTO{}, http.StatusNotFound, fmt.Errorf("geen passende categorieënlijst gevonden")
	}
	if resp.StatusCode != http.StatusOK {
		return models.CategorieShortListDTO{}, resp.StatusCode, fmt.Errorf("onverwachte status van recommendation service")
	}

	var lijst models.CategorieShortListDTO
	if err := json.NewDecoder(resp.Body).Decode(&lijst); err != nil {
		return models.CategorieShortListDTO{}, 0, fmt.Errorf("kon response niet decoden: %v", err)
	}
	return lijst, http.StatusOK, nil
}

func KoppelCategorieOptiesAanAanvraag(db *gorm.DB, patientID string, lijst models.CategorieShortListDTO) error {
	var aanvraag models.Aanvraag
	if err := db.Where("client_id = ?", patientID).First(&aanvraag).Error; err != nil {
		return err
	}
	var ids []int64
	for _, cat := range lijst.Categorielijst {
		ids = append(ids, int64(cat.ID))
	}
	aanvraag.CategorieOpties = ids
	return db.Save(&aanvraag).Error
}

// KiesCategorie utils
func haalAanvraagOp(db *gorm.DB, clientID, behoefteID uuid.UUID) (models.Aanvraag, error) {
	var aanvraag models.Aanvraag
	err := db.Where("client_id = ? AND behoefte_id = ?", clientID, behoefteID).First(&aanvraag).Error
	return aanvraag, err
}

func categorieToegestaan(opties []int64, gekozen int) bool {
	for _, id := range opties {
		if id == int64(gekozen) {
			return true
		}
	}
	return false
}

func slaGekozenCategorieOp(db *gorm.DB, aanvraag *models.Aanvraag, categorieID int) error {
	aanvraag.GekozenCategorieID = &categorieID
	aanvraag.Status = models.CategorieGekozen
	return db.Save(aanvraag).Error
}

// Startproductaanvraag utils
func validateProductAanvraagInput(input models.ProductAanvraagDTO) error {
	if input.ClientID == "" || input.BehoefteBeschrijving == "" || input.GekozenCategorieID == nil {
		return fmt.Errorf("verplichte velden ontbreken of zijn ongeldig")
	}
	// Eventueel: check of ClientID een geldige UUID is
	if _, err := uuid.Parse(input.ClientID); err != nil {
		return fmt.Errorf("client_id is geen geldige UUID")
	}
	return nil
}

func haalProductAanvraagOp(db *gorm.DB, clientID uuid.UUID, behoefteBeschrijving string) (models.Aanvraag, error) {
	var aanvraag models.Aanvraag
	err := db.Preload("Behoefte").
		Joins("JOIN behoeftes ON behoeftes.id = aanvraags.behoefte_id").
		Where("aanvraags.client_id = ? AND behoeftes.beschrijving = ?", clientID, behoefteBeschrijving).
		First(&aanvraag).Error
	return aanvraag, err
}

func stuurProductAanvraagNaarRecommendation(input models.ProductAanvraagDTO) error {
	jsonPayload, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("fout bij serialiseren JSON: %v", err)
	}
	url := fmt.Sprintf("%s/recommend/oplossing/", recURL)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("fout bij bouwen request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	recResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fout bij aanroepen recommendation service: %v", err)
	}
	defer recResp.Body.Close()

	if recResp.StatusCode != http.StatusOK && recResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("fout bij ophalen product aanbevelingen: status %d", recResp.StatusCode)
	}
	return nil
}

// VraagProductenLijstOp utils
func vraagProductenLijstOp(clientID string) (models.ProductShortListDTO, int, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/recommend/oplossing/?clientId=%s", recURL, clientID)
	resp, err := client.Get(url)
	if err != nil {
		return models.ProductShortListDTO{}, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return models.ProductShortListDTO{}, http.StatusNotFound, fmt.Errorf("niet gevonden")
	}
	if resp.StatusCode != http.StatusOK {
		return models.ProductShortListDTO{}, resp.StatusCode, fmt.Errorf("onverwachte status")
	}

	var lijst models.ProductShortListDTO
	if err := json.NewDecoder(resp.Body).Decode(&lijst); err != nil {
		return models.ProductShortListDTO{}, 0, err
	}
	return lijst, http.StatusOK, nil
}

func koppelProductOptiesAanAanvraag(db *gorm.DB, clientID string, lijst models.ProductShortListDTO) error {
	var aanvraag models.Aanvraag
	if err := db.Where("client_id = ?", clientID).First(&aanvraag).Error; err != nil {
		return err
	}
	var ids []int64
	for _, prod := range lijst.Productlijst {
		ids = append(ids, int64(prod.EAN))
	}
	aanvraag.ProductOpties = ids
	return db.Save(&aanvraag).Error
}

// KiesProduct utils
func productToegestaan(productOpties []int64, gekozenEAN int64) bool {
	for _, ean := range productOpties {
		if ean == gekozenEAN {
			return true
		}
	}
	return false
}

func slaGekozenProductOp(db *gorm.DB, aanvraag *models.Aanvraag, gekozenEAN int64) error {
	aanvraag.GekozenProductID = &gekozenEAN
	aanvraag.Status = models.ProductGekozen
	return db.Save(aanvraag).Error
}
