package service

import (
	"behoeftebepaling/helper"
	"behoeftebepaling/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func AddAnamneseToECD(ecdURL, onderzoekId string, anamnese models.AnamneseDTO) error {
	url := fmt.Sprintf("%s/onderzoek/%s/anamnese", ecdURL, onderzoekId)
	return helper.PostJSONWithRetry(url, anamnese, http.StatusCreated)
}

func AddMeetresultaatToECD(ecdURL, onderzoekId string, meetresultaat models.MeetresultaatDTO) error {
	url := fmt.Sprintf("%s/onderzoek/%s/meetresultaat", ecdURL, onderzoekId)
	return helper.PostJSONWithRetry(url, meetresultaat, http.StatusCreated)
}

func AddDiagnoseToECD(ecdURL, onderzoekId string, diagnose models.DiagnoseDTO) error {
	url := fmt.Sprintf("%s/onderzoek/%s/diagnose", ecdURL, onderzoekId)
	return helper.PostJSONWithRetry(url, diagnose, http.StatusCreated)
}

func GetClientFromECD(ecdURL string, clientID string) (*models.ClientDTO, error) {
	url := fmt.Sprintf("%s/client/%s", ecdURL, clientID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client ophalen uit ECD mislukt, status: %d", resp.StatusCode)
	}

	var client models.ClientDTO
	if err := json.NewDecoder(resp.Body).Decode(&client); err != nil {
		return nil, err
	}

	return &client, nil
}

// Check of client bestaat in ECD
func ClientExistsInECD(ecdURL string, clientID string) (bool, error) {
	url := fmt.Sprintf("%s/client/%s", ecdURL, clientID)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func CreateClientInECDAndReturnID(ecdURL string, client models.ClientDTO) (uuid.UUID, error) {
	url := fmt.Sprintf("%s/client", ecdURL)
	body, err := json.Marshal(client)
	if err != nil {
		return uuid.Nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return uuid.Nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("client aanmaken in ECD mislukt, status: %d", resp.StatusCode)
	}

	// Verwacht een response zoals: { "clientId": "..." }
	var result struct {
		ClientID uuid.UUID `json:"clientId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return uuid.Nil, fmt.Errorf("kan clientId niet uitlezen uit ECD-response: %w", err)
	}
	return result.ClientID, nil
}

// Check of zorgdossier bestaat voor client
func ZorgdossierExistsForClient(ecdURL string, clientID string) (bool, error) {
	url := fmt.Sprintf("%s/zorgdossier/client/%s", ecdURL, clientID)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

// Maak een nieuw zorgdossier aan in het ECD en retourneer het ID
func CreateZorgdossierInECDAndReturnID(ecdURL string, zorgdossier models.ZorgdossierDTO) (uuid.UUID, error) {
	url := fmt.Sprintf("%s/zorgdossier", ecdURL)
	body, err := json.Marshal(zorgdossier)
	if err != nil {
		return uuid.Nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return uuid.Nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("zorgdossier aanmaken in ECD mislukt, status: %d", resp.StatusCode)
	}

	// Verwacht een response zoals: { "id": "..." }
	var result struct {
		ID uuid.UUID `json:"zorgdossierId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return uuid.Nil, fmt.Errorf("kan id niet uitlezen uit ECD-response: %w", err)
	}
	return result.ID, nil
}

func GetZorgdossierFromECD(ecdURL string, clientID string) (*models.ZorgdossierDTO, error) {
	url := fmt.Sprintf("%s/zorgdossier/client/%s", ecdURL, clientID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zorgdossier ophalen uit ECD mislukt, status: %d", resp.StatusCode)
	}

	var zorgdossier models.ZorgdossierDTO
	if err := json.NewDecoder(resp.Body).Decode(&zorgdossier); err != nil {
		return nil, err
	}

	return &zorgdossier, nil
}

// Check of onderzoek bestaat
func OnderzoekExists(ecdURL string, onderzoekID string) (bool, error) {
	url := fmt.Sprintf("%s/onderzoek/%s", ecdURL, onderzoekID)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func CreateOnderzoekInECDAndReturnID(ecdURL string, onderzoek models.OnderzoekDTO) (uuid.UUID, error) {
	url := fmt.Sprintf("%s/onderzoek", ecdURL)
	body, err := json.Marshal(onderzoek)
	if err != nil {
		return uuid.Nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return uuid.Nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("onderzoek aanmaken in ECD mislukt, status: %d", resp.StatusCode)
	}

	// Verwacht een response zoals: { "onderzoekId": "..." }
	var result struct {
		OnderzoekID uuid.UUID `json:"onderzoekId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return uuid.Nil, fmt.Errorf("kan onderzoekId niet uitlezen uit ECD-response: %w", err)
	}
	return result.OnderzoekID, nil
}

func GetOnderzoekByID(ecdURL string, onderzoekID string) (*models.OnderzoekDTO, error) {
	url := fmt.Sprintf("%s/onderzoek/%s", ecdURL, onderzoekID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("onderzoek ophalen uit ECD mislukt, status: %d", resp.StatusCode)
	}

	var onderzoek models.OnderzoekDTO
	if err := json.NewDecoder(resp.Body).Decode(&onderzoek); err != nil {
		return nil, err
	}

	return &onderzoek, nil
}

// Check of diagnose bestaat voor onderzoek dit kan door get onderzoek te doen en te kijken of diagnose niet leeg is
func DiagnoseExistsForOnderzoek(ecdURL string, onderzoekID string) (bool, error) {
	url := fmt.Sprintf("%s/onderzoek/%s", ecdURL, onderzoekID)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("onderzoek ophalen uit ECD mislukt, status: %d", resp.StatusCode)
	}

	var onderzoek models.OnderzoekDTO
	if err := json.NewDecoder(resp.Body).Decode(&onderzoek); err != nil {
		return false, err
	}

	// Controleer of Diagnose niet leeg is
	return len(onderzoek.Diagnose) > 0, nil
}
