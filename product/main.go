package main

import (
	"log"
	"net/http"

	"productservice/data_access/data_handling" // Import the new database package
	"productservice/handlers"
)

func main() {
	//load correct config
	config, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded: DatabaseDSN=%s, ServerPort=%s", config.DatabaseDSN, config.ServerPort)

	//initialize config
	db, err := data_handling.InitDB(config.DatabaseDSN)
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

	log.Printf("ProductService draait op %s...", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, nil))
}
