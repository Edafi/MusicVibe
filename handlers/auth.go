package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("1111")

type Credentials struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthHandler struct {
	DB *sql.DB
}

// Регистрация
func (handler *AuthHandler) Register(response http.ResponseWriter, request *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(request.Body).Decode(&creds); err != nil {
		log.Println("Decode error:", err)
		http.Error(response, "Invalid input", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Hash error:", err)
		http.Error(response, "Error hashing password", http.StatusInternalServerError)
		return
	}

	id := uuid.New().String()
	role := "user"

	query := `INSERT INTO user (id, username, email, passwd_hash, role, has_complete_setup) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = handler.DB.Exec(query, id, creds.Username, creds.Email, hashedPassword, role, false)
	if err != nil {
		log.Println("Insert error:", err)
		http.Error(response, "Error inserting user", http.StatusInternalServerError)
		return
	}

	// Создание JWT
	expiration := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: id,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println("Token error:", err)
		http.Error(response, "Error signing token", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(map[string]interface{}{
		"token": tokenString,
		"user": map[string]interface{}{
			"id":                id,
			"email":             creds.Email,
			"username":          creds.Username,
			"hasCompletedSetup": false, // по умолчанию
		},
	})
}

// Логин
func (handler *AuthHandler) Login(response http.ResponseWriter, request *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(request.Body).Decode(&creds); err != nil {
		log.Println("Login decode error:", err)
		http.Error(response, "Invalid input", http.StatusBadRequest)
		return
	}

	var userID, username, email, passwordHash, role string
	query := `SELECT id, username, email, passwd_hash, role FROM user WHERE email = ?`
	err := handler.DB.QueryRow(query, creds.Email).Scan(&userID, &username, &email, &passwordHash, &role)
	if err != nil {
		log.Println("User not found:", err)
		http.Error(response, "User not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(creds.Password)); err != nil {
		log.Println("Password mismatch:", err)
		http.Error(response, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expiration := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println("Token error:", err)
		http.Error(response, "Error signing token", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(map[string]interface{}{
		"token": tokenString,
		"user": map[string]interface{}{
			"id":                userID,
			"email":             email,
			"username":          username,
			"hasCompletedSetup": false, // можно потом получить это из базы, если добавишь флаг
		},
	})
}

func (handler *AuthHandler) Me(response http.ResponseWriter, request *http.Request) {
	authHeader := request.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("Missing token")
		http.Error(response, "Missing token", http.StatusUnauthorized)
		return
	}

	tokenString := authHeader[len("Bearer "):]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		log.Println("Invalid token")
		http.Error(response, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Получаем основные поля пользователя
	var name, email, avatarUrl, backgroundUrl, description string
	var hasCompletedSetup bool
	query := `
		SELECT username, email, avatar_path, background_path, description, has_complete_setup 
		FROM user
		WHERE id = ?`
	err = handler.DB.QueryRow(query, claims.UserID).Scan(
		&name, &email, &avatarUrl, &backgroundUrl, &description, &hasCompletedSetup)
	if err != nil {
		log.Println("UserID from token:", claims.UserID)
		log.Println("User not found")
		http.Error(response, "User not found", http.StatusNotFound)
		return
	}

	// Получаем жанры пользователя
	genres := []string{}
	rows, err := handler.DB.Query(`
		SELECT g.name
		FROM genre g
		JOIN user_genre ug ON g.id = ug.genre_id
		WHERE ug.user_id = ?`, claims.UserID)
	if err != nil {
		log.Println("Genres error:", err)
	}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var genre string
			if err := rows.Scan(&genre); err == nil {
				genres = append(genres, genre)
			}
		}
	}

	// Получаем соцсети пользователя
	socialLinks := map[string]string{}
	socialRows, err := handler.DB.Query(`
		SELECT sn.name, us.url
		FROM social_network sn
		JOIN user_social_network us ON sn.id = us.social_id
		WHERE us.user_id = ?`, claims.UserID)
	if err != nil {
		log.Println("Social_networks error:", err)
	}
	if err == nil {
		defer socialRows.Close()
		for socialRows.Next() {
			var name, url string
			if err := socialRows.Scan(&name, &url); err == nil {
				socialLinks[name] = url
			}
		}
	}

	// Отправляем ответ
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":                claims.UserID,
			"name":              name,
			"email":             email,
			"avatarUrl":         avatarUrl,
			"backgroundUrl":     backgroundUrl,
			"description":       description,
			"genres":            genres,
			"hasCompletedSetup": hasCompletedSetup,
			"socialLinks":       socialLinks,
		},
	})
}
