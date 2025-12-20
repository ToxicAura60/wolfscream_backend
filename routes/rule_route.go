package routes

import (
	"wolfscream/handlers"
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
)

func RuleRoutes() chi.Router {
	router := chi.NewRouter()

	router.With(middlewares.AuthMiddleware).Get("/", handlers.ListRules)
	router.With(middlewares.AuthMiddleware).Post("/", handlers.AddRule)

	return router

}