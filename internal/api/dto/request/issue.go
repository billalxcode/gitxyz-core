package dto

type CreateIssueRequest struct {
	Title  string   `json:"title" binding:"required,max=255"`
	Body   string   `json:"body"`
	Labels []string `json:"labels"`
}

type UpdateIssueRequest struct {
	Title  *string  `json:"title" binding:"omitempty,max=255"`
	Body   *string  `json:"body"`
	State  *string  `json:"state" binding:"omitempty,oneof=open closed"`
	Labels []string `json:"labels"`
}

type CreateIssueCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type CreateLabelRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Color       string `json:"color" binding:"omitempty,len=7"`
	Description string `json:"description"`
}

type AssignIssueRequest struct {
	Username string `json:"username" binding:"required"`
}
