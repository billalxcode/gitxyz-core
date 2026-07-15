package dto

import (
	"gitxyz/internal/models"
	"time"
)

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
