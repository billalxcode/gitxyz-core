DROP TABLE IF EXISTS patch_comments;
DROP TABLE IF EXISTS patch_reviews;
DROP TABLE IF EXISTS patch_reviewers;
DROP TABLE IF EXISTS patch_files;
DROP TABLE IF EXISTS patch_commits;
DROP TABLE IF EXISTS patch_requests;

ALTER TABLE repositories DROP COLUMN IF EXISTS last_item_number;
