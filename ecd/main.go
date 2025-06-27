package main

import (
	handler "ecd/api"
	"ecd/config"
	database "ecd/data"
	"ecd/data/repository"
	"ecd/service"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	handler := &handler.Handler{}

	config, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded: DatabaseDSN=%s, ServerPort=%s", config.DatabaseDSN, config.ServerPort)

	//dsn := "host=localhost user=Admin password=Admin1232 dbname=ecd port=5432 sslmode=disable TimeZone=Europe/Amsterdam"
	db, err := database.InitDB(config.DatabaseDSN)
	if err != nil {
		panic("Failed to connect to the database: " + err.Error())
	}
	repository := repository.NewGormRepository(db)
	service := service.NewECDService(repository)
	fmt.Print(service)

	// Swagger docs endpoint
	r.Get("/swagger/*", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./docs/swagger"))).ServeHTTP)

	r.Route("/api", func(r chi.Router) {
		r.Route("/health", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})
		})

		r.Route("/client", func(r chi.Router) {
			r.Get("/{id}", handler.GetClientHandler)
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
		})
	})

	http.ListenAndServe(":8082", r)
}
