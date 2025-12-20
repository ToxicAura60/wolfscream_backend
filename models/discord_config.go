package models

import "time"

type DiscordConfig struct {
	Id        int
  ChannelId string
  ScheduledMessageId int
  CreatedAt time.Time 
}