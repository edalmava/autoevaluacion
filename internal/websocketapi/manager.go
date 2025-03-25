// Nuevas estructuras para WebSockets
package websocketapi

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/edalmava/student-behavior-api/internal/api/middleware"
	"github.com/edalmava/student-behavior-api/internal/db/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type WebSocketEvent struct {
	Type        string      `json:"type"`
	Data        interface{} `json:"data"`
	StudentID   int         `json:"student_id,omitempty"`
	GradeID     int         `json:"grade_id,omitempty"`
	StudentName string      `json:"student_name,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

type WebSocketManager struct {
	clients        map[*websocket.Conn]bool
	gradeClients   map[int]map[*websocket.Conn]bool
	studentClients map[int]map[*websocket.Conn]bool
	broadcast      chan WebSocketEvent
	register       chan *websocket.Conn
	unregister     chan *websocket.Conn
	mutex          sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Permitir cualquier origen para desarrollo, modificar para producción
	},
}

// Instancia global del WebSocketManager
var WsManager = WebSocketManager{
	clients:        make(map[*websocket.Conn]bool),
	gradeClients:   make(map[int]map[*websocket.Conn]bool),
	studentClients: make(map[int]map[*websocket.Conn]bool),
	broadcast:      make(chan WebSocketEvent),
	register:       make(chan *websocket.Conn),
	unregister:     make(chan *websocket.Conn),
}

// Implementación del gestor de WebSockets
func (manager *WebSocketManager) Run() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()

		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
			}

			// Eliminar cliente de las suscripciones por grado
			for gradeID, clients := range manager.gradeClients {
				if _, ok := clients[client]; ok {
					delete(manager.gradeClients[gradeID], client)
					if len(manager.gradeClients[gradeID]) == 0 {
						delete(manager.gradeClients, gradeID)
					}
				}
			}

			// Eliminar cliente de las suscripciones por estudiante
			for studentID, clients := range manager.studentClients {
				if _, ok := clients[client]; ok {
					delete(manager.studentClients[studentID], client)
					if len(manager.studentClients[studentID]) == 0 {
						delete(manager.studentClients, studentID)
					}
				}
			}
			manager.mutex.Unlock()

		case event := <-manager.broadcast:
			manager.mutex.Lock()
			// Enviar a todos los clientes generales
			for client := range manager.clients {
				err := client.WriteJSON(event)
				if err != nil {
					log.Printf("Error al enviar mensaje: %v", err)
					client.Close()
					delete(manager.clients, client)
				}
			}

			// Enviar a clientes suscritos al grado específico
			if event.GradeID > 0 {
				if clients, ok := manager.gradeClients[event.GradeID]; ok {
					for client := range clients {
						err := client.WriteJSON(event)
						if err != nil {
							log.Printf("Error al enviar mensaje a cliente de grado: %v", err)
							client.Close()
							delete(clients, client)
							if len(clients) == 0 {
								delete(manager.gradeClients, event.GradeID)
							}
						}
					}
				}
			}

			// Enviar a clientes suscritos al estudiante específico
			if event.StudentID > 0 {
				if clients, ok := manager.studentClients[event.StudentID]; ok {
					for client := range clients {
						err := client.WriteJSON(event)
						if err != nil {
							log.Printf("Error al enviar mensaje a cliente de estudiante: %v", err)
							client.Close()
							delete(clients, client)
							if len(clients) == 0 {
								delete(manager.studentClients, event.StudentID)
							}
						}
					}
				}
			}
			manager.mutex.Unlock()
		}
	}
}

// Suscribir un cliente a un grado específico
func (manager *WebSocketManager) subscribeToGrade(conn *websocket.Conn, gradeID int) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if _, exists := manager.gradeClients[gradeID]; !exists {
		manager.gradeClients[gradeID] = make(map[*websocket.Conn]bool)
	}
	manager.gradeClients[gradeID][conn] = true
}

// Suscribir un cliente a un estudiante específico
func (manager *WebSocketManager) subscribeToStudent(conn *websocket.Conn, studentID int) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if _, exists := manager.studentClients[studentID]; !exists {
		manager.studentClients[studentID] = make(map[*websocket.Conn]bool)
	}
	manager.studentClients[studentID][conn] = true
}

// Handler para WebSocket general
func WsHandler(w http.ResponseWriter, r *http.Request) {
	// Extraer token del parámetro de consulta
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Se requiere token", http.StatusUnauthorized)
		return
	}

	// Validar el token
	_, err := middleware.ParseToken(tokenString)
	if err != nil {
		http.Error(w, "Token inválido", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error al actualizar la conexión a WebSocket: %v", err)
		return
	}

	// Registrar nuevo cliente
	WsManager.register <- conn

	// Rutina de lectura (para mantener la conexión abierta)
	go func() {
		defer func() {
			WsManager.unregister <- conn
		}()

		for {
			// Leer mensajes del cliente (no hacemos nada con ellos por ahora)
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Error en WebSocket: %v", err)
				}
				break
			}
		}
	}()
}

// Handler para WebSocket por grado
func WsGradeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gradeID, err := strconv.Atoi(vars["gradeId"])
	if err != nil {
		http.Error(w, "ID de grado inválido", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error al actualizar la conexión a WebSocket: %v", err)
		return
	}

	// Registrar cliente general
	WsManager.register <- conn

	// Suscribir al grado específico
	WsManager.subscribeToGrade(conn, gradeID)

	// Rutina de lectura
	go func() {
		defer func() {
			WsManager.unregister <- conn
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Error en WebSocket: %v", err)
				}
				break
			}
		}
	}()
}

// Handler para WebSocket por estudiante
func WsStudentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	studentID, err := strconv.Atoi(vars["studentId"])
	if err != nil {
		http.Error(w, "ID de estudiante inválido", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error al actualizar la conexión a WebSocket: %v", err)
		return
	}

	// Registrar cliente general
	WsManager.register <- conn

	// Suscribir al estudiante específico
	WsManager.subscribeToStudent(conn, studentID)

	// Rutina de lectura
	go func() {
		defer func() {
			WsManager.unregister <- conn
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Error en WebSocket: %v", err)
				}
				break
			}
		}
	}()
}

// Añadir un método para notificar cambios en estudiantes
func notifyStudentChange(student models.Student, changeType string) {
	event := WebSocketEvent{
		Type:        changeType,
		Data:        student,
		StudentID:   student.ID,
		GradeID:     student.GradeID,
		StudentName: student.Name,
		Timestamp:   time.Now(),
	}
	WsManager.broadcast <- event
}

func NotifyEvaluationAdd(evaluationID int, studentID int, evalDate time.Time, average float64, gradeID int, studentName string) {
	// Notificar a los clientes WebSocket
	evaluation := models.Evaluation{
		ID:        evaluationID,
		StudentID: studentID,
		Date:      evalDate,
		Average:   average,
		// Las calificaciones detalladas no se incluyen en la notificación para simplificar
	}

	// Crear el evento WebSocket
	event := WebSocketEvent{
		Type:        "new_evaluation",
		Data:        evaluation,
		StudentID:   studentID,
		GradeID:     gradeID,
		StudentName: studentName,
		Timestamp:   time.Now(),
	}

	// Enviar el evento a todos los clientes interesados
	WsManager.broadcast <- event
}
