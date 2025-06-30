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

// func GetOnderzoekenVanCliënt(ecdURL, clientId string) ([]models.Onderzoek, error) {
// 	url := fmt.Sprintf("%s/client/%s/onderzoeken", ecdURL, clientId)
// 	var onderzoeken []models.Onderzoek
// 	err := helper.GetJSONWithRetry(url, &onderzoeken, http.StatusOK)
// 	if err != nil {
// 		return nil, fmt.Errorf("fout bij ophalen onderzoeken: %w", err)
// 	}
// 	return onderzoeken, nil
// }