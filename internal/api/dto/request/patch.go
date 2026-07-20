package dto

type CreatePatchRequest struct {
	Title        string `json:"title" binding:"required,max=255"`
	Body         string `json:"body"`
	SourceBranch string `json:"source_branch" binding:"required"`
	TargetBranch string `json:"target_branch" binding:"required"`
}

type UpdatePatchRequest struct {
	Title *string `json:"title" binding:"omitempty,max=255"`
	Body  *string `json:"body"`
	State *string `json:"state" binding:"omitempty,oneof=open merged closed"`
}

type SubmitPatchReviewRequest struct {
	State string `json:"state" binding:"required,oneof=approved changes_requested commented"`
	Body  string `json:"body"`
}

type CreatePatchCommentRequest struct {
	Body     string `json:"body" binding:"required"`
	FilePath string `json:"file_path"`
	Line     *int   `json:"line"`
}

type AssignPatchReviewerRequest struct {
	Username string `json:"username" binding:"required"`
}
