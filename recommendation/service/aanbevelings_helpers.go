package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"recommendation/model"
	"strconv"
	"strings"
	"time"

	"io"

	"gorm.io/gorm"
)

type AanbevelingHelpers struct {
	repo                    IAanbevelingsOpslag
	categorieenAILijstMaker ICategorieenAILijstMaker
	oplossingenAILijstMaker IOplossingenAILijstMaker
	productServiceURL       string
}

func NewAanbevelingHelpers(repo IAanbevelingsOpslag, categorieenAILijstMaker ICategorieenAILijstMaker, oplossingenAILijstMaker IOplossingenAILijstMaker, productURL string) IAanbevelingHelpers {

	return &AanbevelingHelpers{
		repo:                    repo,
		categorieenAILijstMaker: categorieenAILijstMaker,
		oplossingenAILijstMaker: oplossingenAILijstMaker,
		productServiceURL:       productURL,
	}
}

func (s *AanbevelingHelpers) HaalAlleTagsOp(ctx context.Context, categoryID *int) ([]string, error) {
	log.Println("AanbevelingHelpers: Ophalen alle tags van product service")

	urlStr := fmt.Sprintf("%s/categorieen/tags", s.productServiceURL)
	if categoryID != nil {
		urlStr = fmt.Sprintf("%s?categorieID=%d", urlStr, *categoryID)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("fout bij aanmaken HTTP-verzoek voor tags: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fout bij verzenden HTTP-verzoek naar product service voor tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("product service retourneerde foutstatus %d voor tags, kon body niet lezen: %w", resp.StatusCode, readErr)
		}
		var errorBody map[string]interface{}
		if jsonErr := json.Unmarshal(bodyBytes, &errorBody); jsonErr == nil {
			return nil, fmt.Errorf("product service retourneerde foutstatus %d voor tags: %v", resp.StatusCode, errorBody)
		}
		return nil, fmt.Errorf("product service retourneerde foutstatus %d voor tags, response body: %s", resp.StatusCode, string(bodyBytes))
	}

	var tagsResponse []model.Tag
	if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil {
		return nil, fmt.Errorf("fout bij decoderen tags response van product service: %w", err)
	}

	var tagNames []string
	for _, tag := range tagsResponse {
		tagNames = append(tagNames, tag.Naam)
	}

	log.Printf("Product Service retourneerde tags: %v", tagNames)
	return tagNames, nil
}

func (s *AanbevelingHelpers) HaalCategorieënOp(ctx context.Context, budget float64) ([]model.Category, error) {
	log.Printf("AanbevelingHelpers: Ophalen categorieën van product service met budget: %.2f", budget)

	url := fmt.Sprintf("%s/categorieen?budget=%.2f", s.productServiceURL, budget)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fout bij aanmaken HTTP-verzoek voor categorieën: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fout bij verzenden HTTP-verzoek naar product service voor categorieën: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err != nil {
			return nil, fmt.Errorf("product service retourneerde foutstatus %d en kon foutbody niet decoderen: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("product service retourneerde foutstatus %d voor categorieën: %v", resp.StatusCode, errorBody)
	}

	var categories []model.Category
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return nil, fmt.Errorf("fout bij decoderen categorieën response van product service: %w", err)
	}

	log.Printf("Product Service retourneerde categorieën: %v", categories)
	return categories, nil
}

func (s *AanbevelingHelpers) HaalCategorieenOpMetIDs(ctx context.Context, ids []int) ([]model.Category, error) {
	log.Printf("AanbevelingHelpers: Ophalen categorieën van product service met ID's: %v", ids)

	if len(ids) == 0 {
		return []model.Category{}, nil
	}

	var idStrings []string
	for _, id := range ids {
		idStrings = append(idStrings, strconv.Itoa(id))
	}
	idsParam := url.QueryEscape(strings.Join(idStrings, ","))

	url := fmt.Sprintf("%s/categorieen?ids=%s", s.productServiceURL, idsParam)
	log.Print(url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fout bij aanmaken HTTP-verzoek voor categorieën by ID: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fout bij verzenden HTTP-verzoek naar product service voor categorieën by ID: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err != nil {
			return nil, fmt.Errorf("product service retourneerde foutstatus %d en kon foutbody niet decoderen voor categorieën by ID: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("product service retourneerde foutstatus %d voor categorieën by ID: %v", resp.StatusCode, errorBody)
	}

	var categories []model.Category
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		return nil, fmt.Errorf("fout bij decoderen categorieën response van product service by ID: %w", err)
	}

	log.Printf("Product Service retourneerde categorieën by ID: %v", categories)
	return categories, nil
}

func (s *AanbevelingHelpers) HaalProductenOp(ctx context.Context, tags []string, budget float64, categorieën []int) ([]model.Product, error) {
	log.Printf("AanbevelingHelpers: Ophalen producten van product service met tags: %v, budget: %.2f, categorieën: %v", tags, budget, categorieën)

	tagString := url.QueryEscape(strings.Join(tags, ","))

	var categorieStrings []string
	for _, cat := range categorieën {
		categorieStrings = append(categorieStrings, strconv.Itoa(cat))
	}
	categorieString := url.QueryEscape(strings.Join(categorieStrings, ","))

	url := fmt.Sprintf("%s/product?tags=%s&budget=%.2f&categorieen=%s", s.productServiceURL, tagString, budget, categorieString)
	log.Print(url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fout bij aanmaken HTTP-verzoek voor producten: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fout bij verzenden HTTP-verzoek naar product service voor producten: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err != nil {
			return nil, fmt.Errorf("product service retourneerde foutstatus %d en kon foutbody niet decoderen: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("product service retourneerde foutstatus %d voor producten: %v", resp.StatusCode, errorBody)
	}

	var products []model.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, fmt.Errorf("fout bij decoderen producten response van product service: %w", err)
	}

	log.Printf("Product Service retourneerde Producten: %v", products)
	return products, nil
}

func (s *AanbevelingHelpers) HaalProductenOpMetEANs(ctx context.Context, eans []int) ([]model.Product, error) {
	log.Printf("AanbevelingHelpers: Ophalen producten van product service met EAN's: %v", eans)

	if len(eans) == 0 {
		return []model.Product{}, nil
	}

	var eanStrings []string
	for _, ean := range eans {
		eanStrings = append(eanStrings, strconv.Itoa(ean))
	}
	eansParam := url.QueryEscape(strings.Join(eanStrings, ","))

	url := fmt.Sprintf("%s/product?eans=%s", s.productServiceURL, eansParam)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fout bij aanmaken HTTP-verzoek voor producten by EAN: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fout bij verzenden HTTP-verzoek naar product service voor producten by EAN: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("product service retourneerde foutstatus %d voor producten by EAN, kon body niet lezen: %w", resp.StatusCode, readErr)
		}
		var errorBody map[string]interface{}
		if jsonErr := json.Unmarshal(bodyBytes, &errorBody); jsonErr == nil {
			return nil, fmt.Errorf("product service retourneerde foutstatus %d voor producten by EAN: %v", resp.StatusCode, errorBody)
		}
		return nil, fmt.Errorf("product service retourneerde foutstatus %d voor producten by EAN, response body: %s", resp.StatusCode, string(bodyBytes))
	}

	var products []model.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, fmt.Errorf("fout bij decoderen producten response van product service by EAN: %w", err)
	}

	log.Printf("Product Service retourneerde Producten by EAN: %v", products)
	return products, nil
}

func (s *AanbevelingHelpers) MaakPassendeCategorieënLijst(ctx context.Context, patientID string, budget float64, behoeften string) (*model.PassendeCategorieënLijst, error) {
	log.Printf("Service: Start MaakPassendeCategorieënLijst voor patientID: %s met budget: %.2f", patientID, budget)

	budgetFittingCategoriesRaw, err := s.HaalCategorieënOp(ctx, budget)
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen budget-passende categorieën van product service: %w", err)
	}
	log.Printf("Product Service retourneerde budget-passende categorieën: %v", budgetFittingCategoriesRaw)

	if len(budgetFittingCategoriesRaw) == 0 {
		log.Printf("Geen budget-passende categorieën gevonden voor cliënt %s. Retourneer lege lijst.", patientID)
		return &model.PassendeCategorieënLijst{Categories: []model.Category{}}, nil
	}

	selectedCategoryIDs, err := s.categorieenAILijstMaker.MaakPassendeCategorieënLijst(ctx, behoeften, budgetFittingCategoriesRaw)
	if err != nil {
		return nil, fmt.Errorf("fout bij het maken van passende categorieënlijst door AI: %w", err)
	}
	log.Printf("AI genereerde geselecteerde categorie-ID's: %v", selectedCategoryIDs)

	existingRec, err := s.repo.HaalAanbevelingOpMetCliëntID(ctx, patientID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("fout bij ophalen bestaande aanbeveling: %w", err)
	}

	var passendeLijst *model.PassendeCategorieënLijst

	if existingRec != nil && existingRec.PassendeCategorieënID != nil {
		passendeLijst, err = s.repo.HaalPassendeCategorieënLijstOpMetID(ctx, *existingRec.PassendeCategorieënID)
		if err != nil {
			return nil, fmt.Errorf("fout bij ophalen bestaande passende categorieënlijst met ID %d: %w", *existingRec.PassendeCategorieënID, err)
		}
		if passendeLijst == nil {
			log.Printf("Waarschuwing: Passende categorieënlijst met ID %d niet gevonden, maar wel gerefereerd. Nieuwe lijst wordt aangemaakt.", *existingRec.PassendeCategorieënID)
			passendeLijst = &model.PassendeCategorieënLijst{
				CategoryIDs: model.ConvertIntSliceToPQInt64Array(selectedCategoryIDs),
			}
			if err := s.repo.MaakPassendeCategorieënLijstDB(ctx, passendeLijst); err != nil {
				return nil, fmt.Errorf("fout bij opslaan nieuwe passende categorieënlijst na niet gevonden bestaande: %w", err)
			}
			log.Printf("Nieuwe passende categorieënlijst aangemaakt met ID: %d", passendeLijst.ID)
			existingRec.PassendeCategorieënID = &passendeLijst.ID
		} else {
			passendeLijst.CategoryIDs = model.ConvertIntSliceToPQInt64Array(selectedCategoryIDs)
			if err := s.repo.WerkPassendeCategorieënLijstBijDB(ctx, passendeLijst); err != nil {
				return nil, fmt.Errorf("fout bij bijwerken passende categorieënlijst met ID %d: %w", passendeLijst.ID, err)
			}
			log.Printf("Bestaande passende categorieënlijst met ID %d bijgewerkt.", passendeLijst.ID)
		}
	} else {
		passendeLijst = &model.PassendeCategorieënLijst{
			CategoryIDs: model.ConvertIntSliceToPQInt64Array(selectedCategoryIDs),
		}
		if err := s.repo.MaakPassendeCategorieënLijstDB(ctx, passendeLijst); err != nil {
			return nil, fmt.Errorf("fout bij opslaan nieuwe passende categorieënlijst: %w", err)
		}
		log.Printf("Nieuwe passende categorieënlijst opgeslagen met ID: %d", passendeLijst.ID)
		if existingRec != nil {
			existingRec.PassendeCategorieënID = &passendeLijst.ID
		}
	}

	if existingRec != nil {
		existingRec.Versie++
		existingRec.AanmaakDatum = time.Now()
		if err := s.repo.WerkAanbevelingBij(ctx, existingRec); err != nil {
			return nil, fmt.Errorf("fout bij bijwerken aanbeveling voor cliënt %s: %w", patientID, err)
		}
		log.Printf("Bestaande aanbeveling voor cliënt %s bijgewerkt naar versie %d (categorieën) met PassendeCategorieënID: %d", patientID, existingRec.Versie, *existingRec.PassendeCategorieënID)
	} else {
		rec := &model.Aanbeveling{
			ClientID:              patientID,
			Versie:                1,
			AanmaakDatum:          time.Now(),
			PassendeCategorieënID: &passendeLijst.ID,
			OplossingenLijstID:    nil,
		}
		if err := s.repo.SlaAanbevelingOp(ctx, rec); err != nil {
			return nil, fmt.Errorf("fout bij creëren nieuwe aanbeveling voor cliënt %s: %w", patientID, err)
		}
		log.Printf("Nieuwe aanbeveling gecreëerd voor cliënt %s met ID %d en versie %d (categorieën) met PassendeCategorieënID: %d", patientID, rec.ID, rec.Versie, *rec.PassendeCategorieënID)
	}

	fullCategories, err := s.HaalCategorieenOpMetIDs(ctx, selectedCategoryIDs)
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen volledige categorie-details voor response: %w", err)
	}
	passendeLijst.Categories = fullCategories

	return passendeLijst, nil
}

func (s *AanbevelingHelpers) MaakOplossingenLijst(ctx context.Context, clientID string, budget float64, behoeften string, categoryID *int) (*model.OplossingenLijst, error) {
	log.Printf("Service: Start MaakOplossingenLijst voor clientID: %s met budget: %.2f", clientID, budget)

	allAvailableTags, err := s.HaalAlleTagsOp(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen alle tags van product service: %w", err)
	}
	log.Printf("Product Service retourneerde alle tags: %v", allAvailableTags)

	selectedTags, err := s.oplossingenAILijstMaker.MaakRelevanteTags(ctx, behoeften, allAvailableTags)
	if err != nil {
		return nil, fmt.Errorf("fout bij het maken van relevante tags door AI: %w", err)
	}
	log.Printf("AI genereerde geselecteerde tags: %v", selectedTags)

	var relevanteCategorieIDs []int

	if categoryID != nil {
		relevanteCategorieIDs = append(relevanteCategorieIDs, *categoryID)
		log.Printf("Gebruikt expliciet meegegeven categorie ID: %d", *categoryID)
	} else {
		existingRec, err := s.repo.HaalAanbevelingOpMetCliëntID(ctx, clientID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("fout bij ophalen bestaande aanbeveling voor oplossingenlijst: %w", err)
		}

		if existingRec == nil || existingRec.PassendeCategorieënID == nil {
			return nil, fmt.Errorf("geen bestaande passende categorieënlijst gevonden voor cliënt %s. Genereer deze eerst of geef een categoryId mee.", clientID)
		}

		passendeLijstFromDB, err := s.repo.HaalPassendeCategorieënLijstOpMetID(ctx, *existingRec.PassendeCategorieënID)
		if err != nil {
			return nil, fmt.Errorf("fout bij ophalen passende categorieën lijst voor cliënt %s: %w", clientID, err)
		}
		if passendeLijstFromDB == nil {
			return nil, fmt.Errorf("passende categorieënlijst met ID %d niet gevonden voor cliënt %s", *existingRec.PassendeCategorieënID, clientID)
		}

		relevanteCategorieIDs = model.ConvertPQInt64ArrayToIntSlice(passendeLijstFromDB.CategoryIDs)
		log.Printf("Bestaande categorie-ID's gevonden voor cliënt %s: %v", clientID, relevanteCategorieIDs)
	}

	if len(relevanteCategorieIDs) == 0 {
		return nil, fmt.Errorf("geen relevante categorieën beschikbaar om producten voor op te halen")
	}

	finalProducts, err := s.HaalProductenOp(ctx, selectedTags, budget, relevanteCategorieIDs)
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen producten van product service met geselecteerde tags en categorieën: %w", err)
	}
	log.Printf("Product Service retourneerde finale Producten: %v", finalProducts)

	var oplossingenLijst *model.OplossingenLijst
	var productEANs []int64
	for _, p := range finalProducts {
		productEANs = append(productEANs, p.EAN)
	}

	existingRec, err := s.repo.HaalAanbevelingOpMetCliëntID(ctx, clientID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("fout bij ophalen bestaande aanbeveling voor oplossingenlijst update/create: %w", err)
	}

	if existingRec != nil && existingRec.OplossingenLijstID != nil {
		oplossingenLijst, err = s.repo.HaalOplossingenLijstOpMetID(ctx, *existingRec.OplossingenLijstID)
		if err != nil {
			return nil, fmt.Errorf("fout bij ophalen bestaande oplossingenlijst met ID %d: %w", *existingRec.OplossingenLijstID, err)
		}
		if oplossingenLijst == nil {
			log.Printf("Waarschuwing: Oplossingenlijst met ID %d niet gevonden, maar wel gerefereerd. Nieuwe lijst wordt aangemaakt.", *existingRec.OplossingenLijstID)
			oplossingenLijst = &model.OplossingenLijst{
				ProductEANs: model.ConvertInt64SliceToPQInt64Array(productEANs),
			}
			if err := s.repo.MaakOplossingenLijstDB(ctx, oplossingenLijst); err != nil {
				return nil, fmt.Errorf("fout bij opslaan nieuwe oplossingenlijst na niet gevonden bestaande: %w", err)
			}
			log.Printf("Nieuwe oplossingenlijst aangemaakt met ID: %d", oplossingenLijst.ID)
			existingRec.OplossingenLijstID = &oplossingenLijst.ID
		} else {
			oplossingenLijst.ProductEANs = model.ConvertInt64SliceToPQInt64Array(productEANs)
			if err := s.repo.WerkOplossingenLijstBijDB(ctx, oplossingenLijst); err != nil {
				return nil, fmt.Errorf("fout bij bijwerken oplossingenlijst met ID %d: %w", oplossingenLijst.ID, err)
			}
			log.Printf("Bestaande oplossingenlijst met ID %d bijgewerkt.", oplossingenLijst.ID)
		}
	} else {
		oplossingenLijst = &model.OplossingenLijst{
			ProductEANs: model.ConvertInt64SliceToPQInt64Array(productEANs),
		}
		if err := s.repo.MaakOplossingenLijstDB(ctx, oplossingenLijst); err != nil {
			return nil, fmt.Errorf("fout bij opslaan nieuwe oplossingenlijst: %w", err)
		}
		log.Printf("Nieuwe oplossingenlijst opgeslagen met ID: %d", oplossingenLijst.ID)
		if existingRec != nil {
			existingRec.OplossingenLijstID = &oplossingenLijst.ID
		}
	}

	if existingRec != nil {
		existingRec.Versie++
		existingRec.AanmaakDatum = time.Now()
		if err := s.repo.WerkAanbevelingBij(ctx, existingRec); err != nil {
			return nil, fmt.Errorf("fout bij bijwerken aanbeveling voor cliënt %s: %w", clientID, err)
		}
		log.Printf("Bestaande aanbeveling voor cliënt %s bijgewerkt naar versie %d (oplossingen) met OplossingenLijstID: %d", clientID, existingRec.Versie, *existingRec.OplossingenLijstID)
	} else {
		rec := &model.Aanbeveling{
			ClientID:              clientID,
			Versie:                1,
			AanmaakDatum:          time.Now(),
			PassendeCategorieënID: nil,
			OplossingenLijstID:    &oplossingenLijst.ID,
		}
		if err := s.repo.SlaAanbevelingOp(ctx, rec); err != nil {
			return nil, fmt.Errorf("fout bij creëren nieuwe aanbeveling voor cliënt %s: %w", clientID, err)
		}
		log.Printf("Nieuwe aanbeveling gecreëerd voor cliënt %s met ID %d en versie %d (oplossingen) met OplossingenLijstID: %d", clientID, rec.ID, rec.Versie, *rec.OplossingenLijstID)
	}

	oplossingenLijst.Products = finalProducts

	return oplossingenLijst, nil
}

func (s *AanbevelingHelpers) HaalPassendeCategorieënLijstOp(ctx context.Context, patientID string) (*model.PassendeCategorieënLijst, error) {
	log.Printf("Service: HaalPassendeCategorieënLijstOp voor patientID: %s", patientID)
	rec, err := s.repo.HaalAanbevelingOpMetCliëntID(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen aanbeveling voor cliënt %s: %w", patientID, err)
	}
	if rec == nil || rec.PassendeCategorieënID == nil {
		return nil, fmt.Errorf("geen passende categorieënlijst gevonden voor cliënt %s", patientID)
	}

	passendeLijstFromDB, err := s.repo.HaalPassendeCategorieënLijstOpMetID(ctx, *rec.PassendeCategorieënID)
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen passende categorieënlijst uit database: %w", err)
	}
	if passendeLijstFromDB == nil {
		return nil, fmt.Errorf("passende categorieënlijst met ID %d niet gevonden", *rec.PassendeCategorieënID)
	}

	var intCategoryIDs []int
	for _, id := range passendeLijstFromDB.CategoryIDs {
		intCategoryIDs = append(intCategoryIDs, int(id))
	}
	fullCategories, err := s.HaalCategorieenOpMetIDs(ctx, intCategoryIDs)
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen volledige categorie-details van product service: %w", err)
	}

	passendeLijstFromDB.Categories = fullCategories
	return passendeLijstFromDB, nil
}

func (s *AanbevelingHelpers) HaalOplossingenLijstOp(ctx context.Context, clientID string) (*model.OplossingenLijst, error) {
	log.Printf("Service: HaalOplossingenLijstOp voor clientID: %s", clientID)
	rec, err := s.repo.HaalAanbevelingOpMetCliëntID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen aanbeveling voor cliënt %s: %w", clientID, err)
	}
	if rec == nil || rec.OplossingenLijstID == nil {
		return nil, fmt.Errorf("geen oplossingenlijst gevonden voor cliënt %s", clientID)
	}

	oplossingenLijstFromDB, err := s.repo.HaalOplossingenLijstOpMetID(ctx, *rec.OplossingenLijstID)
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen oplossingenlijst uit database: %w", err)
	}
	if oplossingenLijstFromDB == nil {
		return nil, fmt.Errorf("oplossingenlijst met ID %d niet gevonden", *rec.OplossingenLijstID)
	}

	fullProducts, err := s.HaalProductenOpMetEANs(ctx, model.ConvertPQInt64ArrayToIntSlice(oplossingenLijstFromDB.ProductEANs))
	if err != nil {
		return nil, fmt.Errorf("fout bij ophalen volledige product-details van product service: %w", err)
	}

	oplossingenLijstFromDB.Products = fullProducts
	return oplossingenLijstFromDB, nil
}
