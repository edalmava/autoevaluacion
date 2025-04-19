package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/edalmava/autoevaluacion/internal/db/models"
	"github.com/edalmava/autoevaluacion/internal/db/sqlite"

	"github.com/edalmava/autoevaluacion/internal/websocketapi"

	"github.com/gorilla/mux"
)

// Handler para obtener estudiantes por grado
func GetStudentsHandler(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	gradeName := r.URL.Query().Get("grade")
	if gradeName == "" {
		http.Error(w, "Se requiere especificar un grado", http.StatusBadRequest)
		return
	}

	// Obtener ID del grado
	var gradeID int
	err := db.QueryRow("SELECT id FROM grades WHERE name = ?", gradeName).Scan(&gradeID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Grado no encontrado", http.StatusNotFound)
		} else {
			http.Error(w, "Error al buscar grado", http.StatusInternalServerError)
			log.Printf("Error al buscar grado: %v", err)
		}
		return
	}

	// Obtener estudiantes del grado
	rows, err := db.Query("SELECT id, name, grade_id FROM students WHERE grade_id = ? ORDER BY name", gradeID)
	if err != nil {
		http.Error(w, "Error al obtener estudiantes", http.StatusInternalServerError)
		log.Printf("Error al consultar estudiantes: %v", err)
		return
	}
	defer rows.Close()

	students := []models.Student{}
	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.ID, &student.Name, &student.GradeID); err != nil {
			http.Error(w, "Error al procesar estudiantes", http.StatusInternalServerError)
			log.Printf("Error al escanear estudiante: %v", err)
			return
		}
		students = append(students, student)
	}

	response := models.StudentsResponse{
		Students: students,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handler para guardar una evaluación
func SaveEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	var evalReq models.EvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&evalReq); err != nil {
		http.Error(w, "Error al procesar la solicitud", http.StatusBadRequest)
		log.Printf("Error al decodificar solicitud: %v", err)
		return
	}

	// Comenzar una transacción
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		log.Printf("Error al iniciar transacción: %v", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	// 1. Verificar si el estudiante existe, si no, crearlo
	var studentID int
	var gradeID int

	err = tx.QueryRow("SELECT id FROM grades WHERE name = ?", evalReq.Grade).Scan(&gradeID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Grado no encontrado", http.StatusBadRequest)
		} else {
			http.Error(w, "Error al verificar grado", http.StatusInternalServerError)
			log.Printf("Error al buscar grado: %v", err)
		}
		return
	}

	err = tx.QueryRow("SELECT id FROM students WHERE name = ? AND grade_id = ?", evalReq.StudentName, gradeID).Scan(&studentID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Crear el estudiante
			result, err := tx.Exec("INSERT INTO students (name, grade_id) VALUES (?, ?)", evalReq.StudentName, gradeID)
			if err != nil {
				http.Error(w, "Error al crear estudiante", http.StatusInternalServerError)
				log.Printf("Error al insertar estudiante: %v", err)
				return
			}
			studentID64, err := result.LastInsertId()
			if err != nil {
				http.Error(w, "Error al obtener ID del estudiante", http.StatusInternalServerError)
				log.Printf("Error al obtener LastInsertId: %v", err)
				return
			}
			studentID = int(studentID64)
		} else {
			http.Error(w, "Error al verificar estudiante", http.StatusInternalServerError)
			log.Printf("Error al buscar estudiante: %v", err)
			return
		}
	}

	// 2. Crear una nueva evaluación
	evalDate, err := time.Parse(time.RFC3339, evalReq.Date)
	if err != nil {
		evalDate = time.Now() // Usar la hora actual si hay error en el formato
	}

	result, err := tx.Exec("INSERT INTO evaluations (student_id, date, average) VALUES (?, ?, ?)",
		studentID, evalDate, evalReq.Average)
	if err != nil {
		http.Error(w, "Error al guardar evaluación", http.StatusInternalServerError)
		log.Printf("Error al insertar evaluación: %v", err)
		return
	}

	evaluationID64, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error al obtener ID de evaluación", http.StatusInternalServerError)
		log.Printf("Error al obtener LastInsertId: %v", err)
		return
	}
	evaluationID := int(evaluationID64)

	// 3. Guardar las calificaciones individuales
	// Obtener todos los conceptos para mapear nombres a IDs
	rows, err := tx.Query("SELECT id, name FROM rating_concepts")
	if err != nil {
		http.Error(w, "Error al obtener conceptos", http.StatusInternalServerError)
		log.Printf("Error al consultar conceptos: %v", err)
		return
	}

	conceptMap := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			http.Error(w, "Error al procesar conceptos", http.StatusInternalServerError)
			log.Printf("Error al escanear concepto: %v", err)
			rows.Close()
			return
		}
		conceptMap[name] = id
	}
	rows.Close()

	// Procesar cada calificación
	for i := 1; i <= 10; i++ {
		ratingKey := fmt.Sprintf("rating%d", i)
		ratingValue, ok := evalReq.Ratings[ratingKey]
		if !ok {
			continue // Omitir si no está presente
		}

		// Obtener el ID del concepto correspondiente
		conceptName := fmt.Sprintf("concept%d", i)
		conceptID, ok := conceptMap[conceptName]
		if !ok {
			conceptID = i // Usar el índice si no hay coincidencia
		}

		_, err = tx.Exec("INSERT INTO ratings (evaluation_id, concept_id, value) VALUES (?, ?, ?)",
			evaluationID, conceptID, ratingValue)
		if err != nil {
			http.Error(w, "Error al guardar calificación", http.StatusInternalServerError)
			log.Printf("Error al insertar calificación: %v", err)
			return
		}
	}

	// Confirmar la transacción
	if err = tx.Commit(); err != nil {
		http.Error(w, "Error al finalizar la transacción", http.StatusInternalServerError)
		log.Printf("Error al hacer commit: %v", err)
		return
	}

	// Responder con éxito
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      evaluationID,
		"message": "Evaluación guardada correctamente",
	})

	websocketapi.NotifyEvaluationAdd(evaluationID, studentID, evalDate, evalReq.Average, gradeID, evalReq.StudentName)

}

// Handler para obtener evaluaciones de un estudiante
func GetStudentEvaluationsHandler(w http.ResponseWriter, r *http.Request) {
	db := sqlite.GetDB()

	vars := mux.Vars(r)
	studentID := vars["studentId"]

	// Obtener todas las evaluaciones del estudiante
	rows, err := db.Query(`
		SELECT e.id, e.date, e.average 
		FROM evaluations e
		WHERE e.student_id = ?
		ORDER BY e.date DESC
	`, studentID)
	if err != nil {
		http.Error(w, "Error al obtener evaluaciones", http.StatusInternalServerError)
		log.Printf("Error al consultar evaluaciones: %v", err)
		return
	}
	defer rows.Close()

	evaluations := []models.Evaluation{}
	for rows.Next() {
		var eval models.Evaluation
		if err := rows.Scan(&eval.ID, &eval.Date, &eval.Average); err != nil {
			http.Error(w, "Error al procesar evaluaciones", http.StatusInternalServerError)
			log.Printf("Error al escanear evaluación: %v", err)
			return
		}
		eval.StudentID, _ = strconv.Atoi(studentID)

		// Obtener las calificaciones de esta evaluación
		ratingRows, err := db.Query(`
			SELECT r.concept_id, r.value
			FROM ratings r
			WHERE r.evaluation_id = ?
		`, eval.ID)
		if err != nil {
			http.Error(w, "Error al obtener calificaciones", http.StatusInternalServerError)
			log.Printf("Error al consultar calificaciones: %v", err)
			return
		}

		ratings := []models.Rating{}
		for ratingRows.Next() {
			var rating models.Rating
			if err := ratingRows.Scan(&rating.ConceptID, &rating.Value); err != nil {
				http.Error(w, "Error al procesar calificaciones", http.StatusInternalServerError)
				log.Printf("Error al escanear calificación: %v", err)
				ratingRows.Close()
				return
			}
			ratings = append(ratings, rating)
		}
		ratingRows.Close()

		eval.Ratings = ratings
		evaluations = append(evaluations, eval)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evaluations)
}
