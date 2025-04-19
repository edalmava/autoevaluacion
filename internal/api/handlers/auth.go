// internal/api/handlers/auth.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edalmava/autoevaluacion/internal/api/middleware"
	"github.com/edalmava/autoevaluacion/internal/db/models"
	"github.com/edalmava/autoevaluacion/internal/db/sqlite"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest representa los datos de inicio de sesión
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse representa la respuesta de inicio de sesión
type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// LoginHandler maneja el inicio de sesión
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Buscar usuario en la base de datos
	db := sqlite.GetDB()
	var user models.User
	var hashedPassword string
	err := db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ? AND active = 1", req.Username).Scan(
		&user.ID, &user.Username, &hashedPassword, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generar token JWT
	token, err := middleware.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Devolver respuesta con token y datos del usuario
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: token,
		User:  user,
	})
}

// RegisterHandler maneja el registro de usuarios (solo para administradores)
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar que el usuario es administrador
	role, ok := r.Context().Value("role").(string)
	if !ok || role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar datos
	if user.Username == "" || user.Password == "" || user.Role == "" {
		http.Error(w, "Username, password, and role are required", http.StatusBadRequest)
		return
	}

	// Hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Insertar usuario en la base de datos
	db := sqlite.GetDB()
	result, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		user.Username, string(hashedPassword), user.Role)

	if err != nil {
		//http.Error(w, "Failed to create user", http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	// Obtener ID del nuevo usuario
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}
	user.ID = int(id)
	user.Password = "" // No devolver la contraseña

	// Devolver respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetMeHandler devuelve información del usuario autenticado
func GetMeHandler(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value("username").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Buscar usuario en la base de datos
	db := sqlite.GetDB()
	var user models.User
	err := db.QueryRow("SELECT id, username, role, active FROM users WHERE username = ?", username).Scan(
		&user.ID, &user.Username, &user.Role, &user.Active)

	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Devolver respuesta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
