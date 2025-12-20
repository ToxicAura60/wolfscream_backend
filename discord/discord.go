package discord

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

var DiscordBot *discordgo.Session

func init() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalf("DISCORD_BOT_TOKEN environment variable is not set")
	}

	var err error
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	err = bot.Open()
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}

	DiscordBot = bot
}
