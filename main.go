package main

import (
	"log"
	"net/http"

	"github.com/Edafi/MusicVibe/routes"
)

func main() {
	handler := routes.SetupRoutes()
	log.Println("Server is running on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", handler))
}
