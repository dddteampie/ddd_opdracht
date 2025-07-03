package main

import (
	"aanvraagverwerking/handlers"
	"aanvraagverwerking/pkg/config"
	aanvraagverwerking_repo "aanvraagverwerking/repository"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Load config from .env
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded")

	// Initialize database
	db, err := aanvraagverwerking_repo.InitDB(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	handlers.InitHandlers(db) // als je dependency injection gebruikt

	r := mux.NewRouter()
	r.HandleFunc("/aanvraagverwerking/api/health", handlers.HealthCheckHandler).Methods("GET")
	r.HandleFunc("/aanvraagverwerking/aanvraag", handlers.StartAanvraag).Methods("POST")
	r.HandleFunc("/aanvraagverwerking/aanvraag/{id}", handlers.GetAanvraagByID).Methods("GET")
	r.HandleFunc("/aanvraagverwerking/aanvraag/client/{clientId}", handlers.GetAanvragenByClientID).Methods("GET")

	r.HandleFunc("/aanvraagverwerking/aanvraag/categorie", handlers.StartCategorieAanvraag).Methods("PUT")
	r.HandleFunc("/aanvraagverwerking/aanvraag/categorie/kies", handlers.KiesCategorie).Methods("POST")
	r.HandleFunc("/aanvraagverwerking/aanvraag/product", handlers.StartProductAanvraag).Methods("PUT")
	r.HandleFunc("/aanvraagverwerking/aanvraag/product/kies", handlers.KiesProduct).Methods("POST")

	r.HandleFunc("/aanvraagverwerking/aanvraag/recommendatie/categorie/", handlers.HaalPassendeCategorieenLijstOp).Methods("GET")
	r.HandleFunc("/aanvraagverwerking/aanvraag/recommendatie/product/", handlers.HaalPassendeProductenLijstOp).Methods("GET")
	log.Printf("Behoeftebepaling-service draait op %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(cfg.ServerPort, r))
}
