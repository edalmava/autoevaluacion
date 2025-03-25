package sqlite

import (
	"database/sql"
	"os"
)

var DB *sql.DB

// InitDB inicializa la conexión a la base de datos
func InitDB(dbPath string) (*sql.DB, error) {
	// Verificar si el archivo existe
	_, err := os.Stat(dbPath)
	dbExists := !os.IsNotExist(err)

	// Abrir o crear la base de datos
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Si la base de datos no existía, inicializarla
	if !dbExists {
		err = initTables(db)
		if err != nil {
			return nil, err
		}
	}

	// Guardar la referencia global
	DB = db
	return db, nil
}

// GetDB retorna la instancia de la base de datos
func GetDB() *sql.DB {
	return DB
}

// initTables crea las tablas necesarias en la base de datos
func initTables(db *sql.DB) error {
	// Tabla de grados
	_, err := db.Exec(`
	CREATE TABLE grades (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	`)
	if err != nil {
		return err
	}

	// Tabla de estudiantes
	_, err = db.Exec(`
	CREATE TABLE students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		grade_id INTEGER,
		FOREIGN KEY (grade_id) REFERENCES grades(id)
	);
	`)
	if err != nil {
		return err
	}

	// Tabla de conceptos de evaluación
	_, err = db.Exec(`
	CREATE TABLE rating_concepts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);
	`)
	if err != nil {
		return err
	}

	// Tabla de evaluaciones
	_, err = db.Exec(`
	CREATE TABLE evaluations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER,
		date DATETIME,
		average REAL,
		FOREIGN KEY (student_id) REFERENCES students(id)
	);
	`)
	if err != nil {
		return err
	}

	// Tabla de calificaciones individuales
	_, err = db.Exec(`
	CREATE TABLE ratings (
		evaluation_id INTEGER,
		concept_id INTEGER,
		value INTEGER,
		PRIMARY KEY (evaluation_id, concept_id),
		FOREIGN KEY (evaluation_id) REFERENCES evaluations(id),
		FOREIGN KEY (concept_id) REFERENCES rating_concepts(id)
	);
	`)
	if err != nil {
		return err
	}

	// En la función initTables de db.go
	// Tabla de usuarios
	_, err = db.Exec(`
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		role TEXT NOT NULL
	);
	`)
	if err != nil {
		return err
	}

	// Insertar un usuario admin por defecto
	_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		"admin",
		"$2a$10$NFeFpXLNrCTNwz3pqjY1HOMRj4tK37GAOd76ZrmqdfD5nJNyvjlNC", // "admin2025" hasheado
		"admin")
	if err != nil {
		return err
	}

	// Insertar datos iniciales de grados
	gradesToInsert := []string{"6D", "6E", "7D", "7E", "8D", "8E", "9C", "9D", "11A", "11B", "11C"}
	for _, grade := range gradesToInsert {
		_, err = db.Exec("INSERT INTO grades (name) VALUES (?)", grade)
		if err != nil {
			return err
		}
	}

	// Insertar conceptos de evaluación
	concepts := []string{
		"Participación activa en clase",
		"Respeto a compañeros y profesores",
		"Puntualidad en la entrega de tareas",
		"Trabajo en equipo",
		"Organización y disciplina",
		"Asistencia regular a clases",
		"Actitud positiva hacia el aprendizaje",
		"Capacidad para seguir instrucciones",
		"Comportamiento durante actividades grupales",
		"Compromiso con su propio desarrollo académico",
	}

	for _, concept := range concepts {
		_, err = db.Exec("INSERT INTO rating_concepts (name) VALUES (?)", concept)
		if err != nil {
			return err
		}
	}

	// Insertar algunos estudiantes de ejemplo
	students := []struct {
		name  string
		grade string
	}{
		{"Ana García", "6D"},
		{"Juan Pérez", "6D"},
		{"María Rodríguez", "6E"},
		{"Carlos Sánchez", "7D"},
		{"Laura Martínez", "7E"},
		{"Pablo Fernández", "8D"},
	}

	for _, student := range students {
		// Obtener ID del grado
		var gradeID int
		err = db.QueryRow("SELECT id FROM grades WHERE name = ?", student.grade).Scan(&gradeID)
		if err != nil {
			return err
		}

		// Insertar estudiante
		_, err = db.Exec("INSERT INTO students (name, grade_id) VALUES (?, ?)", student.name, gradeID)
		if err != nil {
			return err
		}
	}

	return nil
}
