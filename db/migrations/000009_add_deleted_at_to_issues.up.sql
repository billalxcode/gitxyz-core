ALTER TABLE issues ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_issues_deleted_at ON issues(deleted_at);

ALTER TABLE labels ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_labels_deleted_at ON labels(deleted_at);

ALTER TABLE issue_comments ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_issue_comments_deleted_at ON issue_comments(deleted_at);
