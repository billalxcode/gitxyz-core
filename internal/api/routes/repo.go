package routes

import (
	"gitxyz/internal/api/controllers"
	"gitxyz/internal/api/middlewares"
	"gitxyz/internal/models"
)

func (r *RoutesImpl) RegisterRepositories() {
	controller := controllers.NewRepoController(r.db)

	// /api/repos — repository collection.
	repos := r.engine.Group("/api/repos")
	repos.Use(middlewares.AuthRequired())
	repos.POST("", controller.Create)

	// /api/repos/:owner/:reponame — single repository + collaborators + policies.
	repo := r.engine.Group("/api/repos/:owner/:reponame")
	repo.Use(middlewares.AuthRequired())

	// Read requires repo:read scope (or admin/owner role via RequireScope).
	repo.GET("", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.Get)
	// Write requires repo:write scope (or admin/owner role via RequireScope).
	repo.PATCH("", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.Update)
	repo.DELETE("", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.Delete)

	// Collaborator management — repo:write scope.
	repo.GET("/collaborators", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.ListCollaborators)
	repo.POST("/collaborators", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.AddCollaborator)
	repo.PUT("/collaborators/:username", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.UpdateCollaborator)
	repo.DELETE("/collaborators/:username", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.RemoveCollaborator)

	// Policy (ABAC) management — repo:write scope.
	repo.GET("/policies", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.ListPolicies)
	repo.POST("/policies", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.AddPolicy)
	repo.DELETE("/policies/:id", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.RemovePolicy)
}
