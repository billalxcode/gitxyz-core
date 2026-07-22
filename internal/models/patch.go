package models

import (
	"time"
)

// Patch request state constants.
const (
	PatchStateOpen   = "open"
	PatchStateMerged = "merged"
	PatchStateClosed = "closed"
)

// ValidPatchState reports whether state is a recognized patch state.
func ValidPatchState(state string) bool {
	return state == PatchStateOpen || state == PatchStateMerged || state == PatchStateClosed
}

// PatchFile status constants.
const (
	PatchFileAdded    = "added"
	PatchFileModified = "modified"
	PatchFileDeleted  = "deleted"
	PatchFileRenamed  = "renamed"
)

// ValidPatchFileStatus reports whether status is a recognized file status.
func ValidPatchFileStatus(status string) bool {
	switch status {
	case PatchFileAdded, PatchFileModified, PatchFileDeleted, PatchFileRenamed:
		return true
	default:
		return false
	}
}

// PatchReview state constants.
const (
	PatchReviewApproved         = "approved"
	PatchReviewChangesRequested = "changes_requested"
	PatchReviewCommented        = "commented"
)

// ValidPatchReviewState reports whether state is a recognized review state.
func ValidPatchReviewState(state string) bool {
	switch state {
	case PatchReviewApproved, PatchReviewChangesRequested, PatchReviewCommented:
		return true
	default:
		return false
	}
}

// PatchRequest is a request to merge source_branch into target_branch.
type PatchRequest struct {
	Base

	RepoID         string     `json:"repo_id" gorm:"type:uuid;not null;uniqueIndex:uniq_repo_patch"`
	Number         int        `json:"number" gorm:"not null;uniqueIndex:uniq_repo_patch"`
	Title          string     `json:"title" gorm:"size:255;not null"`
	Body           string     `json:"body" gorm:"type:text"`
	SourceBranch   string     `json:"source_branch" gorm:"size:255;not null"`
	TargetBranch   string     `json:"target_branch" gorm:"size:255;not null"`
	AuthorID       string     `json:"author_id" gorm:"type:uuid;not null"`
	State          string     `json:"state" gorm:"size:10;not null;default:'open'"`
	BaseSHA        string     `json:"base_sha" gorm:"size:40"`
	HeadSHA        string     `json:"head_sha" gorm:"size:40"`
	MergeCommitSHA *string    `json:"merge_commit_sha" gorm:"size:40"`
	IsMergeable    *bool      `json:"is_mergeable"`
	MergedAt       *time.Time `json:"merged_at"`
	ClosedAt       *time.Time `json:"closed_at"`

	Author    *User           `json:"author" gorm:"foreignKey:AuthorID;references:ID"`
	Commits   []PatchCommit   `json:"commits" gorm:"foreignKey:PatchID;references:ID"`
	Files     []PatchFile     `json:"files" gorm:"foreignKey:PatchID;references:ID"`
	Reviewers []PatchReviewer `json:"reviewers" gorm:"foreignKey:PatchID;references:ID"`
	Reviews   []PatchReview   `json:"reviews" gorm:"foreignKey:PatchID;references:ID"`
	Comments  []PatchComment  `json:"comments" gorm:"foreignKey:PatchID;references:ID"`
}

func (PatchRequest) TableName() string { return "patch_requests" }

// PatchCommit is a snapshot of a single commit in the patch.
type PatchCommit struct {
	Base

	PatchID     string    `json:"patch_id" gorm:"type:uuid;not null"`
	SHA         string    `json:"sha" gorm:"size:40;not null"`
	Message     string    `json:"message" gorm:"type:text"`
	AuthorName  string    `json:"author_name" gorm:"size:255"`
	AuthorEmail string    `json:"author_email" gorm:"size:255"`
	AuthorDate  time.Time `json:"author_date"`
}

func (PatchCommit) TableName() string { return "patch_commits" }

// PatchFile is a snapshot of a single changed file in the patch.
type PatchFile struct {
	Base

	PatchID  string `json:"patch_id" gorm:"type:uuid;not null"`
	FilePath string `json:"file_path" gorm:"type:text;not null"`
	Status   string `json:"status" gorm:"size:20;not null"`
	Diff     string `json:"diff" gorm:"type:text"`
}

func (PatchFile) TableName() string { return "patch_files" }

// PatchReviewer is an explicit assignment of a reviewer to a patch.
type PatchReviewer struct {
	Base

	PatchID string `json:"patch_id" gorm:"type:uuid;not null;uniqueIndex:uniq_patch_reviewer"`
	UserID  string `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:uniq_patch_reviewer"`

	User *User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

func (PatchReviewer) TableName() string { return "patch_reviewers" }

// PatchReview is a review submission by an assigned reviewer.
type PatchReview struct {
	Base

	PatchID  string `json:"patch_id" gorm:"type:uuid;not null"`
	AuthorID string `json:"author_id" gorm:"type:uuid;not null"`
	State    string `json:"state" gorm:"size:20;not null"`
	Body     string `json:"body" gorm:"type:text"`

	Author *User `json:"author" gorm:"foreignKey:AuthorID;references:ID"`
}

func (PatchReview) TableName() string { return "patch_reviews" }

// PatchComment is a comment on a patch (can be inline on a file/line).
type PatchComment struct {
	Base

	PatchID  string  `json:"patch_id" gorm:"type:uuid;not null"`
	AuthorID string  `json:"author_id" gorm:"type:uuid;not null"`
	Body     string  `json:"body" gorm:"type:text;not null"`
	FilePath *string `json:"file_path" gorm:"type:text"`
	Line     *int    `json:"line"`

	Author *User `json:"author" gorm:"foreignKey:AuthorID;references:ID"`
}

func (PatchComment) TableName() string { return "patch_comments" }
