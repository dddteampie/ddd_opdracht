package helper

import (
	"behoeftebepaling/models"
	"fmt"
)

func NotifyAanvraagverwerking(behoefte models.Behoefte, client models.Client) error {
	aanvraagURL := "http://host.docker.internal:8085/aanvraagverwerking/aanvraag"
	payload := map[string]interface{}{
		"client":   client,
		"behoefte": behoefte,
	}
	// Gebruik de helper met retry
	if err := PostJSONWithRetry(aanvraagURL, payload, 201); err != nil {
		return fmt.Errorf("aanvraagverwerking gaf een fout terug: %w", err)
	}
	return nil
}
