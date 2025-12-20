package routes

import (
	"wolfscream/handlers"
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
)

func DiscordRoutes() chi.Router {
	router := chi.NewRouter()


	router.With(middlewares.AuthMiddleware).Get("/guild/{guildId}/channel", handlers.ListDiscordChannels)
	router.With(middlewares.AuthMiddleware).Get("/guild", handlers.ListDiscordGuilds)


	return router
	
}
