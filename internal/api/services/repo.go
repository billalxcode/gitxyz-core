package services

import (
	"errors"
	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"gorm.io/gorm"
)

type RepoService interface {
	CreateRepository(repo *models.Repository) error
	GetRepository(owner, name string) (*models.Repository, error)
	ListRepositories(owner string) ([]models.Repository, error)
	UpdateRepository(owner, name string, repo *models.Repository) (*models.Repository, error)
	DeleteRepository(owner, name string) error

	// Collaborators (repository members).
	ListCollaborators(owner, name string) ([]models.RepositoryMember, error)
	AddCollaborator(owner, name, username, role string) (*models.RepositoryMember, error)
	UpdateCollaborator(owner, name, username, role string) (*models.RepositoryMember, error)
	RemoveCollaborator(owner, name, username string) error

	// Policies (ABAC).
	ListPolicies(owner, name string) ([]models.Policy, error)
	AddPolicy(owner, name, subjectType, subjectID, action, resourceID, effect string) (*models.Policy, error)
	RemovePolicy(owner, name, policyID string) error
}

type RepoServiceImpl struct {
	Repository repository.RepoRepository
	Members    repository.RepoMemberRepository
	Policies   repository.PolicyRepository
	Users      repository.UserRepository
}

func NewRepoService(db *gorm.DB) RepoService {
	return &RepoServiceImpl{
		Repository: repository.NewRepoRepository(db),
		Members:    repository.NewRepoMemberRepository(db),
		Policies:   repository.NewPolicyRepository(db),
		Users:      repository.NewUserRepository(db),
	}
}

func (s *RepoServiceImpl) CreateRepository(repo *models.Repository) error {
	if repo.Name == "" {
		return errors.New("repository name is required")
	}
	if s.Repository.ExistsByName(repo.Name) {
		return errors.New("repository already exists")
	}
	if repo.UserID == "" {
		return errors.New("user id is required")
	}

	// The on-disk path is derived from repo.ID at runtime (volume_path/<repoID>),
	// so no physical_path column is stored.
	return s.Repository.Create(repo)
}

// resolveRepo returns the repo identified by owner+name, verifying ownership.
func (s *RepoServiceImpl) resolveRepo(owner, name string) (*models.Repository, error) {
	user, err := s.Users.FindByUsername(owner)
	if err != nil {
		return nil, errors.New("owner not found")
	}
	repo, err := s.Repository.FindByUserAndName(user.ID.String(), name)
	if err != nil {
		return nil, errors.New("repository not found")
	}
	return &repo, nil
}

func (s *RepoServiceImpl) GetRepository(owner, name string) (*models.Repository, error) {
	return s.resolveRepo(owner, name)
}

func (s *RepoServiceImpl) ListRepositories(owner string) ([]models.Repository, error) {
	user, err := s.Users.FindByUsername(owner)
	if err != nil {
		return nil, errors.New("owner not found")
	}
	var repos []models.Repository
	if err := s.Repository.(*repository.RepoRepositoryImpl).ListByUser(user.ID.String(), &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func (s *RepoServiceImpl) UpdateRepository(owner, name string, patch *models.Repository) (*models.Repository, error) {
	repo, err := s.resolveRepo(owner, name)
	if err != nil {
		return nil, err
	}
	repo.Description = patch.Description
	repo.IsPrivate = patch.IsPrivate
	repo.IsActive = patch.IsActive
	if err := s.Repository.Update(repo); err != nil {
		return nil, err
	}
	return repo, nil
}

func (s *RepoServiceImpl) DeleteRepository(owner, name string) error {
	repo, err := s.resolveRepo(owner, name)
	if err != nil {
		return err
	}
	return s.Repository.Delete(repo.ID.String())
}

func (s *RepoServiceImpl) ListCollaborators(owner, name string) ([]models.RepositoryMember, error) {
	repo, err := s.resolveRepo(owner, name)
	if err != nil {
		return nil, err
	}
	return s.Members.FindByRepo(repo.ID.String())
}

func (s *RepoServiceImpl) AddCollaborator(owner, name, username, role string) (*models.RepositoryMember, error) {
	if !models.ValidRepoRole(role) {
		return nil, errors.New("invalid repository role")
	}
	repo, err := s.resolveRepo(owner, name)
	if err != nil {
		return nil, err
	}
	user, err := s.Users.FindByUsername(username)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if repo.UserID == user.ID.String() {
		return nil, errors.New("cannot add the repository owner as a collaborator")
	}
	member := &models.RepositoryMember{
		UserID: user.ID.String(),
		RepoID: repo.ID.String(),
		Role:   role,
	}
	if err := s.Members.Add(member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *RepoServiceImpl) UpdateCollaborator(owner, name, username, role string) (*models.RepositoryMember, error) {
	if !models.ValidRepoRole(role) {
		return nil, errors.New("invalid repository role")
	}
	repo, err := s.resolveRepo(owner, name)
	if err != nil {
		return nil, err
	}
	user, err := s.Users.FindByUsername(username)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if err := s.Members.UpdateRole(user.ID.String(), repo.ID.String(), role); err != nil {
		return nil, err
	}
	m, err := s.Members.FindByUserAndRepo(user.ID.String(), repo.ID.String())
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *RepoServiceImpl) RemoveCollaborator(owner, name, username string) error {
	repo, err := s.resolveRepo(owner, name)
	if err != nil {
		return err
	}
	user, err := s.Users.FindByUsername(username)
	if err != nil {
		return errors.New("user not found")
	}
	return s.Members.Remove(user.ID.String(), repo.ID.String())
}

func (s *RepoServiceImpl) ListPolicies(owner, name string) ([]models.Policy, error) {
	repo, err := s.resolveRepo(owner, name)
	if err != nil {
		return nil, err
	}
	return s.Policies.FindApplicable("user", "*", "*", "repository", repo.ID.String())
}

func (s *RepoServiceImpl) AddPolicy(owner, name, subjectType, subjectID, action, resourceID, effect string) (*models.Policy, error) {
	if effect != "allow" && effect != "deny" {
		return nil, errors.New("effect must be allow or deny")
	}
	if subjectType != "user" && subjectType != "role" {
		return nil, errors.New("subject_type must be user or role")
	}
	repo, err := s.resolveRepo(owner, name)
	if err != nil {
		return nil, err
	}
	pol := &models.Policy{
		SubjectType:  subjectType,
		SubjectID:    subjectID,
		Action:       action,
		ResourceType: "repository",
		ResourceID:   repo.ID.String(),
		Effect:       effect,
	}
	if err := s.Policies.Add(pol); err != nil {
		return nil, err
	}
	return pol, nil
}

func (s *RepoServiceImpl) RemovePolicy(owner, name, policyID string) error {
	repo, err := s.resolveRepo(owner, name)
	if err != nil {
		return err
	}
	// Ensure the policy belongs to this repository before deleting.
	policies, err := s.Policies.FindApplicable("user", "*", "*", "repository", repo.ID.String())
	if err != nil {
		return err
	}
	found := false
	for _, p := range policies {
		if p.ID.String() == policyID {
			found = true
			break
		}
	}
	if !found {
		return errors.New("policy not found for this repository")
	}
	return s.Policies.Remove(policyID)
}
