package main

import (
	"log"
	"net/http"

	"product/handlers"
	"product/pkg/auth"
	"product/pkg/config"
	product_repo "product/repository"
)

func corsMiddleware(allowedOrigin string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestOrigin := r.Header.Get("Origin")

			if allowedOrigin == "*" || requestOrigin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
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
	config, err := config.LoadConfig("product.env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded")

	db, err := product_repo.InitDB(config.DatabaseDSN)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	handlers.InitHandlers(db)

	authConfig := auth.AuthZMiddlewareConfig{
		RolesClaimName: "realm_access",
		DevMode:        config.AuthzDevMode,
	}

	mux := http.NewServeMux()

	corsHandler := corsMiddleware(config.CorsOrigin)(mux)

	mux.Handle("/product/product/suppliers", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.HaalProductLeveraarsOp)))
	mux.Handle("/product/product", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.HaalProductenOp)))
	mux.Handle("/product/categorieen", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.HaalCategorieenOp)))
	mux.Handle("/product/review", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker", "user"}, http.HandlerFunc(handlers.PlaatsReview)))
	mux.Handle("/product/product/offer", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.VoegProductAanbodToe)))
	mux.Handle("/product/product/add", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.VoegNieuwProductToe)))
	mux.Handle("/product/categorieen/tags", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.HaalTagsOp)))
	mux.HandleFunc("/product/api/health", handlers.HealthCheckHandler)

	log.Printf("Product-service draait op %s...", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, corsHandler))
}
