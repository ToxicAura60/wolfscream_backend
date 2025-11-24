package models

import "time"

type Severity struct {
	Id					int    `json:"id"`
	Name        string `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}