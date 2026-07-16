package services

import (
	"errors"
	"gitxyz/internal/helper"
	"gitxyz/internal/models"
	"gitxyz/internal/repository"

	"gorm.io/gorm"
)

type UserService interface {
	// SSH keys
	AddSSHKey(userID, title, publicKey string) (models.SSHKey, error)
	ListSSHKeys(userID string) ([]models.SSHKey, error)
	DeleteSSHKey(userID, keyID string) error

	// Personal access tokens
	CreateToken(userID, name, scopes string) (models.PersonalAccessToken, string, error)
	ListTokens(userID string) ([]models.PersonalAccessToken, error)
	DeleteToken(userID, tokenID string) error
}

type UserServiceImpl struct {
	Keys   repository.SSHKeyRepository
	Tokens repository.PATRepository
}

func NewUserService(db *gorm.DB) UserService {
	return &UserServiceImpl{
		Keys:   repository.NewSSHKeyRepository(db),
		Tokens: repository.NewPATRepository(db),
	}
}

func (s *UserServiceImpl) AddSSHKey(userID, title, publicKey string) (models.SSHKey, error) {
	if title == "" {
		return models.SSHKey{}, errors.New("key title is required")
	}
	if publicKey == "" {
		return models.SSHKey{}, errors.New("public key is required")
	}

	fingerprint, err := helper.FingerprintSSHKey(publicKey)
	if err != nil {
		return models.SSHKey{}, err
	}

	if s.Keys.ExistsByFingerprint(fingerprint) {
		return models.SSHKey{}, errors.New("this ssh key is already registered")
	}

	key := &models.SSHKey{
		Title:       title,
		PublicKey:   publicKey,
		Fingerprint: fingerprint,
		UserID:      userID,
	}

	if err := s.Keys.Create(key); err != nil {
		return models.SSHKey{}, err
	}
	return *key, nil
}

func (s *UserServiceImpl) ListSSHKeys(userID string) ([]models.SSHKey, error) {
	return s.Keys.FindByUserID(userID)
}

func (s *UserServiceImpl) DeleteSSHKey(userID, keyID string) error {
	key, err := s.Keys.FindByID(keyID)
	if err != nil {
		return err
	}
	if key.UserID != userID {
		return errors.New("not authorized to delete this key")
	}
	return s.Keys.Delete(keyID)
}

func (s *UserServiceImpl) CreateToken(userID, name, scopes string) (models.PersonalAccessToken, string, error) {
	if name == "" {
		return models.PersonalAccessToken{}, "", errors.New("token name is required")
	}

	plain, prefix, hash, err := helper.GenerateToken()
	if err != nil {
		return models.PersonalAccessToken{}, "", err
	}

	token := &models.PersonalAccessToken{
		Name:        name,
		TokenHash:   hash,
		TokenPrefix: prefix,
		Scopes:      scopes,
		UserID:      userID,
	}

	if err := s.Tokens.Create(token); err != nil {
		return models.PersonalAccessToken{}, "", err
	}

	return *token, plain, nil
}

func (s *UserServiceImpl) ListTokens(userID string) ([]models.PersonalAccessToken, error) {
	return s.Tokens.FindByUserID(userID)
}

func (s *UserServiceImpl) DeleteToken(userID, tokenID string) error {
	token, err := s.Tokens.FindByID(tokenID)
	if err != nil {
		return err
	}
	if token.UserID != userID {
		return errors.New("not authorized to delete this token")
	}
	return s.Tokens.Delete(tokenID)
}

var _ = gorm.ErrRecordNotFound
