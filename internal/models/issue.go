package models

import (
	"time"
)

// Issue state constants.
const (
	IssueStateOpen   = "open"
	IssueStateClosed = "closed"
)

// ValidIssueState reports whether state is a recognized issue state.
func ValidIssueState(state string) bool {
	return state == IssueStateOpen || state == IssueStateClosed
}

// Issue represents a tracked issue/bug/feature request in a repository.
type Issue struct {
	Base

	Number     int        `json:"number" gorm:"not null;uniqueIndex:uniq_repo_issue"`
	RepoID     string     `json:"repo_id" gorm:"type:uuid;not null;uniqueIndex:uniq_repo_issue"`
	Title      string     `json:"title" gorm:"size:255;not null"`
	Body       string     `json:"body" gorm:"type:text"`
	State      string     `json:"state" gorm:"size:10;not null;default:'open'"`
	AuthorID   string     `json:"author_id" gorm:"type:uuid;not null"`
	AssigneeID *string    `json:"assignee_id" gorm:"type:uuid"`
	Labels     []Label    `json:"labels" gorm:"many2many:issue_labels;"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ClosedAt   *time.Time `json:"closed_at"`

	Author   *User `json:"author" gorm:"foreignKey:AuthorID;references:ID"`
	Assignee *User `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID;references:ID"`
}

func (Issue) TableName() string { return "issues" }

// Label is a tag that can be attached to issues.
type Label struct {
	Base

	RepoID      string `json:"repo_id" gorm:"type:uuid;not null"`
	Name        string `json:"name" gorm:"size:50;not null"`
	Color       string `json:"color" gorm:"size:7;not null;default:'#cccccc'"`
	Description string `json:"description" gorm:"type:text"`
}

func (Label) TableName() string { return "labels" }

// IssueComment is a comment on an issue.
type IssueComment struct {
	Base

	IssueID  string `json:"issue_id" gorm:"type:uuid;not null"`
	AuthorID string `json:"author_id" gorm:"type:uuid;not null"`
	Body     string `json:"body" gorm:"type:text;not null"`

	Author *User `json:"author" gorm:"foreignKey:AuthorID;references:ID"`
}

func (IssueComment) TableName() string { return "issue_comments" }

// IssueAssignee is an explicit many-to-many assignment row (used by the
// POST/DELETE /issues/:number/assignees endpoints).
type IssueAssignee struct {
	IssueID   string    `json:"issue_id" gorm:"type:uuid;not null;uniqueIndex:uniq_issue_assignee"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:uniq_issue_assignee"`
	CreatedAt time.Time `json:"created_at"`
}

func (IssueAssignee) TableName() string { return "issue_assignees" }
