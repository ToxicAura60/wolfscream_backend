package models

import "time"

type Interval struct {
	Id                 string
	ScheduledMessageId int
	Value      int
	Unit       string
	CreatedAt  time.Time
}