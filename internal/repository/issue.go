package repository

import (
	"gitxyz/internal/models"

	"gorm.io/gorm"
)

// IssueRepository provides data access for issues, labels, comments and assignees.
type IssueRepository interface {
	// Issues
	Create(issue *models.Issue) error
	FindByRepo(repoID string, dest *[]models.Issue) error
	FindByNumber(repoID string, number int) (models.Issue, error)
	FindByID(id string) (models.Issue, error)
	Update(issue *models.Issue) error
	Delete(issue *models.Issue) error
	NextNumber(repoID string) (int, error)

	// Labels
	CreateLabel(label *models.Label) error
	FindLabels(repoID string, dest *[]models.Label) error
	FindLabelByID(repoID, labelID string) (models.Label, error)

	// Comments
	CreateComment(comment *models.IssueComment) error
	FindComments(issueID string, dest *[]models.IssueComment) error
	FindCommentByID(id string) (models.IssueComment, error)

	// Assignees
	AddAssignee(issueID, userID string) error
	RemoveAssignee(issueID, userID string) error
	FindAssignees(issueID string, dest *[]models.User) error
}

type IssueRepositoryImpl struct {
	db *gorm.DB
}

func NewIssueRepository(db *gorm.DB) *IssueRepositoryImpl {
	return &IssueRepositoryImpl{db: db}
}

func (r *IssueRepositoryImpl) Create(issue *models.Issue) error {
	return r.db.Create(issue).Error
}

func (r *IssueRepositoryImpl) FindByRepo(repoID string, dest *[]models.Issue) error {
	return r.db.
		Preload("Author").
		Preload("Assignee").
		Preload("Labels").
		Where("repo_id = ?", repoID).
		Order("number ASC").
		Find(dest).Error
}

func (r *IssueRepositoryImpl) FindByNumber(repoID string, number int) (models.Issue, error) {
	var issue models.Issue
	err := r.db.
		Preload("Author").
		Preload("Assignee").
		Preload("Labels").
		Where("repo_id = ? AND number = ?", repoID, number).
		First(&issue).Error
	return issue, err
}

func (r *IssueRepositoryImpl) FindByID(id string) (models.Issue, error) {
	var issue models.Issue
	err := r.db.
		Preload("Author").
		Preload("Assignee").
		Preload("Labels").
		Where("id = ?", id).
		First(&issue).Error
	return issue, err
}

func (r *IssueRepositoryImpl) Update(issue *models.Issue) error {
	return r.db.Model(issue).Updates(map[string]interface{}{
		"title":       issue.Title,
		"body":        issue.Body,
		"state":       issue.State,
		"assignee_id": issue.AssigneeID,
		"closed_at":   issue.ClosedAt,
	}).Error
}

func (r *IssueRepositoryImpl) Delete(issue *models.Issue) error {
	return r.db.Delete(issue).Error
}

func (r *IssueRepositoryImpl) NextNumber(repoID string) (int, error) {
	var max int
	err := r.db.
		Model(&models.Issue{}).
		Where("repo_id = ?", repoID).
		Select("COALESCE(MAX(number), 0)").
		Scan(&max).Error
	if err != nil {
		return 0, err
	}
	return max + 1, nil
}

func (r *IssueRepositoryImpl) CreateLabel(label *models.Label) error {
	return r.db.Create(label).Error
}

func (r *IssueRepositoryImpl) FindLabels(repoID string, dest *[]models.Label) error {
	return r.db.Where("repo_id = ?", repoID).Order("name ASC").Find(dest).Error
}

func (r *IssueRepositoryImpl) FindLabelByID(repoID, labelID string) (models.Label, error) {
	var label models.Label
	err := r.db.Where("repo_id = ? AND id = ?", repoID, labelID).First(&label).Error
	return label, err
}

func (r *IssueRepositoryImpl) CreateComment(comment *models.IssueComment) error {
	return r.db.Create(comment).Error
}

func (r *IssueRepositoryImpl) FindComments(issueID string, dest *[]models.IssueComment) error {
	return r.db.
		Preload("Author").
		Where("issue_id = ?", issueID).
		Order("created_at ASC").
		Find(dest).Error
}

func (r *IssueRepositoryImpl) FindCommentByID(id string) (models.IssueComment, error) {
	var comment models.IssueComment
	err := r.db.
		Preload("Author").
		Where("id = ?", id).
		First(&comment).Error
	return comment, err
}

func (r *IssueRepositoryImpl) AddAssignee(issueID, userID string) error {
	assignee := models.IssueAssignee{IssueID: issueID, UserID: userID}
	return r.db.Create(&assignee).Error
}

func (r *IssueRepositoryImpl) RemoveAssignee(issueID, userID string) error {
	return r.db.
		Where("issue_id = ? AND user_id = ?", issueID, userID).
		Delete(&models.IssueAssignee{}).Error
}

func (r *IssueRepositoryImpl) FindAssignees(issueID string, dest *[]models.User) error {
	return r.db.
		Joins("JOIN issue_assignees ON issue_assignees.user_id = users.id").
		Where("issue_assignees.issue_id = ?", issueID).
		Find(dest).Error
}
