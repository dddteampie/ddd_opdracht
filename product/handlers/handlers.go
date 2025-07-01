package handlers

import (
	"encoding/json"
	"net/http"
	models "product/model"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

var DB *gorm.DB

func InitHandlers(db *gorm.DB) {
	DB = db
}

func HaalProductLeveraarsOp(w http.ResponseWriter, r *http.Request) {
	eanStr := r.URL.Query().Get("ean")
	ean, _ := strconv.Atoi(eanStr)

	var aanboden []models.ProductAanbod
	if err := DB.Where("product_ean = ?", ean).Find(&aanboden).Error; err != nil {
		http.Error(w, "Database fout", http.StatusInternalServerError)
		return
	}

	// verzamel unieke leverancier_ids
	leverancierMap := make(map[uint]bool)
	var leverancierIDs []uint
	for _, aanbod := range aanboden {
		if !leverancierMap[aanbod.LeverancierID] {
			leverancierMap[aanbod.LeverancierID] = true
			leverancierIDs = append(leverancierIDs, aanbod.LeverancierID)
		}
	}

	var leveranciers []models.Supplier
	if err := DB.Where("id IN ?", leverancierIDs).Find(&leveranciers).Error; err != nil {
		http.Error(w, "Leveranciers ophalen mislukt", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(leveranciers)
}

func HaalProductenOp(w http.ResponseWriter, r *http.Request) {
	var producten []models.Product

	query := DB.Model(&models.Product{}).
		Joins("JOIN product_aanbods ON product_aanbods.product_ean = products.ean").
		Preload("ProductAanbod.Supplier").
		Preload("Tags").
		Preload("Categorieen").
		Where("products.deleted_at IS NULL")

	eansStr := r.URL.Query().Get("eans")
	if eansStr != "" {
		stringEANs := strings.Split(eansStr, ",")
		var int64EANs []int64
		for _, sEAN := range stringEANs {
			ean, err := strconv.ParseInt(strings.TrimSpace(sEAN), 10, 64)
			if err == nil {
				int64EANs = append(int64EANs, ean)
			}
		}
		if len(int64EANs) > 0 {
			query = query.Where("products.ean IN (?)", int64EANs)
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.Product{})
			return
		}
	}

	budgetStr := r.URL.Query().Get("budget")
	if budgetStr != "" {
		float_budget, err := strconv.ParseFloat(budgetStr, 64)
		if err != nil {
			http.Error(w, "Ongeldige budget-parameter", http.StatusBadRequest)
			return
		}
		budget := int(float_budget)
		query = query.Where("product_aanbods.prijs <= ?", budget)
	}

	tagsStr := r.URL.Query().Get("tags")
	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		if len(tags) > 0 && tags[0] != "" {
			query = query.
				Joins("JOIN product_tags ON product_tags.product_ean = products.ean").
				Joins("JOIN tags ON tags.id = product_tags.tag_id").
				Where("tags.naam IN (?)", tags).
				Group("products.ean").
				Having("COUNT(DISTINCT tags.naam) = ?", len(tags))
		}
	}

	categorieenStr := r.URL.Query().Get("categorieen")
	if categorieenStr != "" {
		categorieIDs := []int{}
		for _, idStr := range strings.Split(categorieenStr, ",") {
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err == nil {
				categorieIDs = append(categorieIDs, id)
			}
		}
		if len(categorieIDs) > 0 {
			query = query.
				Joins("JOIN product_categories ON product_categories.product_ean = products.ean").
				Where("product_categories.categorie_id IN (?)", categorieIDs)
		}
	}

	err := query.Find(&producten).Error
	if err != nil {
		http.Error(w, "Database fout bij ophalen producten: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(producten)
}

func HaalCategorieenOp(w http.ResponseWriter, r *http.Request) {
	var categorieen []models.Categorie
	query := DB.Model(&models.Categorie{})

	budgetStr := r.URL.Query().Get("budget")
	if budgetStr != "" {
		float_budget, err := strconv.ParseFloat(budgetStr, 64)
		if err != nil {
			http.Error(w, "Ongeldige budget-parameter", http.StatusBadRequest)
			return
		}
		budget := int(float_budget)
		query = query.Where("price_range <= ?", budget)
	}

	idsStr := r.URL.Query().Get("ids")
	if idsStr != "" {
		stringIDs := strings.Split(idsStr, ",")
		var intIDs []int
		for _, sID := range stringIDs {
			id, err := strconv.Atoi(strings.TrimSpace(sID))
			if err == nil {
				intIDs = append(intIDs, id)
			}
		}
		if len(intIDs) > 0 {
			query = query.Where("id IN (?)", intIDs)
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.Categorie{})
			return
		}
	}

	if err := query.Find(&categorieen).Error; err != nil {
		http.Error(w, "Database fout bij ophalen categorieën: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categorieen)
}

func PlaatsReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode niet toegestaan", http.StatusMethodNotAllowed)
		return
	}

	var newReview models.Review
	err := json.NewDecoder(r.Body).Decode(&newReview)
	if err != nil {
		http.Error(w, "Ongeldige review data: "+err.Error(), http.StatusBadRequest)
		return
	}

	if newReview.ProductEAN == 0 || newReview.Naam == "" || newReview.Score < 1 || newReview.Score > 5 {
		http.Error(w, "Ontbrekende of ongeldige verplichte reviewvelden (EAN, Naam, Score)", http.StatusBadRequest)
		return
	}

	if err := DB.Create(&newReview).Error; err != nil {
		http.Error(w, "Kon review niet opslaan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newReview)
}

func VoegNieuwProductToe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode niet toegestaan", http.StatusMethodNotAllowed)
		return
	}

	var newProduct models.Product
	err := json.NewDecoder(r.Body).Decode(&newProduct)
	if err != nil {
		http.Error(w, "Ongeldige product data: "+err.Error(), http.StatusBadRequest)
		return
	}

	if newProduct.EAN == 0 || newProduct.Naam == "" {
		http.Error(w, "EAN en Naam zijn verplichte velden", http.StatusBadRequest)
		return
	}

	if err := DB.Create(&newProduct).Error; err != nil {
		http.Error(w, "Kon product niet toevoegen: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProduct)
}

func VoegProductAanbodToe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode niet toegestaan", http.StatusMethodNotAllowed)
		return
	}

	var newOffer models.ProductAanbod
	err := json.NewDecoder(r.Body).Decode(&newOffer)
	if err != nil {
		http.Error(w, "Ongeldige aanbod data: "+err.Error(), http.StatusBadRequest)
		return
	}

	if newOffer.ProductEAN == 0 || newOffer.Prijs <= 0 || newOffer.Voorraad < 0 || newOffer.LeverancierID == 0 {
		http.Error(w, "ProductEAN, Prijs, Voorraad en LeverancierID zijn verplichte velden", http.StatusBadRequest)
		return
	}

	var product models.Product
	if err := DB.Where("ean = ?", newOffer.ProductEAN).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product met opgegeven EAN niet gevonden", http.StatusNotFound)
			return
		}
		http.Error(w, "Database fout bij zoeken product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := DB.Create(&newOffer).Error; err != nil {
		http.Error(w, "Kon aanbod niet toevoegen: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newOffer)
}
func HaalTagsOp(w http.ResponseWriter, r *http.Request) {
	var tags []models.Tag
	query := DB.Model(&models.Tag{})

	// Filter op categorieID indien aanwezig
	categorieIDStr := r.URL.Query().Get("categorieID")
	if categorieIDStr != "" {
		categorieID, err := strconv.Atoi(categorieIDStr)
		if err != nil {
			http.Error(w, "Ongeldige categorieID-parameter", http.StatusBadRequest)
			return
		}

		query = query.
			Joins("JOIN product_tags ON product_tags.tag_id = tags.id").
			Joins("JOIN product_categories ON product_categories.product_ean = product_tags.product_ean").
			Where("product_categories.categorie_id = ?", categorieID).
			Group("tags.id")
	}

	if err := query.Find(&tags).Error; err != nil {
		http.Error(w, "Database fout bij ophalen tags: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}
