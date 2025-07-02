package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"recommendation/handlers"
	"recommendation/pkg/auth"
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

	authConfig := auth.AuthZMiddlewareConfig{
		RolesClaimName: "realm_access",
		DevMode:        config.AuthzDevMode,
	}

	r := mux.NewRouter()

	r.Handle("/recommend/categorie/", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(aanbevelingsHandler.MaakPassendeCategorieënLijstHandler))).Methods("PUT")
	r.Handle("/recommend/oplossing/", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(aanbevelingsHandler.MaakOplossingenLijstHandler))).Methods("PUT")
	r.Handle("/recommend/categorie/", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(aanbevelingsHandler.HaalPassendeCategorieënLijstOpHandler))).Methods("GET")
	r.Handle("/recommend/oplossing/", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(aanbevelingsHandler.HaalOplossingenLijstOpHandler))).Methods("GET")
	r.HandleFunc("/health", aanbevelingsHandler.HealthCheckHandler).Methods("GET")

	log.Printf("Aanbevelingsservice draait op poort %s", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, r))
}
