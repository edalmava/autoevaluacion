// internal/api/handlers/pages.go
package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Configuración para las páginas
var (
	TemplatesDir = "./web/templates"
	StaticDir    = "./web/static"
)

// PageHandler es un generador de handlers para servir páginas HTML
func PageHandler(templateName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templatePath := filepath.Join(TemplatesDir, templateName)

		// Verificar si el archivo existe antes de servirlo
		_, err := os.Stat(templatePath)
		if os.IsNotExist(err) {
			log.Printf("Archivo de plantilla no encontrado: %s", templatePath)
			http.Error(w, "Página no encontrada", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Error al verificar archivo %s: %v", templatePath, err)
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
			return
		}

		// Servir el archivo
		http.ServeFile(w, r, templatePath)
	}
}

// IndexHandler sirve la página principal
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	PageHandler("index.html")(w, r)
}

// AdminHandler sirve la página de administración
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	PageHandler("admin.html")(w, r)
}

// GradesHandler sirve la página de grados
func GradesHandler(w http.ResponseWriter, r *http.Request) {
	PageHandler("grades.html")(w, r)
}

// LoginPageHandler maneja la ruta de la página de login
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// Servir la página de login
	http.ServeFile(w, r, "./web/static/login.html")
}

// LoginPageHandler maneja la ruta de la página de login
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	PageHandler("dashboard.html")(w, r)
}
