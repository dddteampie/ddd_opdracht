package main

import (
	"log"
	"net/http"

	"product/handlers"
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

	mux := http.NewServeMux()

	mux.HandleFunc("/product/suppliers", handlers.HaalProductLeveraarsOp)
	mux.HandleFunc("/product", handlers.HaalProductenOp)
	mux.HandleFunc("/categorieen", handlers.HaalCategorieenOp)
	mux.HandleFunc("/review", handlers.PlaatsReview)
	mux.HandleFunc("/product/offer", handlers.VoegProductAanbodToe)
	mux.HandleFunc("/product/add", handlers.VoegNieuwProductToe)
	mux.HandleFunc("/categorieen/tags", handlers.HaalTagsOp)

	log.Printf("Product-service draait op %s...", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, mux))
}
