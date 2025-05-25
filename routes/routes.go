package routes

import (
	"database/sql"
	"net/http"

	"github.com/Edafi/MusicVibe/handlers"
	"github.com/Edafi/MusicVibe/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func SetupRoutes(db *sql.DB) http.Handler {
	router := mux.NewRouter()

	// обработчики пользователя
	userHandler := &handlers.UserHandler{DB: db}
	router.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	router.HandleFunc("/user/{id}", userHandler.GetUser).Methods("GET")
	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// жанровые обработчики
	genreHandler := &handlers.GenreHandler{DB: db}
	router.HandleFunc("/genres", genreHandler.GetGenres).Methods("GET")
	router.HandleFunc("/user/genres", genreHandler.PostUserGenres).Methods("POST")

	// обработчики музыкантов
	musicianHandler := &handlers.MusicianHandler{DB: db}
	router.HandleFunc("/musicians", musicianHandler.GetMusicians).Methods("GET")
	router.HandleFunc("/user/following", musicianHandler.PostUserFollowing).Methods("POST")

	// обработчики регистрации/логина
	authHandler := &handlers.AuthHandler{DB: db}
	router.HandleFunc("/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/login", authHandler.Login).Methods("POST")

	homeHandler := &handlers.HomeHandler{DB: db}
	secured := router.PathPrefix("/").Subrouter()
	secured.Use(middleware.JWTMiddleware)

	secured.HandleFunc("/tracks/recommended", homeHandler.GetRecommendedTracks).Methods("GET")
	secured.HandleFunc("/albums/recommended", homeHandler.GetRecommendedAlbums).Methods("GET")
	secured.HandleFunc("/tracks/tracked", homeHandler.GetTrackedTracks).Methods("GET")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
	})

	return c.Handler(router)
}
