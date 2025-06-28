// package main

// import (
// 	"behoeftebepaling/handlers"
// 	"fmt"
// 	"net/http"

// 	"github.com/gorilla/mux"
// )

// func main() {
// 	r := mux.NewRouter()
// 	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
// 	r.HandleFunc("/behoefte", handlers.CreateBehoefte).Methods("POST")
// 	r.HandleFunc("/behoefte/onderzoek/{onderzoekId}", handlers.GetBehoefteByOnderzoekID).Methods("GET")
// 	r.HandleFunc("/onderzoek/{onderzoekId}/anamnese", handlers.KoppelAnamneseHandler).Methods("POST")
// 	r.HandleFunc("/onderzoek/{onderzoekId}/meetresultaat", handlers.KoppelMeetresultaatHandler).Methods("POST")
// 	r.HandleFunc("/onderzoek/{onderzoekId}/diagnose", handlers.KoppelDiagnoseHandler).Methods("POST")
// 	fmt.Println("Server is running on port 8080")
// 	http.ListenAndServe(":8080", r)
// }

package main

import (
    "log"
    "net/http"
    "behoeftebepaling/handlers"
    "behoeftebepaling/pkg/config"
    behoefte_repo "behoeftebepaling/repository"
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
    db, err := behoefte_repo.InitDB(cfg.DatabaseDSN)
    if err != nil {
        log.Fatalf("Database initialization failed: %v", err)
    }
    handlers.InitHandlers(db) // als je dependency injection gebruikt

    r := mux.NewRouter()
    r.HandleFunc("/behoefte", handlers.CreateBehoefte).Methods("POST")
    r.HandleFunc("/behoefte/onderzoek/{onderzoekId}", handlers.GetBehoefteByOnderzoekID).Methods("GET")
    r.HandleFunc("/onderzoek/{onderzoekId}/anamnese", handlers.KoppelAnamneseHandler).Methods("POST")
    r.HandleFunc("/onderzoek/{onderzoekId}/meetresultaat", handlers.KoppelMeetresultaatHandler).Methods("POST")
    r.HandleFunc("/onderzoek/{onderzoekId}/diagnose", handlers.KoppelDiagnoseHandler).Methods("POST")

    log.Printf("Behoeftebepaling-service draait op %s...", cfg.ServerPort)
    log.Fatal(http.ListenAndServe(cfg.ServerPort, r))
}
