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
	config, err := config.LoadConfig(".env")
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

	http.HandleFunc("/product/suppliers", handlers.HaalProductLeveraarsOp)
	http.HandleFunc("/product", handlers.HaalProductenOp)
	http.HandleFunc("/categorieen", handlers.HaalCategorieenOp)
	http.HandleFunc("/review", handlers.PlaatsReview)
	http.HandleFunc("/product/offer", handlers.VoegProductAanbodToe)
	http.HandleFunc("/product/add", handlers.VoegNieuwProductToe)

	log.Printf("Product-service draait op %s...", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, nil))
}
