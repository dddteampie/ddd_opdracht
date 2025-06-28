package service

import (
	"behoeftebepaling/helper"
	"behoeftebepaling/models"
	"fmt"
	"net/http"
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

// Check of diagnose bestaat voor onderzoek
func DiagnoseExistsForOnderzoek(ecdURL string, onderzoekID string) (bool, error) {
    url := fmt.Sprintf("%s/onderzoek/%s/diagnose", ecdURL, onderzoekID)
    resp, err := http.Get(url)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    return resp.StatusCode == http.StatusOK, nil
}

// func GetOnderzoekenVanCliënt(ecdURL, clientId string) ([]models.Onderzoek, error) {
// 	url := fmt.Sprintf("%s/client/%s/onderzoeken", ecdURL, clientId)
// 	var onderzoeken []models.Onderzoek
// 	err := helper.GetJSONWithRetry(url, &onderzoeken, http.StatusOK)
// 	if err != nil {
// 		return nil, fmt.Errorf("fout bij ophalen onderzoeken: %w", err)
// 	}
// 	return onderzoeken, nil
// }