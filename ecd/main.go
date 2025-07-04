package main

import (
	handler "ecd/api"
	"ecd/config"
	database "ecd/data"
	"ecd/data/repository"
	"ecd/service"
	"log"
	"net/http"

	middleware_ecd "ecd/api/middleware"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

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
	r.Use(middleware_ecd.CorsMiddleware(config.CorsOrigin))

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
