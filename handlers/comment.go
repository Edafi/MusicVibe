package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Edafi/MusicVibe/middleware"
	"github.com/Edafi/MusicVibe/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentHandler struct {
	DB            *sql.DB
	MongoDatabase *mongo.Database
}

func (handler *CommentHandler) GetTrackComments(response http.ResponseWriter, request *http.Request) {
	trackID := mux.Vars(request)["id"]

	// Получение комментариев из MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := handler.MongoDatabase.Collection("track_comments")

	filter := bson.M{"track_id": trackID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Println("MongoDB error:", err)
		http.Error(response, "Error fetching comments", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []models.CommentResponse = make([]models.CommentResponse, 0)

	for cursor.Next(ctx) {
		var comment struct {
			ID        primitive.ObjectID `bson:"_id"`
			TrackID   string             `bson:"track_id"`
			UserID    string             `bson:"user_id"`
			Comment   string             `bson:"comment"`
			CreatedAt time.Time          `bson:"created_at"`
		}

		if err := cursor.Decode(&comment); err != nil {
			log.Println("Decode error:", err)
			continue
		}

		// Получение информации о пользователе из MariaDB
		var user models.CommentAuthor
		user.ID = comment.UserID

		query := `SELECT name, avatar_path FROM user WHERE id = ?`
		err := handler.DB.QueryRow(query, comment.UserID).Scan(&user.Name, &user.AvatarURL)
		if err != nil {
			log.Println("SQL error for user", comment.UserID, ":", err)
			user.Name = "Неизвестный пользователь"
			user.AvatarURL = "/avatarUser/default.png"
		}

		results = append(results, models.CommentResponse{
			ID:        comment.ID.Hex(),
			Text:      comment.Comment,
			CreatedAt: comment.CreatedAt,
			User:      user,
		})
	}

	if err := cursor.Err(); err != nil {
		log.Println("Cursor error:", err)
		http.Error(response, "Error reading comments", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(results)
}

func (handler *CommentHandler) PostTrackComment(response http.ResponseWriter, request *http.Request) {
	userID := request.Context().Value(middleware.ContextUserIDKey).(string)
	trackID := mux.Vars(request)["id"]

	var req models.CreateCommentRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(response, "Invalid input", http.StatusBadRequest)
		return
	}

	comment := models.TrackComment{
		TrackID:   trackID,
		UserID:    userID,
		Comment:   req.Text,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := handler.MongoDatabase.Collection("track_comments").InsertOne(ctx, comment)
	if err != nil {
		log.Println("PostTrackComment - Error inserting comment:", err)
		http.Error(response, "Error saving comment", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusCreated)
}
