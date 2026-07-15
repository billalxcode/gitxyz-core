package repository

import (
	"errors"
	"gitxyz/internal/helper"
	"gitxyz/internal/models"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id string) (models.User, error)
	FindByUsername(username string) (models.User, error)
	FindByEmail(email string) (models.User, error)
	FindAll() ([]models.User, error)
	Update(user *models.User) error
	Delete(id string) error
	ExistsByUsername(username string) bool
	ExistsByEmail(email string) bool
	Authenticate(usernameOrEmail, password string) (models.User, error)
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db: db,
	}
}

func checkUser(result *gorm.DB) (models.User, error) {
	if result.Error != nil {
		return models.User{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.User{}, gorm.ErrRecordNotFound
	}
	// ambil model dari result
	if result.Statement != nil {
		if user, ok := result.Statement.Dest.(*models.User); ok {
			return *user, nil
		}
		if userSlice, ok := result.Statement.Dest.(*[]models.User); ok && len(*userSlice) > 0 {
			return (*userSlice)[0], nil
		}
	}
	return models.User{}, errors.New("unable to parse result")
}

func (r *UserRepositoryImpl) Create(user *models.User) error {
	result := r.db.Create(user)
	return result.Error
}

func (r *UserRepositoryImpl) FindByID(id string) (models.User, error) {
	var user models.User
	result := r.db.First(&user, "id = ?", id)
	return checkUser(result)
}

func (r *UserRepositoryImpl) FindByUsername(username string) (models.User, error) {
	var user models.User
	result := r.db.First(&user, "username = ?", username)
	return checkUser(result)
}

func (r *UserRepositoryImpl) FindByEmail(email string) (models.User, error) {
	var user models.User
	result := r.db.First(&user, "email = ?", email)
	return checkUser(result)
}

func (r *UserRepositoryImpl) FindAll() ([]models.User, error) {
	var users []models.User
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *UserRepositoryImpl) Update(user *models.User) error {
	result := r.db.Save(user)
	return result.Error
}

func (r *UserRepositoryImpl) Delete(id string) error {
	result := r.db.Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepositoryImpl) ExistsByUsername(username string) bool {
	var count int64
	r.db.Model(&models.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}

func (r *UserRepositoryImpl) ExistsByEmail(email string) bool {
	var count int64
	r.db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func (r *UserRepositoryImpl) Authenticate(usernameOrEmail, password string) (models.User, error) {
	var user models.User

	result := r.db.Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).First(&user)
	if result.Error != nil {
		return models.User{}, gorm.ErrRecordNotFound
	}

	if err := helper.ComparePassword(user.Password, password); err != nil {
		return models.User{}, errors.New("invalid credentials")
	}

	user.LastLoginAt = time.Now()
	if err := r.db.Model(&user).Update("last_login_at", user.LastLoginAt).Error; err != nil {
		return models.User{}, err
	}

	return user, nil
}
