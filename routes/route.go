package routes

import (
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middlewares.CORS())
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(router chi.Router) {
		router.Mount("/table", SchemaRoutes())
		router.Mount("/message-template", TemplateRoutes())
		router.Mount("/scheduled-message", ScheduledMessageRoutes())
		router.Mount("/discord", DiscordRoutes())
		router.Mount("/rule", RuleRoutes())
		router.Mount("/platform", PlatformRoutes())
	})

	return r
}
