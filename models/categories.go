package models

import "time"

type Category struct {
	Id 					int				
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
