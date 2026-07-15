package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"gitxyz/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

var (
	revokedTokens = map[string]struct{}{}
	revokedMu     sync.RWMutex
)

func getJWTSecret() string {
	secret := viper.GetString("jwt_secret")
	if secret == "" {
		secret = "gitxyz-dev-secret-change-me"
	}
	return secret
}

func getJWTExpiry() time.Duration {
	hours := viper.GetInt("jwt_expiry_hours")
	if hours <= 0 {
		hours = 24
	}
	return time.Duration(hours) * time.Hour
}

func getRefreshTokenExpiry() time.Duration {
	day := viper.GetInt("jwt_refresh_expiry_days")
	if day <= 0 {
		day = 7
	}
	return time.Duration(day) * 24 * time.Hour
}

func GenerateToken(user *models.User, tokenType string) (string, error) {
	expiry := getJWTExpiry()
	if tokenType == "refresh" {
		expiry = getRefreshTokenExpiry()
	}

	claims := Claims{
		UserID:    user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getJWTSecret()))
}

func ParseToken(tokenString string) (*Claims, error) {
	revokedMu.RLock()
	_, revoked := revokedTokens[tokenString]
	revokedMu.RUnlock()
	if revoked {
		return nil, errors.New("token revoked")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(getJWTSecret()), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func RevokeToken(tokenString string) {
	revokedMu.Lock()
	defer revokedMu.Unlock()
	revokedTokens[tokenString] = struct{}{}
}

func IsTokenRevoked(tokenString string) bool {
	revokedMu.RLock()
	defer revokedMu.RUnlock()
	_, ok := revokedTokens[tokenString]
	return ok
}
