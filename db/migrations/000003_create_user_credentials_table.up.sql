CREATE TABLE ssh_keys (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    public_key TEXT NOT NULL,
    fingerprint VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(fingerprint)
);

CREATE INDEX idx_ssh_keys_user_id ON ssh_keys(user_id);
CREATE INDEX idx_ssh_keys_deleted_at ON ssh_keys(deleted_at);

CREATE TABLE personal_access_tokens (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    token_prefix VARCHAR(16) NOT NULL,
    scopes TEXT,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(token_hash)
);

CREATE INDEX idx_personal_access_tokens_user_id ON personal_access_tokens(user_id);
CREATE INDEX idx_personal_access_tokens_deleted_at ON personal_access_tokens(deleted_at);
