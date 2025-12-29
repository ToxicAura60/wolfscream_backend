package routes

import (
	"wolfscream/handlers"
	"wolfscream/middlewares"

	"github.com/go-chi/chi/v5"
)

func SchemaRoutes() chi.Router {
	router := chi.NewRouter()

	router.With(middlewares.AuthMiddleware).Post("/", handlers.CreateTable)
	router.With(middlewares.AuthMiddleware).Get("/", handlers.ListTables)
	router.With(middlewares.AuthMiddleware).Put("/{table-name}", handlers.UpdateTable)
	router.With(middlewares.AuthMiddleware).Delete("/{table-name}", handlers.DropTable)

	router.With(middlewares.AuthMiddleware).Post("/{table-name}/column", handlers.AddColumn)
	router.With(middlewares.AuthMiddleware).Get("/{table-name}/column", handlers.ListColumns)
	router.With(middlewares.AuthMiddleware).Delete("/{table-name}/column/{column}", handlers.DeleteColumn)
	router.With(middlewares.AuthMiddleware).Put("/{table-name}/column/{column}", handlers.UpdateColumn)

	router.With(middlewares.AuthMiddleware).Get("/{table-name}/data", handlers.GetData)
	router.With(middlewares.AuthMiddleware).Post("/{table-name}/data", handlers.InsertData)
	router.With(middlewares.AuthMiddleware).Delete("/{table-name}/data/{columnId}", handlers.DeleteData)

	return router

}
