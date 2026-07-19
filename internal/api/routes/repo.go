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

	// /api/users/:username/repos — list a user's repositories.
	userRepos := r.engine.Group("/api/users/:username/repos")
	userRepos.Use(middlewares.AuthRequired())
	userRepos.GET("", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.List)

	// /api/repos/:owner/:reponame — single repository + collaborators + policies.
	repo := r.engine.Group("/api/repos/:owner/:reponame")
	repo.Use(middlewares.AuthRequired())

	// Read requires repo:read scope (or admin/owner role via RequireScope).
	repo.GET("", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.Get)
	// Write requires repo:write scope (or admin/owner role via RequireScope).
	repo.PATCH("", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.Update)
	repo.DELETE("", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.Delete)

	// Branches.
	repo.GET("/branches", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.ListBranches)
	repo.DELETE("/branches/:branch", middlewares.RequireScope(r.db, models.ScopeRepoWrite), controller.DeleteBranch)

	// Commits.
	repo.GET("/commits", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.ListCommits)
	repo.GET("/commits/:sha", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.GetCommit)

	// Contents / file browsing.
	repo.GET("/contents/*path", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.GetContents)
	repo.GET("/raw/*path", middlewares.RequireScope(r.db, models.ScopeRepoRead), controller.GetFile)

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
