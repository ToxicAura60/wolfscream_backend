package models

import "time"

type ScheduledMessage struct {
  Id           int       `json:"id"`
  Name         string    `json:"name"`
  Message	     string    `json:"message"`
	Rule			   string    `json:"rule"`
	PlatformId   int       `json:"platform_id"`
	TableId      int       `json:"table_id"`
	Description  *string   `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	ScheduleType string    `json:"schedule_type"`
}

