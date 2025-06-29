package main

import (
    "log"
    "net/http"
    "aanvraagverwerking/handlers"
    "aanvraagverwerking/pkg/config"
    aanvraagverwerking_repo "aanvraagverwerking/repository"
    "github.com/gorilla/mux"
)

func main() {
    // Load config from .env
    cfg, err := config.LoadConfig(".env")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    log.Printf("Configuration loaded: DatabaseDSN=%s, ServerPort=%s", cfg.DatabaseDSN, cfg.ServerPort)

    // Initialize database
    db, err := aanvraagverwerking_repo.InitDB(cfg.DatabaseDSN)
    if err != nil {
        log.Fatalf("Database initialization failed: %v", err)
    }
    handlers.InitHandlers(db) // als je dependency injection gebruikt

    r := mux.NewRouter()
    r.HandleFunc("/aanvraag", handlers.StartAanvraag).Methods("POST")

	
    log.Printf("Behoeftebepaling-service draait op %s...", cfg.ServerPort)
    log.Fatal(http.ListenAndServe(cfg.ServerPort, r))
}