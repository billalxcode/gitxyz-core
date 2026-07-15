package routes

import (
	"gitxyz/internal/api/controllers"
	"gitxyz/internal/api/middlewares"
)

func (r *RoutesImpl) RegisterRepositories() {
	controller := controllers.NewRepoController(r.db)

	routes := r.engine.Group("/api/repos")
	protectedRoutes := routes.Group("/")
	protectedRoutes.Use(middlewares.AuthRequired())

	protectedRoutes.POST("", controller.Create)
}
