package models

type Student struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	GradeID int    `json:"grade_id"`
}

// Para la respuesta de estudiantes por grado
type StudentsResponse struct {
	Students []Student `json:"students"`
}
