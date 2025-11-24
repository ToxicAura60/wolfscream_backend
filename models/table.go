package models

import "time"

type Table struct {
	Name        string    `json:"name"`
	Description *string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
