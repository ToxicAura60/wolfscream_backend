package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"wolfscream/discord"
	"wolfscream/models"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
)

// --------------------
// List Discord Guilds
// --------------------
func ListDiscordGuilds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	guilds, err := discord.DiscordBot.UserGuilds(100, "", "", false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to get guilds: %v", err),
		})
		return
	}

	var data []models.Guild

	for _, guild := range guilds {
		var imageUrl *string = nil

		if guild.Icon != "" {
			url := ""

			if strings.HasPrefix(guild.Icon, "a_") {
				url = fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.gif", guild.ID, guild.Icon)
			} else {
				url = fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.png", guild.ID, guild.Icon)
			}

			imageUrl = &url
		}

		data = append(data,
			models.Guild{
        Id:       guild.ID,
        Name:     guild.Name,
        ImageUrl: imageUrl,
    	},
		)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"data": data,
	})
}
// --------------------
// List Discord Guilds End
// --------------------




// --------------------
// List Discord Channels
// --------------------
func ListDiscordChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	guildID := chi.URLParam(r, "guildId")

	channels, err := discord.DiscordBot.GuildChannels(guildID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": fmt.Sprintf("Failed to get channels: %v", err),
		})
		return
	}

	var result []models.DiscordChannel

	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildText {
			result = append(result, models.DiscordChannel{
				Id:   channel.ID,
				Name: channel.Name,
			})
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"data": result,
	})
}
// --------------------
// List Discord Channels End
// --------------------


