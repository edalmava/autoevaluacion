// internal/db/models/grade.go
package models

type Grade struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}
