CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_type VARCHAR(20) NOT NULL,
    subject_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(20) NOT NULL,
    resource_id VARCHAR(255) NOT NULL DEFAULT '*',
    effect VARCHAR(10) NOT NULL DEFAULT 'allow',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(subject_type, subject_id, action, resource_type, resource_id),
    CONSTRAINT chk_policy_effect CHECK (effect IN ('allow', 'deny'))
);

CREATE INDEX idx_policies_lookup
    ON policies(subject_type, subject_id, action, resource_type, resource_id);
