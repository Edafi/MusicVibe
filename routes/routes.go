package routes

import (
	"database/sql"
	"net/http"

	"github.com/Edafi/MusicVibe/handlers"
	"github.com/Edafi/MusicVibe/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(db *sql.DB, mongoDatabase *mongo.Database) http.Handler {
	router := mux.NewRouter()

	// обработчики пользователя
	userHandler := &handlers.UserHandler{DB: db}
	router.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	router.HandleFunc("/user/{id}", userHandler.GetUser).Methods("GET")
	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

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
	secured.HandleFunc("/auth/me", authHandler.Me).Methods("GET")

	homeHandler := &handlers.HomeHandler{DB: db}
	secured.HandleFunc("/tracks/recommended", homeHandler.GetRecommendedTracks).Methods("GET")
	secured.HandleFunc("/albums/recommended", homeHandler.GetRecommendedAlbums).Methods("GET")
	secured.HandleFunc("/tracks/tracked", homeHandler.GetTrackedTracks).Methods("GET")
	secured.HandleFunc("/home/tracks/recommended", homeHandler.GetHomeRecommendedTracks).Methods("GET")
	secured.HandleFunc("/home/albums/recommended", homeHandler.GetHomeRecommendedAlbums).Methods("GET")
	secured.HandleFunc("/home/tracks/tracked", homeHandler.GetHomeTrackedTracks).Methods("GET")

	searchHandler := &handlers.SearchHandler{DB: db}
	secured.HandleFunc("/tracks/new", searchHandler.GetNewTracks).Methods("GET")
	secured.HandleFunc("/tracks/chart", searchHandler.GetChartTracks).Methods("GET")
	secured.HandleFunc("/tracks/search", searchHandler.SearchTracks).Methods("GET")

	trackHandler := &handlers.TrackHandler{DB: db}
	secured.HandleFunc("/track/{id}", trackHandler.GetTrack).Methods("GET")

	commentHandler := &handlers.CommentHandler{DB: db, MongoDatabase: mongoDatabase}
	secured.HandleFunc("/comments/track/{id}", commentHandler.GetTrackComments).Methods("GET")
	secured.HandleFunc("/comments/track/{id}", commentHandler.PostTrackComment).Methods("POST")

	albumHandler := &handlers.AlbumHandler{DB: db}
	secured.HandleFunc("/album/{id}", albumHandler.GetAlbum).Methods("GET")
	secured.HandleFunc("/album/{id}/tracks", albumHandler.GetAlbumTracks).Methods("GET")

	favorites := &handlers.FavoritesHandler{DB: db}
	secured.HandleFunc("/favorites", favorites.GetFavoriteTracks).Methods("GET")
	secured.HandleFunc("/favorites/{id}", favorites.AddFavoriteTrack).Methods("POST")
	secured.HandleFunc("/favorites/{id}", favorites.DeleteFavoriteTrack).Methods("DELETE")
	secured.HandleFunc("/favorites/albums", favorites.GetFavoriteAlbums).Methods("GET")
	secured.HandleFunc("/favorites/albums/{id}", favorites.AddFavoriteAlbum).Methods("POST")
	secured.HandleFunc("/favorites/albums/{id}", favorites.DeleteFavoriteAlbum).Methods("DELETE")

	following := &handlers.FollowingHandler{DB: db}
	secured.HandleFunc("/following", following.GetFollowingMusicians).Methods("GET")
	secured.HandleFunc("/following/{id}", following.FollowMusician).Methods("POST")
	secured.HandleFunc("/following/{id}", following.UnfollowMusician).Methods("DELETE")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		Debug:          false,
	})

	return c.Handler(router)
}
