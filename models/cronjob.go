package models

import "time"

type CronJob struct {
  Id int
  ScheduledMessageId int
  Minute *int
  Hour *int
  DayOfMonth *int
  Month *int
  DayOfWeek *int
  CreatedAt time.Time
}