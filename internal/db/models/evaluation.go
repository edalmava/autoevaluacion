package models

import "time"

type Evaluation struct {
	ID        int       `json:"id"`
	StudentID int       `json:"student_id"`
	Date      time.Time `json:"date"`
	Ratings   []Rating  `json:"ratings"`
	Average   float64   `json:"average"`
}

// Para recibir evaluaciones del frontend
type EvaluationRequest struct {
	StudentName string         `json:"studentName"`
	Grade       string         `json:"grade"`
	Ratings     map[string]int `json:"ratings"`
	Average     float64        `json:"average"`
	Date        string         `json:"date"`
}
