package discord

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

var DiscordBot *discordgo.Session

func init() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalf("DISCORD_BOT_TOKEN environment variable is not set")
		return
	}

	var err error
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
		return
	}

	err = bot.Open()
	if err != nil {
		fmt.Println("Gagal membuka connection:", err)
		return
	}

	DiscordBot = bot
}
