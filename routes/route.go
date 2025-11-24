package routes

import (
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var Router *chi.Mux

func init() {
	Router = chi.NewRouter()

	Router.Use(middleware.Logger)
	Router.Use(middlewares.CORS())
	Router.Use(middleware.Recoverer)

	Router.Route("/api/v1", func(router chi.Router) {
		router.Mount("/table", SchemaRoutes())
		router.Mount("/template", TemplateRoutes())
		router.Mount("/category", CategoryRoutes())
	})
}
