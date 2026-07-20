package services

import (
	"errors"
	"time"

	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"gorm.io/gorm"
)

type IssueService interface {
	ListIssues(owner, name string) ([]models.Issue, error)
	CreateIssue(owner, name, authorID string, title, body string, labelNames []string) (*models.Issue, error)
	GetIssue(owner, name string, number int) (*models.Issue, error)
	UpdateIssue(owner, name string, number int, patch *models.Issue, labelNames []string) (*models.Issue, error)
	DeleteIssue(owner, name string, number int) error

	ListLabels(owner, name string) ([]models.Label, error)
	CreateLabel(owner, name, labelName, color, description string) (*models.Label, error)

	ListComments(owner, name string, number int) ([]models.IssueComment, error)
	CreateComment(owner, name string, number int, authorID, body string) (*models.IssueComment, error)

	ListAssignees(owner, name string, number int) ([]models.User, error)
	AddAssignee(owner, name string, number int, username string) error
	RemoveAssignee(owner, name string, number int, username string) error
}

type IssueServiceImpl struct {
	RepoService RepoService
	Issues      repository.IssueRepository
	Users       repository.UserRepository
}

func NewIssueService(db *gorm.DB) IssueService {
	return &IssueServiceImpl{
		RepoService: NewRepoService(db),
		Issues:      repository.NewIssueRepository(db),
		Users:       repository.NewUserRepository(db),
	}
}

func (s *IssueServiceImpl) resolve(owner, name string) (*models.Repository, error) {
	return s.RepoService.GetRepository(owner, name)
}

func (s *IssueServiceImpl) ListIssues(owner, name string) ([]models.Issue, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	var issues []models.Issue
	if err := s.Issues.FindByRepo(repo.ID.String(), &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

func (s *IssueServiceImpl) CreateIssue(owner, name, authorID, title, body string, labelNames []string) (*models.Issue, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	number, err := s.Issues.NextNumber(repo.ID.String())
	if err != nil {
		return nil, err
	}

	issue := &models.Issue{
		RepoID:   repo.ID.String(),
		Number:   number,
		Title:    title,
		Body:     body,
		State:    models.IssueStateOpen,
		AuthorID: authorID,
	}

	if len(labelNames) > 0 {
		if err := s.attachLabels(repo.ID.String(), issue, labelNames); err != nil {
			return nil, err
		}
	}

	if err := s.Issues.Create(issue); err != nil {
		return nil, err
	}

	// Reload with associations (Author, Assignee, Labels) so the returned
	// issue is fully populated for the response DTO.
	created, err := s.Issues.FindByID(issue.ID.String())
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *IssueServiceImpl) GetIssue(owner, name string, number int) (*models.Issue, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	issue, err := s.Issues.FindByNumber(repo.ID.String(), number)
	if err != nil {
		return nil, errors.New("issue not found")
	}
	return &issue, nil
}

func (s *IssueServiceImpl) UpdateIssue(owner, name string, number int, patch *models.Issue, labelNames []string) (*models.Issue, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	issue, err := s.Issues.FindByNumber(repo.ID.String(), number)
	if err != nil {
		return nil, errors.New("issue not found")
	}

	if patch.Title != "" {
		issue.Title = patch.Title
	}
	if patch.Body != "" {
		issue.Body = patch.Body
	}
	if patch.State != "" {
		issue.State = patch.State
		if patch.State == models.IssueStateClosed && issue.ClosedAt == nil {
			now := time.Now()
			issue.ClosedAt = &now
		}
		if patch.State == models.IssueStateOpen {
			issue.ClosedAt = nil
		}
	}

	if labelNames != nil {
		if err := s.attachLabels(repo.ID.String(), &issue, labelNames); err != nil {
			return nil, err
		}
	}

	if err := s.Issues.Update(&issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (s *IssueServiceImpl) DeleteIssue(owner, name string, number int) error {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return err
	}
	issue, err := s.Issues.FindByNumber(repo.ID.String(), number)
	if err != nil {
		return errors.New("issue not found")
	}
	return s.Issues.Delete(&issue)
}

func (s *IssueServiceImpl) ListLabels(owner, name string) ([]models.Label, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	var labels []models.Label
	if err := s.Issues.FindLabels(repo.ID.String(), &labels); err != nil {
		return nil, err
	}
	return labels, nil
}

func (s *IssueServiceImpl) CreateLabel(owner, name, labelName, color, description string) (*models.Label, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	label := &models.Label{
		RepoID:      repo.ID.String(),
		Name:        labelName,
		Color:       color,
		Description: description,
	}
	if label.Color == "" {
		label.Color = "#cccccc"
	}
	if err := s.Issues.CreateLabel(label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *IssueServiceImpl) ListComments(owner, name string, number int) ([]models.IssueComment, error) {
	issue, err := s.GetIssue(owner, name, number)
	if err != nil {
		return nil, err
	}
	var comments []models.IssueComment
	if err := s.Issues.FindComments(issue.ID.String(), &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (s *IssueServiceImpl) CreateComment(owner, name string, number int, authorID, body string) (*models.IssueComment, error) {
	issue, err := s.GetIssue(owner, name, number)
	if err != nil {
		return nil, err
	}
	comment := &models.IssueComment{
		IssueID:  issue.ID.String(),
		AuthorID: authorID,
		Body:     body,
	}
	if err := s.Issues.CreateComment(comment); err != nil {
		return nil, err
	}

	// Reload with Author preloaded so the response DTO is fully populated.
	created, err := s.Issues.FindCommentByID(comment.ID.String())
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *IssueServiceImpl) ListAssignees(owner, name string, number int) ([]models.User, error) {
	issue, err := s.GetIssue(owner, name, number)
	if err != nil {
		return nil, err
	}
	var users []models.User
	if err := s.Issues.FindAssignees(issue.ID.String(), &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *IssueServiceImpl) AddAssignee(owner, name string, number int, username string) error {
	issue, err := s.GetIssue(owner, name, number)
	if err != nil {
		return err
	}
	user, err := s.Users.FindByUsername(username)
	if err != nil {
		return errors.New("user not found")
	}
	return s.Issues.AddAssignee(issue.ID.String(), user.ID.String())
}

func (s *IssueServiceImpl) RemoveAssignee(owner, name string, number int, username string) error {
	issue, err := s.GetIssue(owner, name, number)
	if err != nil {
		return err
	}
	user, err := s.Users.FindByUsername(username)
	if err != nil {
		return errors.New("user not found")
	}
	return s.Issues.RemoveAssignee(issue.ID.String(), user.ID.String())
}

// attachLabels resolves label names to existing labels and assigns them to the issue.
func (s *IssueServiceImpl) attachLabels(repoID string, issue *models.Issue, labelNames []string) error {
	var all []models.Label
	if err := s.Issues.FindLabels(repoID, &all); err != nil {
		return err
	}
	byName := make(map[string]models.Label, len(all))
	for i := range all {
		byName[all[i].Name] = all[i]
	}

	labels := make([]models.Label, 0, len(labelNames))
	for _, name := range labelNames {
		label, ok := byName[name]
		if !ok {
			return errors.New("label not found: " + name)
		}
		labels = append(labels, label)
	}
	issue.Labels = labels
	return nil
}
