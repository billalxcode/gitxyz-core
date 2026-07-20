package repository

import (
	"gitxyz/internal/models"

	"gorm.io/gorm"
)

// PatchRepository provides data access for patch requests and their relations.
type PatchRepository interface {
	// Patch requests
	Create(patch *models.PatchRequest) error
	FindByRepo(repoID string, dest *[]models.PatchRequest) error
	FindByNumber(repoID string, number int) (models.PatchRequest, error)
	FindByID(id string) (models.PatchRequest, error)
	Update(patch *models.PatchRequest) error
	Delete(patch *models.PatchRequest) error
	NextNumber(repoID string) (int, error)

	// Snapshot commits
	ReplaceCommits(patchID string, commits []models.PatchCommit) error
	FindCommits(patchID string, dest *[]models.PatchCommit) error

	// Snapshot files
	ReplaceFiles(patchID string, files []models.PatchFile) error
	FindFiles(patchID string, dest *[]models.PatchFile) error

	// Reviewers
	AddReviewer(patchID, userID string) error
	RemoveReviewer(patchID, userID string) error
	FindReviewers(patchID string, dest *[]models.User) error
	IsReviewer(patchID, userID string) (bool, error)

	// Reviews
	UpsertReview(review *models.PatchReview) error
	FindReviews(patchID string, dest *[]models.PatchReview) error

	// Comments
	CreateComment(comment *models.PatchComment) error
	FindComments(patchID string, dest *[]models.PatchComment) error
	FindCommentByID(id string) (models.PatchComment, error)

	// Transaction
	WithTx(fn func(tx PatchRepository) error) error
}

type PatchRepositoryImpl struct {
	db *gorm.DB
}

func NewPatchRepository(db *gorm.DB) *PatchRepositoryImpl {
	return &PatchRepositoryImpl{db: db}
}

func (r *PatchRepositoryImpl) Create(patch *models.PatchRequest) error {
	return r.db.Create(patch).Error
}

func (r *PatchRepositoryImpl) FindByRepo(repoID string, dest *[]models.PatchRequest) error {
	return r.db.
		Preload("Author").
		Preload("Reviewers.User").
		Where("repo_id = ?", repoID).
		Order("number ASC").
		Find(dest).Error
}

func (r *PatchRepositoryImpl) FindByNumber(repoID string, number int) (models.PatchRequest, error) {
	var patch models.PatchRequest
	err := r.db.
		Preload("Author").
		Preload("Reviewers.User").
		Preload("Reviews.Author").
		Preload("Commits").
		Preload("Files").
		Where("repo_id = ? AND number = ?", repoID, number).
		First(&patch).Error
	return patch, err
}

func (r *PatchRepositoryImpl) FindByID(id string) (models.PatchRequest, error) {
	var patch models.PatchRequest
	err := r.db.
		Preload("Author").
		Preload("Reviewers.User").
		Preload("Reviews.Author").
		Preload("Commits").
		Preload("Files").
		Where("id = ?", id).
		First(&patch).Error
	return patch, err
}

func (r *PatchRepositoryImpl) Update(patch *models.PatchRequest) error {
	return r.db.
		Model(patch).
		Omit("Author", "Reviewers", "Comments", "Commits", "Files", "Reviews").
		Updates(map[string]interface{}{
			"title":            patch.Title,
			"body":             patch.Body,
			"state":            patch.State,
			"merge_commit_sha": patch.MergeCommitSHA,
			"is_mergeable":     patch.IsMergeable,
			"base_sha":         patch.BaseSHA,
			"head_sha":         patch.HeadSHA,
			"merged_at":        patch.MergedAt,
			"closed_at":        patch.ClosedAt,
		}).Error
}

func (r *PatchRepositoryImpl) Delete(patch *models.PatchRequest) error {
	return r.db.Delete(patch).Error
}

// NextNumber atomically increments and returns the repository's item counter.
// Uses a single UPDATE ... RETURNING to avoid the MAX+1 race condition.
func (r *PatchRepositoryImpl) NextNumber(repoID string) (int, error) {
	var next int
	err := r.db.
		Model(&models.Repository{}).
		Where("id = ?", repoID).
		UpdateColumn("last_item_number", gorm.Expr("last_item_number + 1")).
		Select("last_item_number").
		Scan(&next).Error
	if err != nil {
		return 0, err
	}
	return next, nil
}

func (r *PatchRepositoryImpl) ReplaceCommits(patchID string, commits []models.PatchCommit) error {
	if err := r.db.Where("patch_id = ?", patchID).Delete(&models.PatchCommit{}).Error; err != nil {
		return err
	}
	if len(commits) == 0 {
		return nil
	}
	return r.db.Create(&commits).Error
}

func (r *PatchRepositoryImpl) FindCommits(patchID string, dest *[]models.PatchCommit) error {
	return r.db.
		Where("patch_id = ?", patchID).
		Order("author_date ASC").
		Find(dest).Error
}

func (r *PatchRepositoryImpl) ReplaceFiles(patchID string, files []models.PatchFile) error {
	if err := r.db.Where("patch_id = ?", patchID).Delete(&models.PatchFile{}).Error; err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	return r.db.Create(&files).Error
}

func (r *PatchRepositoryImpl) FindFiles(patchID string, dest *[]models.PatchFile) error {
	return r.db.
		Where("patch_id = ?", patchID).
		Order("file_path ASC").
		Find(dest).Error
}

func (r *PatchRepositoryImpl) AddReviewer(patchID, userID string) error {
	reviewer := models.PatchReviewer{PatchID: patchID, UserID: userID}
	return r.db.Create(&reviewer).Error
}

func (r *PatchRepositoryImpl) RemoveReviewer(patchID, userID string) error {
	return r.db.
		Unscoped().
		Where("patch_id = ? AND user_id = ?", patchID, userID).
		Delete(&models.PatchReviewer{}).Error
}

func (r *PatchRepositoryImpl) FindReviewers(patchID string, dest *[]models.User) error {
	return r.db.
		Select("users.*").
		Joins("JOIN patch_reviewers ON patch_reviewers.user_id = users.id AND patch_reviewers.deleted_at IS NULL").
		Where("patch_reviewers.patch_id = ?", patchID).
		Find(dest).Error
}

func (r *PatchRepositoryImpl) IsReviewer(patchID, userID string) (bool, error) {
	var count int64
	err := r.db.
		Model(&models.PatchReviewer{}).
		Where("patch_id = ? AND user_id = ?", patchID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *PatchRepositoryImpl) UpsertReview(review *models.PatchReview) error {
	var existing models.PatchReview
	err := r.db.
		Where("patch_id = ? AND author_id = ?", review.PatchID, review.AuthorID).
		First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(review).Error
	}
	if err != nil {
		return err
	}
	existing.State = review.State
	existing.Body = review.Body
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"state": review.State,
		"body":  review.Body,
	}).Error
}

func (r *PatchRepositoryImpl) FindReviews(patchID string, dest *[]models.PatchReview) error {
	return r.db.
		Preload("Author").
		Where("patch_id = ?", patchID).
		Order("created_at ASC").
		Find(dest).Error
}

func (r *PatchRepositoryImpl) CreateComment(comment *models.PatchComment) error {
	return r.db.Create(comment).Error
}

func (r *PatchRepositoryImpl) FindComments(patchID string, dest *[]models.PatchComment) error {
	return r.db.
		Preload("Author").
		Where("patch_id = ?", patchID).
		Order("created_at ASC").
		Find(dest).Error
}

func (r *PatchRepositoryImpl) FindCommentByID(id string) (models.PatchComment, error) {
	var comment models.PatchComment
	err := r.db.
		Preload("Author").
		Where("id = ?", id).
		First(&comment).Error
	return comment, err
}

// WithTx runs fn inside a transaction, passing a repository bound to the tx.
func (r *PatchRepositoryImpl) WithTx(fn func(tx PatchRepository) error) error {
	return r.db.Transaction(func(gdb *gorm.DB) error {
		return fn(&PatchRepositoryImpl{db: gdb})
	})
}
