package main

import (
	handler "ecd/api"
	"ecd/config"
	database "ecd/data"
	"ecd/data/repository"
	"ecd/service"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func corsMiddleware(allowedOrigin string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestOrigin := r.Header.Get("Origin")

			if allowedOrigin == "*" || requestOrigin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
			} else {
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	handler := &handler.Handler{}

	config, err := config.LoadConfig("ecd.env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.InitDB(config.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	r.Use(corsMiddleware(config.CorsOrigin))

	repository := repository.NewGormRepository(db)
	service := service.NewECDService(repository)
	handler.ECD = service
	r.Route("/ecd", func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Route("/health", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					log.Println("Health check endpoint hit")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("OK"))
				})
			})

			r.Route("/client", func(r chi.Router) {
				r.Get("/{id}", handler.GetClientHandler)
				r.Get("/", handler.GetAllClientsHandler)
				r.Post("/", handler.CreateClientHandler)
			})

			r.Route("/zorgdossier", func(r chi.Router) {
				r.Get("/client/{clientId}", handler.GetZorgdossierByClientIDHandler)
				r.Post("/", handler.CreateZorgdossierHandler)
			})

			r.Route("/onderzoek", func(r chi.Router) {
				r.Post("/", handler.CreateOnderzoekHandler)
				r.Post("/{onderzoekId}/anamnese", handler.AddAnamneseHandler)
				r.Post("/{onderzoekId}/meetresultaat", handler.AddMeetresultaatHandler)
				r.Post("/{onderzoekId}/diagnose", handler.AddDiagnoseHandler)
				r.Get("/{onderzoekId}", handler.GetOnderzoekByIDHandler)
				r.Put("/{onderzoekId}", handler.UpdateOnderzoekHandler)
				r.Get("/dossier/{dossierId}", handler.GetOnderzoekByDossierIdHandler)
			})
		})

	})
	log.Printf("ECD service is running on %s...", config.ServerPort)
	http.ListenAndServe(config.ServerPort, r)
}
