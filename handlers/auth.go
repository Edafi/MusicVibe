package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"log"

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

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Hash error:", err)
		http.Error(response, "Error hashing password", http.StatusInternalServerError)
		return
	}

	id := uuid.New().String()
	role := "user"

	// Вставка пользователя
	query := `INSERT INTO user (id, name, email, passwd_hash, role) VALUES (?, ?, ?, ?, ?)`
	_, err = handler.DB.Exec(query, id, creds.Username, creds.Email, hashedPassword, role)
	if err != nil {
		log.Println("Insert error:", err)
		http.Error(response, "Error inserting user", http.StatusInternalServerError)
		return
	}

	// Создание JWT-токена
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

	// Отправка токена клиенту
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(map[string]string{
		"token": tokenString,
	})
}

// Логин
func (handler *AuthHandler) Login(response http.ResponseWriter, request *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(request.Body).Decode(&creds); err != nil {
		log.Println("Insert error:", err)
		http.Error(response, "Invalid input", http.StatusBadRequest)
		return
	}

	var userID, passwordHash, role string
	err := handler.DB.QueryRow("SELECT id, passwd_hash, role FROM user WHERE email = ?", creds.Email).Scan(&userID, &passwordHash, &role)
	if err != nil {
		log.Println("Insert error:", err)
		http.Error(response, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(creds.Password)); err != nil {
		log.Println("Insert error:", err)
		http.Error(response, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Создание токена
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
		log.Println("Insert error:", err)
		http.Error(response, "Error signing token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(response).Encode(map[string]string{
		"token": tokenString,
	})
}
