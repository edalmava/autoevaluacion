// internal/api/handlers/users.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/edalmava/student-behavior-api/internal/db/models"
	"github.com/edalmava/student-behavior-api/internal/db/sqlite"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// GetUsersHandler obtiene todos los usuarios (solo administradores)
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	rows, err := db.Query("SELECT id, username, role, active FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Active)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// UpdateUserHandler actualiza un usuario existente
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	db := sqlite.GetDB()

	// Si se proporciona una contraseña nueva, actualizarla
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error al procesar la contraseña", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("UPDATE users SET username = ?, password = ?, role = ? WHERE id = ?",
			user.Username, string(hashedPassword), user.Role, userID)
	} else {
		// Si no se proporciona contraseña, actualizar solo nombre y rol
		_, err = db.Exec("UPDATE users SET username = ?, role = ? WHERE id = ?",
			user.Username, user.Role, userID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Obtener usuario actualizado (sin la contraseña)
	var updatedUser models.User
	err = db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", userID).Scan(
		&updatedUser.ID, &updatedUser.Username, &updatedUser.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

// DeleteUserHandler elimina un usuario
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// No permitir eliminar al usuario admin principal (ID 1)
	if userID == 1 {
		http.Error(w, "No se puede eliminar al usuario administrador principal", http.StatusForbidden)
		return
	}

	db := sqlite.GetDB()
	_, err = db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario eliminado correctamente"})
}

// GetUserHandler maneja la solicitud para obtener un usuario específico por su ID
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del usuario de los parámetros de la URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
		return
	}

	db := sqlite.GetDB()

	// Consultar el usuario
	user := models.User{}
	// Nota: No devolvemos el password en la consulta por seguridad
	err = db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Usuario no encontrado", http.StatusNotFound)
		} else {
			http.Error(w, "Error al consultar el usuario", http.StatusInternalServerError)
		}
		return
	}

	// Configurar encabezados y enviar respuesta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ToggleUserHandler cambia el estado activo/inactivo de un usuario
func ToggleUserHandler(w http.ResponseWriter, r *http.Request) {
	// Obtener ID del usuario de los parámetros de la URL
	params := mux.Vars(r)
	idStr := params["id"]

	// Convertir string a int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
		return
	}

	// No permitir desactivar al usuario administrador principal (ID=1)
	if id == 1 {
		http.Error(w, "No se puede cambiar el estado del administrador principal", http.StatusForbidden)
		return
	}

	// Obtener usuario actual de la base de datos
	user, err := getUserByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Usuario no encontrado", http.StatusNotFound)
		} else {
			http.Error(w, "Error al obtener usuario: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Invertir el estado activo
	newActiveState := !user.Active

	// Actualizar estado en la base de datos
	err = updateUserActiveStatus(id, newActiveState)
	if err != nil {
		http.Error(w, "Error al actualizar estado del usuario: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Responder con éxito (sin contenido)
	w.WriteHeader(http.StatusNoContent)
}

// Función auxiliar para actualizar solo el estado activo de un usuario
func updateUserActiveStatus(id int, active bool) error {
	// Preparar la consulta SQL
	query := `UPDATE users SET active = $1 WHERE id = $2`

	db := sqlite.GetDB()

	// Ejecutar la consulta
	_, err := db.Exec(query, active, id)
	return err
}

func getUserByID(id int) (models.User, error) {
	var user models.User

	db := sqlite.GetDB()

	// Consulta SQL para obtener usuario por ID
	query := `SELECT id, username, role, active FROM users WHERE id = $1`

	// Ejecutar la consulta
	row := db.QueryRow(query, id)

	// Escanear resultados
	err := row.Scan(&user.ID, &user.Username, &user.Role, &user.Active)
	if err != nil {
		return user, err
	}

	return user, nil
}
