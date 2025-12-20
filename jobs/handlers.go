package jobs

import (
	"fmt"
	"strings"
	"wolfscream/database"
	"wolfscream/discord"
	"wolfscream/models"
)

type JobFunc func(data map[string]any)

var JobHandlers = map[string]JobFunc{
	"discord": func(data map[string]any) {

		config := data["config"].(models.DiscordConfig)

		channelId := config.ChannelId
	
		messages := strings.Join(data["messages"].([]string), "\n\n")

		if _, err := discord.DiscordBot.ChannelMessageSend(channelId, messages); err != nil {
			logText := fmt.Sprintf("Failed to send message to channel %s: %v", channelId, err)
			database.DB.Exec("INSERT INTO logs (scheduled_message_id, text) VALUES ($1, $2);", config.ScheduledMessageId, logText)

			return
		}
	},
}
