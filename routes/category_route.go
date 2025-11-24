package routes

import (
	"wolfscream/handlers"
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
)

func CategoryRoutes() chi.Router {
	router := chi.NewRouter()


	router.With(middlewares.AuthMiddleware).Get("/", handlers.ListCategories)
	router.With(middlewares.AuthMiddleware).Post("/", handlers.AddCategory)
	router.With(middlewares.AuthMiddleware).Put("/", handlers.UpdateCategory)
	router.With(middlewares.AuthMiddleware).Delete("/{categoryName}", handlers.DeleteCategory)

	
	router.With(middlewares.AuthMiddleware).Post("/{categoryName}", handlers.AddCategoryItem)
	

	return router
	
}
