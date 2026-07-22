-- Add missing deleted_at columns to patch snapshot/reviewer tables.
-- These tables embed models.Base (gorm.DeletedAt) so GORM expects the column.
ALTER TABLE patch_commits ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE patch_files ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE patch_reviewers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_patch_commits_deleted_at ON patch_commits(deleted_at);
CREATE INDEX IF NOT EXISTS idx_patch_files_deleted_at ON patch_files(deleted_at);
CREATE INDEX IF NOT EXISTS idx_patch_reviewers_deleted_at ON patch_reviewers(deleted_at);
