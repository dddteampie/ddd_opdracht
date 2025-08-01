package main

import (
	middleware_bestel "bestel/api/middleware"
	"bestel/config"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Zorg dus dat je een .env in de root van bestel folder hebt
	config, err := config.LoadConfig("bestel.env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	r.Use(middleware_bestel.CorsMiddleware(config.CorsOrigin))

	r.Route("/bestel", func(r chi.Router) {
		r.Route("/health", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})
		})
	})
	http.ListenAndServe(config.ServerPort, r)
}
