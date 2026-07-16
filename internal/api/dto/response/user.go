package dto

import (
	"gitxyz/internal/models"
	"time"
)

type SSHKeyResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToSSHKeyResponse(key *models.SSHKey) SSHKeyResponse {
	return SSHKeyResponse{
		ID:          key.ID.String(),
		Title:       key.Title,
		Fingerprint: key.Fingerprint,
		CreatedAt:   key.CreatedAt,
	}
}

type TokenResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	Scopes      string     `json:"scopes"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func ToTokenResponse(token *models.PersonalAccessToken) TokenResponse {
	return TokenResponse{
		ID:          token.ID.String(),
		Name:        token.Name,
		TokenPrefix: token.TokenPrefix,
		Scopes:      token.Scopes,
		LastUsedAt:  token.LastUsedAt,
		ExpiresAt:   token.ExpiresAt,
		CreatedAt:   token.CreatedAt,
	}
}

func ToSSHKeyResponseSlice(keys []models.SSHKey) []SSHKeyResponse {
	out := make([]SSHKeyResponse, 0, len(keys))
	for i := range keys {
		out = append(out, ToSSHKeyResponse(&keys[i]))
	}
	return out
}

func ToTokenResponseSlice(tokens []models.PersonalAccessToken) []TokenResponse {
	out := make([]TokenResponse, 0, len(tokens))
	for i := range tokens {
		out = append(out, ToTokenResponse(&tokens[i]))
	}
	return out
}

type UserResponse struct {
	ID          string    `json:"id"`
	FullName    string    `json:"full_name"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	IsActive    bool      `json:"is_active"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	Location    string    `json:"location"`
	LastLoginAt time.Time `json:"last_login_at"`
}

func ToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:          user.ID.String(),
		FullName:    user.FullName,
		Username:    user.Username,
		Email:       user.Email,
		IsActive:    user.IsActive,
		Avatar:      user.Avatar,
		Bio:         user.Bio,
		Location:    user.Location,
		LastLoginAt: user.LastLoginAt,
	}
}
