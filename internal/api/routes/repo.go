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

	// Issues — nested under the repository group.
	issueController := controllers.NewIssueController(r.db)

	// Issues collection + single issue.
	issues := repo.Group("/issues")
	issues.GET("", middlewares.RequireScope(r.db, models.ScopeRepoRead), issueController.List)
	issues.POST("", middlewares.RequireScope(r.db, models.ScopeRepoWrite), issueController.Create)
	issues.GET("/:number", middlewares.RequireScope(r.db, models.ScopeRepoRead), issueController.Get)
	issues.PATCH("/:number", middlewares.RequireScope(r.db, models.ScopeRepoWrite), issueController.Update)
	issues.DELETE("/:number", middlewares.RequireScope(r.db, models.ScopeRepoWrite), issueController.Delete)

	// Issue comments.
	issueComments := repo.Group("/issues/:number/comments")
	issueComments.GET("", middlewares.RequireScope(r.db, models.ScopeRepoRead), issueController.ListComments)
	issueComments.POST("", middlewares.RequireScope(r.db, models.ScopeRepoWrite), issueController.CreateComment)

	// Labels.
	labels := repo.Group("/labels")
	labels.GET("", middlewares.RequireScope(r.db, models.ScopeRepoRead), issueController.ListLabels)
	labels.POST("", middlewares.RequireScope(r.db, models.ScopeRepoWrite), issueController.CreateLabel)

	// Issue assignees.
	assignees := repo.Group("/issues/:number/assignees")
	assignees.GET("", middlewares.RequireScope(r.db, models.ScopeRepoRead), issueController.ListAssignees)
	assignees.POST("", middlewares.RequireScope(r.db, models.ScopeRepoWrite), issueController.AddAssignee)
	assignees.DELETE("/:username", middlewares.RequireScope(r.db, models.ScopeRepoWrite), issueController.RemoveAssignee)

	// Patch Requests — nested under the repository group.
	patchController := controllers.NewPatchController(r.db)

	// Patch collection + single patch.
	patches := repo.Group("/patches")
	patches.GET("", middlewares.RequireScope(r.db, models.ScopeRepoRead), patchController.List)
	patches.POST("", middlewares.CollaboratorOrOwner(r.db), patchController.Create)
	patches.GET("/:number", middlewares.RequireScope(r.db, models.ScopeRepoRead), patchController.Get)
	patches.PATCH("/:number", middlewares.CollaboratorOrOwner(r.db), patchController.Update)
	patches.POST("/:number/refresh", middlewares.CollaboratorOrOwner(r.db), patchController.Refresh)
	patches.POST("/:number/merge", middlewares.MaintainerOrOwner(r.db), patchController.Merge)

	// Patch snapshot views.
	patches.GET("/:number/commits", middlewares.RequireScope(r.db, models.ScopeRepoRead), patchController.ListCommits)
	patches.GET("/:number/files", middlewares.RequireScope(r.db, models.ScopeRepoRead), patchController.ListFiles)

	// Patch reviewers.
	patches.POST("/:number/reviewers", middlewares.CollaboratorOrOwner(r.db), patchController.AssignReviewer)
	patches.DELETE("/:number/reviewers/:username", middlewares.CollaboratorOrOwner(r.db), patchController.UnassignReviewer)
	patches.GET("/:number/reviewers", middlewares.RequireScope(r.db, models.ScopeRepoRead), patchController.ListReviewers)

	// Patch reviews.
	patches.GET("/:number/reviews", middlewares.RequireScope(r.db, models.ScopeRepoRead), patchController.ListReviews)
	patches.POST("/:number/reviews", middlewares.RequireScope(r.db, models.ScopeRepoRead), patchController.SubmitReview)

	// Patch comments (public to authenticated readers).
	patches.GET("/:number/comments", middlewares.RequireScope(r.db, models.ScopeRepoRead), patchController.ListComments)
	patches.POST("/:number/comments", middlewares.RequireScope(r.db, models.ScopeRepoRead), patchController.CreateComment)
}
