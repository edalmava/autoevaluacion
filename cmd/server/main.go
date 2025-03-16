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

	// Iniciar servidor
	port := 8080
	serverAddress := fmt.Sprintf(":%d", port)
	fmt.Printf("Servidor iniciado en http://localhost%s\n", serverAddress)
	if err := http.ListenAndServe(serverAddress, r); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
