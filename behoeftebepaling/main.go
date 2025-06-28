package main

import (
	"behoeftebepaling/handlers"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/behoefte", handlers.CreateBehoefte).Methods("POST")
	r.HandleFunc("/behoefte/onderzoek/{onderzoekId}", handlers.GetBehoefteByOnderzoekID).Methods("GET")
	r.HandleFunc("/onderzoek/{onderzoekId}/anamnese", handlers.KoppelAnamneseHandler).Methods("POST")
	r.HandleFunc("/onderzoek/{onderzoekId}/meetresultaat", handlers.KoppelMeetresultaatHandler).Methods("POST")
	r.HandleFunc("/onderzoek/{onderzoekId}/diagnose", handlers.KoppelDiagnoseHandler).Methods("POST")
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", r)
}
