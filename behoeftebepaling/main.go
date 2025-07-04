package main

import (
	"behoeftebepaling/handlers"
	"behoeftebepaling/pkg/auth"
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

    ecdURL := os.Getenv("ECD_URL") 
    handlers.SetECDURL(ecdURL)

    // Initialize database
    db, err := behoefte_repo.InitDB(cfg.DatabaseDSN)
    if err != nil {
        log.Fatalf("Database initialization failed: %v", err)
    }
    handlers.InitHandlers(db) 

    authConfig := auth.AuthZMiddlewareConfig{
		RolesClaimName: "realm_access",
		DevMode:        cfg.AuthzDevMode,
	}

    r := mux.NewRouter()
    r.HandleFunc("/behoeftebepaling/api/health", handlers.HealthCheckHandler).Methods("GET")
    r.Handle("/behoeftebepaling/behoefte", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.CreateBehoefte))).Methods("POST")
    r.Handle("/behoeftebepaling/behoefte/onderzoek/{onderzoekId}", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.GetBehoefteByOnderzoekID))).Methods("GET")
    r.Handle("/behoeftebepaling/behoefte/client", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.GetBehoefteByClientNameAndBirthdate))).Methods("POST")
    r.Handle("/behoeftebepaling/behoefte/client/{clientId}", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.GetBehoefteByClientID))).Methods("GET")
    r.Handle("/behoeftebepaling/behoefte/{behoefteId}/aanvraagverwerking", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.StuurBehoefteNaarAanvraagverwerking))).Methods("POST")

    r.Handle("/behoeftebepaling/ecd/onderzoek/{onderzoekId}/anamnese", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.KoppelAnamneseHandler))).Methods("POST")
    r.Handle("/behoeftebepaling/ecd/onderzoek/{onderzoekId}/meetresultaat", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.KoppelMeetresultaatHandler))).Methods("POST")
    r.Handle("/behoeftebepaling/ecd/onderzoek/{onderzoekId}/diagnose", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.KoppelDiagnoseHandler))).Methods("POST")

    r.Handle("/behoeftebepaling/ecd/client", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.KoppelClientHandler))).Methods("POST")
    r.Handle("/behoeftebepaling/ecd/client/{clientId}", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.GetClientHandler))).Methods("GET")
    r.Handle("/behoeftebepaling/ecd/zorgdossier", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.KoppelZorgdossierHandler))).Methods("POST")
    r.Handle("/behoeftebepaling/ecd/zorgdossier/client/{clientId}", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.GetZorgdossierByClientIDHandler))).Methods("GET")
    r.Handle("/behoeftebepaling/ecd/onderzoek", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.KoppelOnderzoekHandler))).Methods("POST")
    r.Handle("/behoeftebepaling/ecd/onderzoek/{onderzoekId}", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.GetOnderzoekByIdHandler))).Methods("GET")

    log.Printf("Behoeftebepaling-service draait op %s...", cfg.ServerPort)
    log.Fatal(http.ListenAndServe(cfg.ServerPort, r))
}
