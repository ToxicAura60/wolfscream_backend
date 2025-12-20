package routes

import (
	"wolfscream/handlers"
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
)

func TemplateRoutes() chi.Router {
	router := chi.NewRouter()

	router.With(middlewares.AuthMiddleware).Post("/", handlers.AddTemplate)
	router.With(middlewares.AuthMiddleware).Get("/", handlers.ListMessageTemplates)
	router.With(middlewares.AuthMiddleware).Delete("/{messageTemplateId}", handlers.DeleteMessageTemplate)

	return router
	
}
