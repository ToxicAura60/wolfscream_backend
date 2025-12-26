package routes

import (
	"wolfscream/handlers"
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
)

func ScheduledMessageRoutes() chi.Router {
	router := chi.NewRouter()

	router.With(middlewares.AuthMiddleware).Post("/running/{scheduled-message-name}", handlers.EnableScheduledMessage)
	router.With(middlewares.AuthMiddleware).Delete("/running/{scheduled-message-name}", handlers.DisableScheduledMessage)
	router.With(middlewares.AuthMiddleware).Post("/", handlers.AddScheduledMessage)
	router.With(middlewares.AuthMiddleware).Get("/", handlers.ListScheduledMessages)
	router.With(middlewares.AuthMiddleware).Get("/{scheduled-message-name}", handlers.GetScheduledMessage)
	router.With(middlewares.AuthMiddleware).Get("/{scheduled-message-name}/log", handlers.FetchLogs)
	router.With(middlewares.AuthMiddleware).Get("/{scheduled-message-name}/state-history", handlers.FetchStateHistory)

	return router
	
}
