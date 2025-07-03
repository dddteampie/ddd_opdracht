package main

import (
	"behoeftebepaling/handlers"
	"behoeftebepaling/pkg/config"
	behoefte_repo "behoeftebepaling/repository"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Load config from .env
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded")

	ecdURL := os.Getenv("ECD_URL") // of uit je eigen config package
	handlers.SetECDURL(ecdURL)

	// Initialize database
	db, err := behoefte_repo.InitDB(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	handlers.InitHandlers(db) // als je dependency injection gebruikt

	r := mux.NewRouter()
	r.HandleFunc("/api/health", handlers.HealthCheckHandler).Methods("GET")
	r.HandleFunc("/behoefte", handlers.CreateBehoefte).Methods("POST")
	r.HandleFunc("/behoefte/onderzoek/{onderzoekId}", handlers.GetBehoefteByOnderzoekID).Methods("GET")
	r.HandleFunc("/behoefte/client", handlers.GetBehoefteByClientNameAndBirthdate).Methods("POST")
	r.HandleFunc("/behoefte/client/{clientId}", handlers.GetBehoefteByClientID).Methods("GET")
	r.HandleFunc("/behoefte/{behoefteId}/aanvraagverwerking", handlers.StuurBehoefteNaarAanvraagverwerking).Methods("POST")

	r.HandleFunc("/ecd/onderzoek/{onderzoekId}/anamnese", handlers.KoppelAnamneseHandler).Methods("POST")
	r.HandleFunc("/ecd/onderzoek/{onderzoekId}/meetresultaat", handlers.KoppelMeetresultaatHandler).Methods("POST")
	r.HandleFunc("/ecd/onderzoek/{onderzoekId}/diagnose", handlers.KoppelDiagnoseHandler).Methods("POST")

	r.HandleFunc("/ecd/client", handlers.KoppelClientHandler).Methods("POST")
	r.HandleFunc("/ecd/client/{clientId}", handlers.GetClientHandler).Methods("GET")
	r.HandleFunc("/ecd/zorgdossier", handlers.KoppelZorgdossierHandler).Methods("POST")
	r.HandleFunc("/ecd/zorgdossier/client/{clientId}", handlers.GetZorgdossierByClientIDHandler).Methods("GET")
	r.HandleFunc("/ecd/onderzoek", handlers.KoppelOnderzoekHandler).Methods("POST")
	r.HandleFunc("/ecd/onderzoek/{onderzoekId}", handlers.GetOnderzoekByIdHandler).Methods("GET")

	log.Printf("Behoeftebepaling-service draait op %s...", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(cfg.ServerPort, r))
}
