package service

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	models "recommendation/model"

	"google.golang.org/genai"
)

type AICategorieenLijstMaker struct {
	GeminiClient *GeminiClient
}

func NewAICategorieenLijstMaker(client *GeminiClient) *AICategorieenLijstMaker {
	return &AICategorieenLijstMaker{GeminiClient: client}
}

func (a *AICategorieenLijstMaker) MaakPassendeCategorieënLijst(ctx context.Context, behoeften string, availableCategories []models.Category) ([]int, error) {
	log.Printf("AICategorieenLijstMaker: Maak passende categorieënlijst voor behoeften '%s' en beschikbare categorieën %v", behoeften, availableCategories)

	availableCategoryIDs, categoriesForAI := ConvertCategoriesToAiReadyStr(availableCategories)

	prompt := fmt.Sprintf("Gegeven de volgende relevante technologie categorieën (ID en Naam): [%s], en de behoeften van de cliënt: '%s', selecteer de meest passende categorie ID's uit deze lijst (alleen nummers, gescheiden door komma's). Negeer categorieën die niet in de gegeven lijst staan. Voorbeelden van categorie ID's zijn 1, 2, 3.", categoriesForAI, behoeften)

	log.Printf("AICategorieenLijstMaker: Roep Gemini API aan voor categorieën met prompt: '%s'", prompt)
	responseText, err := a.GeminiClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("fout van Gemini service bij ophalen categorieën: %w", err)
	}
	log.Printf("AICategorieenLijstMaker: Gemini API antwoord voor categorieën: '%s'", responseText)

	var selectedCategoryIDs []int
	stringIDs := strings.Split(responseText, ",")
	for _, sID := range stringIDs {
		id, err := strconv.Atoi(strings.TrimSpace(sID))
		if err == nil {
			isValid := slices.Contains(availableCategoryIDs, id)
			if isValid {
				selectedCategoryIDs = append(selectedCategoryIDs, id)
			} else {
				log.Printf("Waarschuwing: AI selecteerde categorie ID %d dat niet beschikbaar was. ID wordt genegeerd.", id)
			}
		} else {
			log.Printf("Waarschuwing: Kon AI-antwoord deel '%s' niet naar een getal converteren: %v", sID, err)
		}
	}

	if len(selectedCategoryIDs) == 0 {
		log.Println("AICategorieenLijstMaker: AI gaf geen geldige categorieën terug.")
		return []int{}, nil
	}

	return selectedCategoryIDs, nil
}

func ConvertCategoriesToAiReadyStr(availableCategories []models.Category) ([]int, string) {
	var categoryDetails []string
	var availableCategoryIDs []int
	for _, cat := range availableCategories {
		categoryDetails = append(categoryDetails, fmt.Sprintf("ID: %d, Naam: %s", cat.ID, cat.Naam))
		availableCategoryIDs = append(availableCategoryIDs, cat.ID)
	}
	categoriesForAI := strings.Join(categoryDetails, "; ")
	return availableCategoryIDs, categoriesForAI
}

type AIOplossingenLijstMaker struct {
	geminiClient *GeminiClient
}

func NewAIOplossingenLijstMaker(client *GeminiClient) *AIOplossingenLijstMaker {
	return &AIOplossingenLijstMaker{geminiClient: client}
}

func (a *AIOplossingenLijstMaker) MaakRelevanteTags(ctx context.Context, behoeften string, allAvailableTags []string) ([]string, error) {
	log.Printf("AIOplossingenLijstMaker: Start voor behoeften '%s'", behoeften)

	if len(allAvailableTags) == 0 {
		log.Println("Geen beschikbare tags, kan niets aanbevelen.")
		return []string{}, nil
	}

	prompt := fmt.Sprintf("Gegeven de volgende beschikbare tags: [%s], en de behoeften van de cliënt: '%s', selecteer de meest relevante tags uit deze lijst. Geef alleen de geselecteerde tags terug, exact zoals ze gegeven zijn, gescheiden door komma's (bijv: smart home, mobiliteit).", strings.Join(allAvailableTags, ", "), behoeften)

	log.Printf("AIOplossingenLijstMaker: Roep Gemini API aan voor tags met prompt: '%s'", prompt)
	responseText, err := a.geminiClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("fout van Gemini service bij ophalen tags: %w", err)
	}
	log.Printf("AIOplossingenLijstMaker: Gemini API antwoord voor tags: '%s'", responseText)

	var selectedTags []string
	stringTags := strings.Split(responseText, ",")
	for _, sTag := range stringTags {
		trimmedTag := strings.TrimSpace(sTag)
		if trimmedTag != "" {
			isValid := false
			for _, availableTag := range allAvailableTags {
				if trimmedTag == availableTag {
					isValid = true
					break
				}
			}
			if isValid {
				selectedTags = append(selectedTags, trimmedTag)
			} else {
				log.Printf("Waarschuwing: AI selecteerde tag '%s' die niet beschikbaar was. Tag wordt genegeerd.", trimmedTag)
			}
		}
	}

	if len(selectedTags) == 0 {
		log.Println("AIOplossingenLijstMaker: AI gaf geen geldige tags terug.")
		return []string{}, nil
	}

	return selectedTags, nil
}

type GeminiClient struct {
	client *genai.Client
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &GeminiClient{client: client}, err
}

func (s *GeminiClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	resp, err := s.client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash-lite",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("fout bij het genereren van content: %w", err)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
		var responseText strings.Builder
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				responseText.WriteString(part.Text)
			}
		}
		return responseText.String(), nil
	}

	return "", fmt.Errorf("geen valide content gevonden in Gemini API response")
}
