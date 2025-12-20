package routes

import (
	"wolfscream/handlers"
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
)

func ScheduledMessageRoutes() chi.Router {
	router := chi.NewRouter()

	router.With(middlewares.AuthMiddleware).Post("/enable", handlers.EnableScheduledMessage)
	router.With(middlewares.AuthMiddleware).Post("/disable", handlers.DisableScheduledMessage)
	router.With(middlewares.AuthMiddleware).Post("/", handlers.AddScheduledMessage)
	router.With(middlewares.AuthMiddleware).Get("/", handlers.ListScheduledMessages)
	router.With(middlewares.AuthMiddleware).Get("/{id}/log", handlers.FetchLogs)

	return router
	
}
