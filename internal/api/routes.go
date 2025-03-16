// internal/api/routes.go
package api

import (
	"net/http"

	"github.com/edalmava/student-behavior-api/internal/api/handlers"
	"github.com/edalmava/student-behavior-api/internal/api/middleware"
	"github.com/edalmava/student-behavior-api/internal/websocketapi"

	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router) {
	// Aplicar middleware a todas las rutas
	r.Use(middleware.CORS)

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/gradesAdmin", handlers.GetGradesHandler).Methods(http.MethodGet)
	api.HandleFunc("/grades", handlers.GetGrades).Methods(http.MethodGet)
	api.HandleFunc("/grades", handlers.CreateGrade).Methods(http.MethodPost)
	api.HandleFunc("/grades/{id}", handlers.UpdateGrade).Methods("PUT")
	api.HandleFunc("/grades/{id}", handlers.DeleteGrade).Methods("DELETE")
	api.HandleFunc("/grades/{id}/toggle", handlers.ToggleGrade).Methods("PUT")

	// Endpoints para los datos
	api.HandleFunc("/students", handlers.GetStudentsHandler).Methods("GET").Queries("grade", "{grade}")
	api.HandleFunc("/evaluation", handlers.SaveEvaluationHandler).Methods("POST")
	api.HandleFunc("/evaluations/{studentId}", handlers.GetStudentEvaluationsHandler).Methods("GET")

	// Rutas estáticas
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	// Rutas de páginas
	r.HandleFunc("/", handlers.IndexHandler)
	r.HandleFunc("/admin", handlers.AdminHandler)
	r.HandleFunc("/grades", handlers.GradesHandler)

	// Nueva ruta para WebSocket
	r.HandleFunc("/ws", websocketapi.WsHandler).Methods("GET")
	r.HandleFunc("/ws/grade/{gradeId}", websocketapi.WsGradeHandler).Methods("GET")
	r.HandleFunc("/ws/student/{studentId}", websocketapi.WsStudentHandler).Methods("GET")
}
