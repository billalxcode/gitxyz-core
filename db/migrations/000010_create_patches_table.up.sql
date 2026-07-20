-- Add atomic counter column for patch request numbering (independent from issues).
ALTER TABLE repositories ADD COLUMN last_item_number INTEGER NOT NULL DEFAULT 0;

CREATE TABLE patch_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    source_branch VARCHAR(255) NOT NULL,
    target_branch VARCHAR(255) NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    state VARCHAR(10) NOT NULL DEFAULT 'open',
    base_sha CHAR(40),
    head_sha CHAR(40),
    merge_commit_sha CHAR(40),
    is_mergeable BOOLEAN,
    merged_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(repo_id, number),
    CONSTRAINT chk_patch_state CHECK (state IN ('open', 'merged', 'closed'))
);

CREATE INDEX idx_patch_requests_repo_id ON patch_requests(repo_id);
CREATE INDEX idx_patch_requests_state ON patch_requests(state);
CREATE INDEX idx_patch_requests_deleted_at ON patch_requests(deleted_at);

CREATE TABLE patch_commits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patch_id UUID NOT NULL REFERENCES patch_requests(id) ON DELETE CASCADE,
    sha CHAR(40) NOT NULL,
    message TEXT,
    author_name VARCHAR(255),
    author_email VARCHAR(255),
    author_date TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_patch_commits_patch_id ON patch_commits(patch_id);
CREATE INDEX idx_patch_commits_deleted_at ON patch_commits(deleted_at);

CREATE TABLE patch_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patch_id UUID NOT NULL REFERENCES patch_requests(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    diff TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_patch_files_patch_id ON patch_files(patch_id);
CREATE INDEX idx_patch_files_deleted_at ON patch_files(deleted_at);

CREATE TABLE patch_reviewers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patch_id UUID NOT NULL REFERENCES patch_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(patch_id, user_id)
);

CREATE INDEX idx_patch_reviewers_patch_id ON patch_reviewers(patch_id);
CREATE INDEX idx_patch_reviewers_user_id ON patch_reviewers(user_id);

CREATE TABLE patch_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patch_id UUID NOT NULL REFERENCES patch_requests(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    state VARCHAR(20) NOT NULL,
    body TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_patch_reviews_patch_id ON patch_reviews(patch_id);
CREATE INDEX idx_patch_reviews_deleted_at ON patch_reviews(deleted_at);

CREATE TABLE patch_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patch_id UUID NOT NULL REFERENCES patch_requests(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    file_path TEXT,
    line INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_patch_comments_patch_id ON patch_comments(patch_id);
CREATE INDEX idx_patch_comments_deleted_at ON patch_comments(deleted_at);
