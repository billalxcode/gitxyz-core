package services

import (
	"errors"
	"fmt"
	"gitxyz/internal/helper"
	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(user *models.User) error
	Login(usernameOrEmail, password string) (models.User, error)
	GetUserByID(id string) (models.User, error)
	GetUserByUsername(username string) (models.User, error)
	UpdateProfile(id string, fullName, bio, location, avatar string) (models.User, error)
	ChangePassword(id, oldPassword, newPassword string) error
}

type AuthServiceImpl struct {
	Repository repository.UserRepository
	Validate   *validator.Validate
}

func NewAuthService(db *gorm.DB) AuthService {
	validate := validator.New()
	userRepository := repository.NewUserRepository(db)

	return &AuthServiceImpl{
		Validate:   validate,
		Repository: userRepository,
	}
}

func (a *AuthServiceImpl) Register(user *models.User) error {
	if a.Repository.ExistsByUsername(user.Username) {
		return errors.New("username already taken")
	}
	if a.Repository.ExistsByEmail(user.Email) {
		return errors.New("email already registered")
	}

	fmt.Println("Create using repository")
	return a.Repository.Create(user)
}

func (a *AuthServiceImpl) Login(usernameOrEmail, password string) (models.User, error) {
	return a.Repository.Authenticate(usernameOrEmail, password)
}

func (a *AuthServiceImpl) GetUserByID(id string) (models.User, error) {
	return a.Repository.FindByID(id)
}

func (a *AuthServiceImpl) GetUserByUsername(username string) (models.User, error) {
	return a.Repository.FindByUsername(username)
}

func (a *AuthServiceImpl) UpdateProfile(id, fullName, bio, location, avatar string) (models.User, error) {
	user, err := a.Repository.FindByID(id)
	if err != nil {
		return models.User{}, err
	}

	if fullName != "" {
		user.FullName = fullName
	}
	if bio != "" {
		user.Bio = bio
	}
	if location != "" {
		user.Location = location
	}
	if avatar != "" {
		user.Avatar = avatar
	}

	if err := a.Repository.Update(&user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (a *AuthServiceImpl) ChangePassword(id, oldPassword, newPassword string) error {
	user, err := a.Repository.FindByID(id)
	if err != nil {
		return err
	}

	if err := helper.ComparePassword(user.Password, oldPassword); err != nil {
		return errors.New("old password is incorrect")
	}

	hashed, err := helper.HashPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashed

	return a.Repository.Update(&user)
}
