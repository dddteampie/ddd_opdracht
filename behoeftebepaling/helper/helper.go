package helper

import (
	"behoeftebepaling/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func NotifyAanvraagverwerking(behoefte models.Behoefte, client models.Client) (string, error) {
    aanvraagURL := "http://aanvraagverwerking-service:8080/aanvraag"
    payload := map[string]interface{}{
        "client":   client,
        "behoefte": behoefte,
    }
    body, _ := json.Marshal(payload)
    resp, err := http.Post(aanvraagURL, "application/json", bytes.NewBuffer(body))
    if err != nil {
        return "", fmt.Errorf("aanvraagverwerking gaf een fout terug: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != 201 {
        return "", fmt.Errorf("aanvraagverwerking gaf status %d", resp.StatusCode)
    }
    var result struct {
        ID string `json:"id"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("kan aanvraag_id niet uitlezen uit response: %w", err)
    }
    return result.ID, nil
}