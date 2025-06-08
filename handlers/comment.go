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
	"go.mongodb.org/mongo-driver/mongo/options"
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
	options := options.Find().SetSort(bson.D{{"created_at", -1}})
	cursor, err := collection.Find(ctx, filter, options)
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
		var musician models.CommentAuthor
		musician.ID = comment.UserID

		if comment.UserID == "" {
			log.Println("GetTrackComments - пустой user_id")
			musician.Name = "Неизвестный пользователь"
			musician.AvatarURL = "/avatarUser/default.png"
		} else {
			query := `SELECT name, avatar_path FROM musician WHERE id = ?`
			err := handler.DB.QueryRow(query, comment.UserID).Scan(&musician.Name, &musician.AvatarURL)
			if err == sql.ErrNoRows {
				log.Println("GetTrackComments - музыкант не найден:", comment.UserID)
				musician.Name = "Неизвестный пользователь"
				musician.AvatarURL = "/avatarUser/default.png"
			} else if err != nil {
				log.Println("GetTrackComments - SQL ошибка:", err)
				musician.Name = "Неизвестный пользователь"
				musician.AvatarURL = "/avatarUser/default.png"
			}
		}

		results = append(results, models.CommentResponse{
			ID:        comment.ID.Hex(),
			Text:      comment.Comment,
			CreatedAt: comment.CreatedAt,
			User:      musician,
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

	var musicianID, musician_avatar_path string
	query := `SELECT m.id, m.avatar_path FROM musician AS m
	JOIN user ON user.id = m.user_id
	WHERE user.id = ?`
	err := handler.DB.QueryRow(query, userID).Scan(&musicianID, &musician_avatar_path)
	if err != nil {
		log.Println("PostTrackComment 1 - SQL error for user", userID, ":", err)
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(response, "Invalid input", http.StatusBadRequest)
		return
	}

	createdAt := time.Now()

	comment := models.TrackComment{
		TrackID:   trackID,
		UserID:    musicianID,
		Comment:   req.Text,
		CreatedAt: createdAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := handler.MongoDatabase.Collection("track_comments").InsertOne(ctx, comment)
	if err != nil {
		log.Println("PostTrackComment - Error inserting comment:", err)
		http.Error(response, "Error saving comment", http.StatusInternalServerError)
		return
	}

	var user models.CommentAuthor
	user.ID = userID
	query = `SELECT name, avatar_path FROM musician WHERE musician.id = ?`
	err = handler.DB.QueryRow(query, musicianID).Scan(&user.Name, &user.AvatarURL)
	if err != nil {
		log.Println("PostTrackComment 2 - SQL error for user", userID, ":", err)
		user.Name = "Неизвестный пользователь"
		user.AvatarURL = "/avatarUser/default.png"
	}

	commentResponse := models.CommentResponse{
		ID:        result.InsertedID.(primitive.ObjectID).Hex(),
		Text:      req.Text,
		CreatedAt: createdAt,
		User:      user,
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(commentResponse)
}
