package models

import "time"

type Platform struct {
    Id                int       `json:"id"`
    Name              string    `json:"name"`
    ImageUrl          *string   `json:"image_url"`
    CreatedAt         time.Time `json:"created_at"`
}
