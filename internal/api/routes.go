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

	// Rutas públicas
	r.HandleFunc("/login", handlers.LoginHandler).Methods(http.MethodPost)
	r.HandleFunc("/login", handlers.LoginPageHandler)

	// Rutas estáticas
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	// Rutas de páginas
	r.HandleFunc("/", handlers.LoginPageHandler)
	r.HandleFunc("/admin", handlers.AdminHandler)
	r.HandleFunc("/grades", handlers.GradesHandler)
	r.HandleFunc("/dashboard", handlers.DashboardHandler)
	r.HandleFunc("/evaluacion", handlers.IndexHandler).Methods(http.MethodGet)
	//r.HandleFunc("/admin/student/{studentId}/grades", handlers.AdminStudentGradesHandler)

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.Auth) // Middleware de autenticación para todas las rutas de la API

	// Rutas para usuarios
	api.HandleFunc("/users/register", handlers.RegisterHandler).Methods(http.MethodPost)
	api.HandleFunc("/users/me", handlers.GetMeHandler).Methods(http.MethodGet)

	// Rutas para administradores
	adminApi := api.PathPrefix("/admin").Subrouter()
	adminApi.Use(middleware.RequireRole("admin", "teacher"))
	adminApi.HandleFunc("/gradesAdmin", handlers.GetGradesHandler).Methods(http.MethodGet)
	adminApi.HandleFunc("/grades/{id}", handlers.GetGradeHandler).Methods(http.MethodGet)
	adminApi.HandleFunc("/grades", handlers.CreateGrade).Methods(http.MethodPost)
	adminApi.HandleFunc("/grades/{id}", handlers.UpdateGrade).Methods(http.MethodPut)
	adminApi.HandleFunc("/grades/{id}", handlers.DeleteGrade).Methods(http.MethodDelete)
	adminApi.HandleFunc("/grades/{id}/toggle", handlers.ToggleGrade).Methods(http.MethodPut)

	// Gestión de usuarios (admin)
	adminApi.HandleFunc("/users", handlers.GetUsersHandler).Methods(http.MethodGet)
	adminApi.HandleFunc("/users", handlers.RegisterHandler).Methods(http.MethodPost)
	adminApi.HandleFunc("/users/{id}", handlers.GetUserHandler).Methods(http.MethodGet)
	adminApi.HandleFunc("/users/{id}", handlers.UpdateUserHandler).Methods(http.MethodPut)
	adminApi.HandleFunc("/users/{id}", handlers.DeleteUserHandler).Methods(http.MethodDelete)
	adminApi.HandleFunc("/users/{id}/toggle", handlers.ToggleUserHandler).Methods(http.MethodPut)

	// Rutas para profesores y administradores
	teacherApi := api.PathPrefix("").Subrouter()
	teacherApi.Use(middleware.RequireRole("admin", "teacher", "student"))
	teacherApi.HandleFunc("/grades", handlers.GetGrades).Methods(http.MethodGet)
	teacherApi.HandleFunc("/students", handlers.GetStudentsHandler).Methods(http.MethodGet).Queries("grade", "{grade}")
	teacherApi.HandleFunc("/evaluation", handlers.SaveEvaluationHandler).Methods(http.MethodPost)
	teacherApi.HandleFunc("/evaluations/{studentId}", handlers.GetStudentEvaluationsHandler).Methods(http.MethodGet)

	// Nueva ruta para WebSocket
	wsRouter := r.PathPrefix("/ws").Subrouter()
	//wsRouter.Use(middleware.Auth)
	wsRouter.HandleFunc("", websocketapi.WsHandler).Methods(http.MethodGet)
	wsRouter.HandleFunc("/grade/{gradeId}", websocketapi.WsGradeHandler).Methods(http.MethodGet)
	wsRouter.HandleFunc("/student/{studentId}", websocketapi.WsStudentHandler).Methods(http.MethodGet)
}
