package models

import (
	"time"

	"gitxyz/internal/helper"

	"gorm.io/gorm"
)

type User struct {
	Base

	FullName string `json:"full_name" gorm:"size:255;not null"`
	Username string `json:"username" gorm:"sze:255;not null;unique"`
	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Password string `json:"-" gorm:"size:255"`

	IsActive bool   `json:"is_active" gorm:"index;not null"`
	Avatar   string `json:"avatar" gorm:"not null"`
	Bio      string `json:"bio"`
	Location string `json:"location"`

	LastLoginAt time.Time `json:"last_login_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if err := u.Base.BeforeCreate(tx); err != nil {
		return err
	}

	passwordHash, err := helper.HashPassword(u.Password)
	if err != nil {
		return err
	}

	u.Password = passwordHash
	u.LastLoginAt = time.Now()
	return nil
}
