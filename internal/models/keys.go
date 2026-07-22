package models

import (
	"fmt"
	"strings"
	"time"
)

// SSHKey represents a user's registered SSH public key, used for Git over SSH
// and as an authentication credential.
type SSHKey struct {
	Base

	Title       string `json:"title" gorm:"size:255;not null"`
	PublicKey   string `json:"-" gorm:"type:text;not null"`
	Fingerprint string `json:"fingerprint" gorm:"size:255;not null"`
	UserID      string `json:"user_id" gorm:"not null;index"`
	User        User   `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

// PersonalAccessToken represents a PAT used to authenticate against the Git
// protocol and the API, instead of the user's password.
type PersonalAccessToken struct {
	Base

	Name        string     `json:"name" gorm:"size:255;not null"`
	TokenHash   string     `json:"-" gorm:"size:255;not null"`
	TokenPrefix string     `json:"token_prefix" gorm:"size:16;not null"`
	Scopes      string     `json:"scopes" gorm:"type:text"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	UserID      string     `json:"user_id" gorm:"not null;index"`
	User        User       `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

// TableName overrides to keep a stable, explicit table name.
func (SSHKey) TableName() string { return "ssh_keys" }

func (PersonalAccessToken) TableName() string { return "personal_access_tokens" }

// PAT scope constants.
const (
	ScopeRepoRead  = "repo:read"
	ScopeRepoWrite = "repo:write"
	ScopeUserRead  = "user:read"
	ScopeUserWrite = "user:write"
	ScopeAdmin     = "admin:*"
)

// ParseScopes splits a comma-separated scope string into a trimmed slice.
func ParseScopes(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// knownScopes is the set of scopes a PAT may request.
var knownScopes = map[string]struct{}{
	ScopeRepoRead:  {},
	ScopeRepoWrite: {},
	ScopeUserRead:  {},
	ScopeUserWrite: {},
	ScopeAdmin:     {},
}

// ValidateScopes returns an error if any scope in raw is not a recognized
// scope constant. The empty string (no scopes) is allowed.
func ValidateScopes(raw string) error {
	for _, s := range ParseScopes(raw) {
		if _, ok := knownScopes[s]; !ok {
			return fmt.Errorf("unknown scope %q", s)
		}
	}
	return nil
}
