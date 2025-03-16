package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/edalmava/student-behavior-api/internal/db/models"
	"github.com/edalmava/student-behavior-api/internal/db/sqlite"

	"github.com/gorilla/mux"
)

// Handler para obtener todos los grados
func GetGrades(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	rows, err := db.Query("SELECT id, name FROM grades WHERE active = 1 ORDER BY name")
	if err != nil {
		http.Error(w, "Error al obtener grados", http.StatusInternalServerError)
		log.Printf("Error al consultar grados: %v", err)
		return
	}
	defer rows.Close()

	grades := []models.Grade{}
	for rows.Next() {
		var grade models.Grade
		if err := rows.Scan(&grade.ID, &grade.Name); err != nil {
			http.Error(w, "Error al procesar grados", http.StatusInternalServerError)
			log.Printf("Error al escanear grado: %v", err)
			return
		}
		grades = append(grades, grade)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grades)
}

// Get all grades (para el panel admin)
func GetGradesHandler(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	rows, err := db.Query("SELECT id, name, active FROM grades")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var grades []models.Grade
	for rows.Next() {
		var g models.Grade
		err := rows.Scan(&g.ID, &g.Name, &g.Active)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		grades = append(grades, g)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grades)
}

// Create new grade
func CreateGrade(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	var grade models.Grade
	err := json.NewDecoder(r.Body).Decode(&grade)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO grades (name, active) VALUES (?, ?)", grade.Name, grade.Active)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	grade.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(grade)
}

// Toggle grade status
func ToggleGrade(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE grades SET active = NOT active WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Estado actualizado"})
}

func DeleteGrade(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM grades WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Grado eliminado"})
}

// UpdateGradeHandler actualiza los datos de un grado
func UpdateGrade(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var updatedGrade models.Grade
	err = json.NewDecoder(r.Body).Decode(&updatedGrade)
	if err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Validar que el nombre no esté vacío
	if updatedGrade.Name == "" {
		http.Error(w, "El nombre del grado es requerido", http.StatusBadRequest)
		return
	}

	// Actualizar en la base de datos
	_, err = db.Exec("UPDATE grades SET name = ?, active = ? WHERE id = ?",
		updatedGrade.Name,
		updatedGrade.Active,
		id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Obtener el grado actualizado para devolverlo
	var grade models.Grade
	row := db.QueryRow("SELECT id, name, active FROM grades WHERE id = ?", id)
	err = row.Scan(&grade.ID, &grade.Name, &grade.Active)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grade)
}
