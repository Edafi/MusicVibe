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

	homeHandler := &handlers.HomeHandler{DB: db}
	secured := router.PathPrefix("/").Subrouter()
	secured.Use(middleware.JWTMiddleware)

	// жанровые обработчики
	genreHandler := &handlers.GenreHandler{DB: db}
	secured.HandleFunc("/genres", genreHandler.GetGenres).Methods("GET")
	secured.HandleFunc("/user/genres", genreHandler.PostUserGenres).Methods("POST")

	// обработчики музыкантов
	musicianHandler := &handlers.MusicianHandler{DB: db}
	secured.HandleFunc("/musicians", musicianHandler.GetMusicians).Methods("GET")
	secured.HandleFunc("/user/following", musicianHandler.PostUserFollowing).Methods("POST")
	secured.HandleFunc("/musician/{id}", musicianHandler.GetMusician).Methods("GET")
	secured.HandleFunc("/musician/{id}/popular-tracks", musicianHandler.GetPopularTracks).Methods("GET")

	// обработчики регистрации/логина
	authHandler := &handlers.AuthHandler{DB: db}
	router.HandleFunc("/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/auth/me", authHandler.Me).Methods("GET")

	secured.HandleFunc("/tracks/recommended", homeHandler.GetRecommendedTracks).Methods("GET")
	secured.HandleFunc("/albums/recommended", homeHandler.GetRecommendedAlbums).Methods("GET")
	secured.HandleFunc("/tracks/tracked", homeHandler.GetTrackedTracks).Methods("GET")

	trackHandler := &handlers.SearchHandler{DB: db}
	secured.HandleFunc("/tracks/new", trackHandler.GetNewTracks).Methods("GET")
	secured.HandleFunc("/tracks/chart", trackHandler.GetChartTracks).Methods("GET")
	secured.HandleFunc("/tracks/search", trackHandler.SearchTracks).Methods("GET")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		Debug:          false,
	})

	return c.Handler(router)
}
