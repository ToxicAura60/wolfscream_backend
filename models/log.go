package models

import "time"

type Log struct {
	Id         string    `json:"id"`
	Text       string    `json:"text"`
	Level      string    `json:"level"`
	CreatedAt  time.Time `json:"created_at"`
	
}