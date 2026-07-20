package services

import (
	"errors"
	"fmt"
	"time"

	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"gorm.io/gorm"
)

// PatchService manages patch requests: creation with snapshot, refresh, merge,
// reviewer assignment, reviews and comments.
type PatchService interface {
	ListPatches(owner, name string) ([]models.PatchRequest, error)
	CreatePatch(owner, name, authorID, title, body, sourceBranch, targetBranch string) (*models.PatchRequest, error)
	GetPatch(owner, name string, number int) (*models.PatchRequest, error)
	UpdatePatch(owner, name string, number int, title, body, state string) (*models.PatchRequest, error)
	RefreshPatch(owner, name string, number int) (*models.PatchRequest, error)
	MergePatch(owner, name string, number int, mergerID string) (*models.PatchRequest, error)

	ListCommits(owner, name string, number int) ([]models.PatchCommit, error)
	ListFiles(owner, name string, number int) ([]models.PatchFile, error)

	AssignReviewer(owner, name string, number int, username string) error
	UnassignReviewer(owner, name string, number int, username string) error
	ListReviewers(owner, name string, number int) ([]models.User, error)

	SubmitReview(owner, name string, number int, authorID, state, body string) error
	ListReviews(owner, name string, number int) ([]models.PatchReview, error)

	CreateComment(owner, name string, number int, authorID, body, filePath string, line *int) (*models.PatchComment, error)
	ListComments(owner, name string, number int) ([]models.PatchComment, error)
}

type PatchServiceImpl struct {
	RepoService RepoService
	Patches     repository.PatchRepository
	Users       repository.UserRepository
	Git         GitQueryService
}

func NewPatchService(db *gorm.DB) PatchService {
	return &PatchServiceImpl{
		RepoService: NewRepoService(db),
		Patches:     repository.NewPatchRepository(db),
		Users:       repository.NewUserRepository(db),
		Git:         NewGitQueryService(db),
	}
}

func (s *PatchServiceImpl) resolve(owner, name string) (*models.Repository, error) {
	return s.RepoService.GetRepository(owner, name)
}

func (s *PatchServiceImpl) ListPatches(owner, name string) ([]models.PatchRequest, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	var patches []models.PatchRequest
	if err := s.Patches.FindByRepo(repo.ID.String(), &patches); err != nil {
		return nil, err
	}
	return patches, nil
}

func (s *PatchServiceImpl) CreatePatch(owner, name, authorID, title, body, sourceBranch, targetBranch string) (*models.PatchRequest, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}

	// Validate branches exist.
	srcExists, err := s.Git.BranchExists(repo.ID.String(), sourceBranch)
	if err != nil {
		return nil, err
	}
	if !srcExists {
		return nil, fmt.Errorf("source branch %q not found", sourceBranch)
	}
	tgtExists, err := s.Git.BranchExists(repo.ID.String(), targetBranch)
	if err != nil {
		return nil, err
	}
	if !tgtExists {
		return nil, fmt.Errorf("target branch %q not found", targetBranch)
	}

	baseSHA, err := s.Git.BranchHead(repo.ID.String(), targetBranch)
	if err != nil {
		return nil, err
	}
	headSHA, err := s.Git.BranchHead(repo.ID.String(), sourceBranch)
	if err != nil {
		return nil, err
	}

	// Snapshot commits & files.
	commits, err := s.Git.SnapshotCommits(repo.ID.String(), baseSHA, headSHA)
	if err != nil {
		return nil, err
	}
	files, err := s.Git.SnapshotFiles(repo.ID.String(), baseSHA, headSHA)
	if err != nil {
		return nil, err
	}

	// Atomic number allocation + create in one transaction.
	var created models.PatchRequest
	err = s.Patches.WithTx(func(tx repository.PatchRepository) error {
		number, err := tx.NextNumber(repo.ID.String())
		if err != nil {
			return err
		}
		patch := models.PatchRequest{
			RepoID:       repo.ID.String(),
			Number:       number,
			Title:        title,
			Body:         body,
			SourceBranch: sourceBranch,
			TargetBranch: targetBranch,
			AuthorID:     authorID,
			State:        models.PatchStateOpen,
			BaseSHA:      baseSHA,
			HeadSHA:      headSHA,
		}
		if err := tx.Create(&patch); err != nil {
			return err
		}
		// Persist snapshot.
		patchCommits := toPatchCommits(patch.ID.String(), commits)
		if err := tx.ReplaceCommits(patch.ID.String(), patchCommits); err != nil {
			return err
		}
		patchFiles := toPatchFiles(patch.ID.String(), files)
		if err := tx.ReplaceFiles(patch.ID.String(), patchFiles); err != nil {
			return err
		}
		// Pin source commits so they are not GC'd.
		if err := s.Git.CreateRef(repo.ID.String(), fmt.Sprintf("refs/patches/%d/head", number), headSHA); err != nil {
			return err
		}
		created = patch
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Check mergeability (best-effort; cache result).
	mergeable, mErr := s.Git.CheckMergeable(repo.ID.String(), baseSHA, headSHA)
	if mErr == nil {
		created.IsMergeable = &mergeable
		if err := s.Patches.Update(&created); err != nil {
			return nil, err
		}
	}

	// Reload with associations.
	reloaded, err := s.Patches.FindByID(created.ID.String())
	if err != nil {
		return nil, err
	}
	return &reloaded, nil
}

func (s *PatchServiceImpl) GetPatch(owner, name string, number int) (*models.PatchRequest, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	patch, err := s.Patches.FindByNumber(repo.ID.String(), number)
	if err != nil {
		return nil, errors.New("patch not found")
	}
	return &patch, nil
}

func (s *PatchServiceImpl) UpdatePatch(owner, name string, number int, title, body, state string) (*models.PatchRequest, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	patch, err := s.Patches.FindByNumber(repo.ID.String(), number)
	if err != nil {
		return nil, errors.New("patch not found")
	}
	if title != "" {
		patch.Title = title
	}
	if body != "" {
		patch.Body = body
	}
	if state != "" && models.ValidPatchState(state) {
		patch.State = state
		if state == models.PatchStateClosed && patch.ClosedAt == nil {
			now := time.Now()
			patch.ClosedAt = &now
		}
	}
	if err := s.Patches.Update(&patch); err != nil {
		return nil, err
	}
	reloaded, err := s.Patches.FindByID(patch.ID.String())
	if err != nil {
		return nil, err
	}
	return &reloaded, nil
}

func (s *PatchServiceImpl) RefreshPatch(owner, name string, number int) (*models.PatchRequest, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	patch, err := s.Patches.FindByNumber(repo.ID.String(), number)
	if err != nil {
		return nil, errors.New("patch not found")
	}
	if patch.State != models.PatchStateOpen {
		return nil, errors.New("only open patches can be refreshed")
	}

	newBase, err := s.Git.BranchHead(repo.ID.String(), patch.TargetBranch)
	if err != nil {
		return nil, err
	}
	newHead, err := s.Git.BranchHead(repo.ID.String(), patch.SourceBranch)
	if err != nil {
		return nil, err
	}
	if newBase == patch.BaseSHA && newHead == patch.HeadSHA {
		return &patch, nil // no-op
	}

	commits, err := s.Git.SnapshotCommits(repo.ID.String(), newBase, newHead)
	if err != nil {
		return nil, err
	}
	files, err := s.Git.SnapshotFiles(repo.ID.String(), newBase, newHead)
	if err != nil {
		return nil, err
	}

	patch.BaseSHA = newBase
	patch.HeadSHA = newHead
	if err := s.Patches.ReplaceCommits(patch.ID.String(), toPatchCommits(patch.ID.String(), commits)); err != nil {
		return nil, err
	}
	if err := s.Patches.ReplaceFiles(patch.ID.String(), toPatchFiles(patch.ID.String(), files)); err != nil {
		return nil, err
	}
	if err := s.Git.CreateRef(repo.ID.String(), fmt.Sprintf("refs/patches/%d/head", patch.Number), newHead); err != nil {
		return nil, err
	}
	mergeable, mErr := s.Git.CheckMergeable(repo.ID.String(), newBase, newHead)
	if mErr == nil {
		patch.IsMergeable = &mergeable
	}
	if err := s.Patches.Update(&patch); err != nil {
		return nil, err
	}
	reloaded, err := s.Patches.FindByID(patch.ID.String())
	if err != nil {
		return nil, err
	}
	return &reloaded, nil
}

func (s *PatchServiceImpl) MergePatch(owner, name string, number int, mergerID string) (*models.PatchRequest, error) {
	repo, err := s.resolve(owner, name)
	if err != nil {
		return nil, err
	}
	patch, err := s.Patches.FindByNumber(repo.ID.String(), number)
	if err != nil {
		return nil, errors.New("patch not found")
	}
	if patch.State != models.PatchStateOpen {
		return nil, errors.New("patch is not open")
	}

	// Guard: at least one assigned reviewer must have approved.
	reviewers, err := s.listReviewersRaw(patch.ID.String())
	if err != nil {
		return nil, err
	}
	if len(reviewers) == 0 {
		return nil, errors.New("at least one reviewer must be assigned")
	}
	approved, err := s.allReviewersApproved(patch.ID.String())
	if err != nil {
		return nil, err
	}
	if !approved {
		return nil, errors.New("all assigned reviewers must approve before merge")
	}

	// Use the CURRENT target branch tip (not the stale snapshot base).
	targetSHA, err := s.Git.BranchHead(repo.ID.String(), patch.TargetBranch)
	if err != nil {
		return nil, err
	}

	merger, err := s.Users.FindByID(mergerID)
	if err != nil {
		return nil, err
	}

	mergeCommit, err := s.Git.PerformMerge(
		repo.ID.String(),
		patch.TargetBranch,
		targetSHA,
		patch.HeadSHA,
		fmt.Sprintf("Merge patch #%d: %s", patch.Number, patch.Title),
		merger.Username,
		merger.Email,
	)
	if errors.Is(err, ErrMergeConflict) {
		return nil, errors.New("merge conflict: please refresh the patch and resolve conflicts")
	}
	if err != nil {
		return nil, err
	}

	// Git succeeded; now update DB.
	now := time.Now()
	patch.State = models.PatchStateMerged
	patch.MergeCommitSHA = &mergeCommit
	patch.MergedAt = &now
	if err := s.Patches.Update(&patch); err != nil {
		return nil, err
	}
	reloaded, err := s.Patches.FindByID(patch.ID.String())
	if err != nil {
		return nil, err
	}
	return &reloaded, nil
}

func (s *PatchServiceImpl) ListCommits(owner, name string, number int) ([]models.PatchCommit, error) {
	patch, err := s.GetPatch(owner, name, number)
	if err != nil {
		return nil, err
	}
	var commits []models.PatchCommit
	if err := s.Patches.FindCommits(patch.ID.String(), &commits); err != nil {
		return nil, err
	}
	return commits, nil
}

func (s *PatchServiceImpl) ListFiles(owner, name string, number int) ([]models.PatchFile, error) {
	patch, err := s.GetPatch(owner, name, number)
	if err != nil {
		return nil, err
	}
	var files []models.PatchFile
	if err := s.Patches.FindFiles(patch.ID.String(), &files); err != nil {
		return nil, err
	}
	return files, nil
}

func (s *PatchServiceImpl) AssignReviewer(owner, name string, number int, username string) error {
	patch, err := s.GetPatch(owner, name, number)
	if err != nil {
		return err
	}
	user, err := s.Users.FindByUsername(username)
	if err != nil {
		return errors.New("user not found")
	}
	return s.Patches.AddReviewer(patch.ID.String(), user.ID.String())
}

func (s *PatchServiceImpl) UnassignReviewer(owner, name string, number int, username string) error {
	patch, err := s.GetPatch(owner, name, number)
	if err != nil {
		return err
	}
	user, err := s.Users.FindByUsername(username)
	if err != nil {
		return errors.New("user not found")
	}
	return s.Patches.RemoveReviewer(patch.ID.String(), user.ID.String())
}

func (s *PatchServiceImpl) ListReviewers(owner, name string, number int) ([]models.User, error) {
	patch, err := s.GetPatch(owner, name, number)
	if err != nil {
		return nil, err
	}
	var reviewers []models.User
	if err := s.Patches.FindReviewers(patch.ID.String(), &reviewers); err != nil {
		return nil, err
	}
	return reviewers, nil
}

func (s *PatchServiceImpl) listReviewersRaw(patchID string) ([]models.User, error) {
	var reviewers []models.User
	if err := s.Patches.FindReviewers(patchID, &reviewers); err != nil {
		return nil, err
	}
	return reviewers, nil
}

func (s *PatchServiceImpl) allReviewersApproved(patchID string) (bool, error) {
	var reviews []models.PatchReview
	if err := s.Patches.FindReviews(patchID, &reviews); err != nil {
		return false, err
	}
	approvedSet := make(map[string]bool)
	for _, r := range reviews {
		if r.State == models.PatchReviewApproved {
			approvedSet[r.AuthorID] = true
		}
	}
	reviewers, err := s.listReviewersRaw(patchID)
	if err != nil {
		return false, err
	}
	if len(reviewers) == 0 {
		return false, nil
	}
	for _, rv := range reviewers {
		if !approvedSet[rv.ID.String()] {
			return false, nil
		}
	}
	return true, nil
}

func (s *PatchServiceImpl) SubmitReview(owner, name string, number int, authorID, state, body string) error {
	if !models.ValidPatchReviewState(state) {
		return errors.New("invalid review state")
	}
	patch, err := s.GetPatch(owner, name, number)
	if err != nil {
		return err
	}
	isRev, err := s.Patches.IsReviewer(patch.ID.String(), authorID)
	if err != nil {
		return err
	}
	if !isRev {
		return errors.New("only assigned reviewers can submit a review")
	}
	review := &models.PatchReview{
		PatchID:  patch.ID.String(),
		AuthorID: authorID,
		State:    state,
		Body:     body,
	}
	return s.Patches.UpsertReview(review)
}

func (s *PatchServiceImpl) ListReviews(owner, name string, number int) ([]models.PatchReview, error) {
	patch, err := s.GetPatch(owner, name, number)
	if err != nil {
		return nil, err
	}
	var reviews []models.PatchReview
	if err := s.Patches.FindReviews(patch.ID.String(), &reviews); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (s *PatchServiceImpl) CreateComment(owner, name string, number int, authorID, body, filePath string, line *int) (*models.PatchComment, error) {
	patch, err := s.GetPatch(owner, name, number)
	if err != nil {
		return nil, err
	}
	comment := &models.PatchComment{
		PatchID:  patch.ID.String(),
		AuthorID: authorID,
		Body:     body,
	}
	if filePath != "" {
		comment.FilePath = &filePath
		comment.Line = line
	}
	if err := s.Patches.CreateComment(comment); err != nil {
		return nil, err
	}
	reloaded, err := s.Patches.FindCommentByID(comment.ID.String())
	if err != nil {
		return nil, err
	}
	return &reloaded, nil
}

func (s *PatchServiceImpl) ListComments(owner, name string, number int) ([]models.PatchComment, error) {
	patch, err := s.GetPatch(owner, name, number)
	if err != nil {
		return nil, err
	}
	var comments []models.PatchComment
	if err := s.Patches.FindComments(patch.ID.String(), &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

// --- converters ---

func toPatchCommits(patchID string, commits []CommitInfo) []models.PatchCommit {
	result := make([]models.PatchCommit, 0, len(commits))
	for _, c := range commits {
		result = append(result, models.PatchCommit{
			PatchID:     patchID,
			SHA:         c.SHA,
			Message:     c.Message,
			AuthorName:  c.Author,
			AuthorEmail: c.Email,
			AuthorDate:  c.Date,
		})
	}
	return result
}

func toPatchFiles(patchID string, files []PatchFileSnapshot) []models.PatchFile {
	result := make([]models.PatchFile, 0, len(files))
	for _, f := range files {
		result = append(result, models.PatchFile{
			PatchID:  patchID,
			FilePath: f.FilePath,
			Status:   f.Status,
			Diff:     f.Diff,
		})
	}
	return result
}
