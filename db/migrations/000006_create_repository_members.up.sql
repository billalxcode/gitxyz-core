CREATE TABLE repository_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, repo_id),
    CONSTRAINT chk_repo_member_role
        CHECK (role IN ('owner', 'maintainer', 'triager', 'reader', 'guest'))
);

CREATE INDEX idx_repo_members_repo_id ON repository_members(repo_id);
CREATE INDEX idx_repo_members_user_id ON repository_members(user_id);
