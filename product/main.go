package main

import (
	"log"
	"net/http"

	"product/handlers"
	"product/pkg/auth"
	"product/pkg/config"
	product_repo "product/repository" // Import the new database package
)

func main() {
	//load correct config
	config, err := config.LoadConfig("product.env")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded")

	//initialize config
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

	mux.Handle("/product/suppliers", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.HaalProductLeveraarsOp)))
	mux.Handle("/product", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.HaalProductenOp)))
	mux.Handle("/categorieen", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.HaalCategorieenOp)))
	mux.Handle("/review", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker", "user"}, http.HandlerFunc(handlers.PlaatsReview)))
	mux.Handle("/product/offer", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.VoegProductAanbodToe)))
	mux.Handle("/product/add", auth.NewAuthZMiddleware(authConfig, []string{"healthcare_worker"}, http.HandlerFunc(handlers.VoegNieuwProductToe)))
	mux.Handle("/categorieen/tags", auth.NewAuthZMiddleware(authConfig, []string{}, http.HandlerFunc(handlers.HaalTagsOp)))

	log.Printf("Product-service draait op %s...", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, mux))
}
