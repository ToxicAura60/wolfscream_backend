package models

type Column struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Length       *int    `json:"length"`
	DefaultValue any    `json:"default"`
}