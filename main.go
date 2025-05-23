package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Edafi/MusicVibe/routes"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "golang:QXmREMXPd321Edafi@tcp(localhost:3306)/audiostream")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	handler := routes.SetupRoutes(db)
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
