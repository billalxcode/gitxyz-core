package services

import (
	"errors"
	"fmt"
	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(user *models.User) error
	Login(usernameOrEmail, password string) (models.User, error)
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
