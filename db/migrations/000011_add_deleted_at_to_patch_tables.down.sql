ALTER TABLE patch_commits DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE patch_files DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE patch_reviewers DROP COLUMN IF EXISTS deleted_at;
