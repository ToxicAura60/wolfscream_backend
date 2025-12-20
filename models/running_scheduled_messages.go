package models

import "time"

type RunningScheduledMessages struct {
	Id                 int
	ScheduledMessageId int 
	CreatedAt          time.Time `json:"created_at"`
}