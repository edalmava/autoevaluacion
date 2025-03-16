package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/edalmava/student-behavior-api/internal/api"
	"github.com/edalmava/student-behavior-api/internal/db/sqlite"
	"github.com/edalmava/student-behavior-api/internal/websocketapi"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Configurar la base de datos
	dbPath := "./data/student_evaluations.db"
	db, err := sqlite.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Error al inicializar la base de datos: %v", err)
	}
	defer db.Close()

	// Configurar router
	r := mux.NewRouter()

	api.SetupRoutes(r)

	// Iniciar el gestor de WebSockets
	go websocketapi.WsManager.Run()

	// Nueva ruta para WebSocket
	r.HandleFunc("/ws", websocketapi.WsHandler).Methods("GET")
	r.HandleFunc("/ws/grade/{gradeId}", websocketapi.WsGradeHandler).Methods("GET")
	r.HandleFunc("/ws/student/{studentId}", websocketapi.WsStudentHandler).Methods("GET")

	// Iniciar servidor
	port := 8080
	fmt.Printf("Servidor iniciado en http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
