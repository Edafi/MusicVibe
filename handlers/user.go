package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Edafi/MusicVibe/models"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	DB *sql.DB
}

func (handler *UserHandler) GetUsers(response http.ResponseWriter, request *http.Request) {
	rows, err := handler.DB.Query("SELECT id, email, passwd_hash, role, display_name, avatar_path, creation_date FROM user")
	if err != nil {
		http.Error(response, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswdHash, &u.Role, &u.DisplayName, &u.AvatarPath, &u.CreationDate); err != nil {
			http.Error(response, "Scan error", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(users)
}

func (handler *UserHandler) GetUser(response http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	id := params["id"]

	var user models.User
	err := handler.DB.QueryRow("SELECT id, email, passwd_hash, role, display_name, avatar_path, creation_date FROM user WHERE id = ?", id).
		Scan(&user.ID, &user.Email, &user.PasswdHash, &user.Role, &user.DisplayName, &user.AvatarPath, &user.CreationDate)
	if err == sql.ErrNoRows {
		http.NotFound(response, request)
		return
	} else if err != nil {
		http.Error(response, "Database error", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(user)
}

func (handler *UserHandler) CreateUser(response http.ResponseWriter, request *http.Request) {
	var u models.User
	if err := json.NewDecoder(request.Body).Decode(&u); err != nil {
		http.Error(response, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := handler.DB.Exec("INSERT INTO user (id, email, passwd_hash, role, display_name, avatar_path) VALUES (?, ?, ?, ?, ?, ?)",
		u.ID, u.Email, u.PasswdHash, u.Role, u.DisplayName, u.AvatarPath)
	if err != nil {
		http.Error(response, "Database insert error", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(u)
}

func (handler *UserHandler) DeleteUser(response http.ResponseWriter, request *http.Request) {
	id := mux.Vars(request)["id"]
	_, err := handler.DB.Exec("DELETE FROM user WHERE id = ?", id)
	if err != nil {
		http.Error(response, "Failed to delete user", http.StatusInternalServerError)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (handler *UserHandler) UpdateUser(response http.ResponseWriter, request *http.Request) {
	id := mux.Vars(request)["id"]
	var user models.User
	if err := json.NewDecoder(request.Body).Decode(&user); err != nil {
		http.Error(response, "Invalid request body", http.StatusBadRequest)
		return
	}
	_, err := handler.DB.Exec("UPDATE user SET email=?, passwd_hash=?, role=?, display_name=?, avatar_path=? WHERE id=?",
		user.Email, user.PasswdHash, user.Role, user.DisplayName, user.AvatarPath, id)
	if err != nil {
		http.Error(response, "Failed to update user", http.StatusInternalServerError)
		return
	}
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(user)
}
