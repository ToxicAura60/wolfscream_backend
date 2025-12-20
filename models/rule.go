package models

import "time"

type Rule struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"created_at"`

	
}

