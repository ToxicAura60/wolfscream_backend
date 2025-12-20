package routes

import (
	"wolfscream/handlers"
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
)

func PlatformRoutes() chi.Router {
	router := chi.NewRouter()


	router.With(middlewares.AuthMiddleware).Get("/", handlers.ListPlatforms)

	

	return router
	
}
