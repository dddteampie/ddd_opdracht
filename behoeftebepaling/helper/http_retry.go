package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/avast/retry-go"
	"errors"
	"log"
)


// Hulpstruct voor herkenbare HTTP status errors
type httpStatusError struct {
    StatusCode int
}

func (e *httpStatusError) Error() string {
    return fmt.Sprintf("onverwachte statuscode: %d", e.StatusCode)
}

// PostJSONWithRetry verstuurt een JSON payload naar de opgegeven URL met retry logica.
// Het probeert maximaal 3 keer met een wachttijd van 2 seconden tussen pogingen.
// Het retourneert een fout als de statuscode niet overeenkomt met de verwachte status.
func PostJSONWithRetry(url string, payload interface{}, expectedStatus int) error {
    body, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Fout bij JSON encoding: %v", err)
        return fmt.Errorf("fout bij JSON encoding: %w", err)
    }

    attempt := 0
    return retry.Do(
        func() error {
            attempt++
            log.Printf("HTTP POST poging %d naar %s", attempt, url)

            req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
            if err != nil {
                log.Printf("Request creatie faalde bij poging %d: %v", attempt, err)
                return fmt.Errorf("request creatie faalde: %w", err)
            }

            req.Header.Set("Content-Type", "application/json")
            client := &http.Client{Timeout: 5 * time.Second}

            resp, err := client.Do(req)
            if err != nil {
                log.Printf("Netwerkfout bij poging %d: %v", attempt, err)
                return fmt.Errorf("netwerkfout: %w", err)
            }
            defer resp.Body.Close()

            if resp.StatusCode != expectedStatus {
                log.Printf("Onverwachte statuscode bij poging %d: %d", attempt, resp.StatusCode)
                return &httpStatusError{StatusCode: resp.StatusCode}
            }

            log.Printf("Succesvolle POST bij poging %d", attempt)
            return nil
        },
        retry.Attempts(3),
        retry.Delay(2*time.Second),
        retry.DelayType(retry.BackOffDelay),
        retry.LastErrorOnly(true),
        retry.RetryIf(func(err error) bool {
            var httpErr *httpStatusError
            if errors.As(err, &httpErr) {
                return httpErr.StatusCode >= 500 && httpErr.StatusCode < 600
            }
            return true
        }),
    )
}

// func GetJSONWithRetry(url string, target interface{}, expectedStatus int) error {
// 	attempt := 0
// 	return retry.Do(
// 		func() error {
// 			attempt++
// 			log.Printf("HTTP GET poging %d naar %s", attempt, url)

// 			client := &http.Client{Timeout: 5 * time.Second}
// 			resp, err := client.Get(url)
// 			if err != nil {
// 				log.Printf("Netwerkfout bij poging %d: %v", attempt, err)
// 				return fmt.Errorf("netwerkfout: %w", err)
// 			}
// 			defer resp.Body.Close()

// 			if resp.StatusCode != expectedStatus {
// 				log.Printf("Onverwachte statuscode bij poging %d: %d", attempt, resp.StatusCode)
// 				return &httpStatusError{StatusCode: resp.StatusCode}
// 			}

// 			if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
// 				log.Printf("Fout bij JSON decoding bij poging %d: %v", attempt, err)
// 				return fmt.Errorf("fout bij JSON decoding: %w", err)
// 			}

// 			log.Printf("Succesvolle GET bij poging %d", attempt)
// 			return nil
// 		},
// 		retry.Attempts(3),
// 		retry.Delay(2*time.Second),
// 		retry.DelayType(retry.BackOffDelay),
// 		retry.LastErrorOnly(true),
// 	)
// }