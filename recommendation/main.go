package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"recommendation/handlers"
	"recommendation/pkg/config"
	"recommendation/repository"
	"recommendation/service"
)

func main() {
	//load correct config
	config, err := config.LoadConfig("recommendation.env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	//initialize config
	db, err := repository.InitDB(config.DatabaseDSN)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	aanbevelingsOpslag := repository.NewAanbevelingsOpslag(db)
	ctx := context.Background()
	geminiClient, err := service.NewGeminiClient(ctx, config.GeminiKey)
	categorieenAILijstMaker := service.NewAICategorieenLijstMaker(geminiClient)
	oplossingenAILijstMaker := service.NewAIOplossingenLijstMaker(geminiClient)

	aanbevelingsSvc := service.NewAanbevelingHelpers(
		aanbevelingsOpslag,
		categorieenAILijstMaker,
		oplossingenAILijstMaker,
		config.ProductServiceURL,
	)

	aanbevelingsHandler := handlers.NewAanbevelingsHandler(aanbevelingsSvc)

	r := mux.NewRouter()

	r.HandleFunc("/recommendation/recommend/categorie/", aanbevelingsHandler.MaakPassendeCategorieënLijstHandler).Methods("PUT")
	r.HandleFunc("/recommendation/recommend/oplossing/", aanbevelingsHandler.MaakOplossingenLijstHandler).Methods("PUT")
	r.HandleFunc("/recommendation/recommend/categorie/", aanbevelingsHandler.HaalPassendeCategorieënLijstOpHandler).Methods("GET")
	r.HandleFunc("/recommendation/recommend/oplossing/", aanbevelingsHandler.HaalOplossingenLijstOpHandler).Methods("GET")
	r.HandleFunc("/recommendation/api/health", aanbevelingsHandler.HealthCheckHandler).Methods("GET")

	log.Printf("Aanbevelingsservice draait op poort %s", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, r))
}
