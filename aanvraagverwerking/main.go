package main

import (
	"aanvraagverwerking/handlers"
	"aanvraagverwerking/pkg/auth"
	"aanvraagverwerking/pkg/config"
	aanvraagverwerking_repo "aanvraagverwerking/repository"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Load config from .env
	cfg, err := config.LoadConfig("aanvraagverwerking.env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded")

	handlers.SetRECURL(cfg.RecommendationUrl)

	// Initialize database
	db, err := aanvraagverwerking_repo.InitDB(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	handlers.InitHandlers(db) // als je dependency injection gebruikt

	authConfig := auth.AuthZMiddlewareConfig{
		RolesClaimName: "realm_access",
		DevMode:        cfg.AuthzDevMode,
	}

	// Router setup
	r := mux.NewRouter()

	r.HandleFunc("/aanvraagverwerking/api/health", handlers.HealthCheckHandler).Methods("GET")
	r.Handle("/aanvraagverwerking/aanvraag", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.StartAanvraag))).Methods("POST")
	r.Handle("/aanvraagverwerking/aanvraag/{id}", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.GetAanvraagByID))).Methods("GET")
	r.Handle("/aanvraagverwerking/aanvraag/client/{clientId}", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.GetAanvragenByClientID))).Methods("GET")

	r.Handle("/aanvraagverwerking/aanvraag/categorie", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.StartCategorieAanvraag))).Methods("PUT")
	r.Handle("/aanvraagverwerking/aanvraag/categorie/kies", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.KiesCategorie))).Methods("POST")
	r.Handle("/aanvraagverwerking/aanvraag/product", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.StartProductAanvraag))).Methods("PUT")
	r.Handle("/aanvraagverwerking/aanvraag/product/kies", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.KiesProduct))).Methods("POST")

	r.Handle("/aanvraagverwerking/aanvraag/recommendatie/categorie/", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.HaalPassendeCategorieenLijstOp))).Methods("GET")
	r.Handle("/aanvraagverwerking/aanvraag/recommendatie/product/", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.HaalPassendeProductenLijstOp))).Methods("GET")

	log.Printf("Behoeftebepaling-service draait op %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(cfg.ServerPort, r))
}
