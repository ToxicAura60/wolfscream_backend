package models

import "time"

type Schedule struct {
	Id           int
	Name         string
	ScheduleType string
  Created_at   time.Time	
}