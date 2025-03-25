// internal/db/models/user.go
package models

type User struct {
	/*ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"` // No mostrar en respuestas JSON
	Role     string `json:"role"`     // "admin" o "teacher"*/
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"` // No mostrar en respuestas JSON
	Role     string `json:"role"`               // "admin", "teacher" o "student"
	Active   bool   `json:"active"`             // Estado del usuario (activo/inactivo)
}

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"` // "admin" o "teacher"
}

// UserUpdate contiene los campos que se pueden actualizar
type UserUpdate struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"`
	Active   bool   `json:"active,omitempty"`
	GradeID  int    `json:"gradeId,omitempty"`
}
