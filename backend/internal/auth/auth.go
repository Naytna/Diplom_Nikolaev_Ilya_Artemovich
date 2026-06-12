package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken = errors.New("некорректный токен")
	ErrExpiredToken = errors.New("срок действия токена истёк")
)

type User struct {
	ID           int64  `json:"id"`
	Role         string `json:"role"`
	FullName     string `json:"full_name"`
	PasswordHash string `json:"-"`
}

type tokenClaims struct {
	UserID   int64  `json:"uid"`
	Role     string `json:"role"`
	FullName string `json:"full_name"`
	Expires  int64  `json:"exp"`
}

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *TokenManager) IssueToken(user User) (string, error) {
	payload, err := json.Marshal(tokenClaims{
		UserID:   user.ID,
		Role:     user.Role,
		FullName: user.FullName,
		Expires:  time.Now().Add(m.ttl).Unix(),
	})
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := m.sign(encodedPayload)

	return encodedPayload + "." + signature, nil
}

func (m *TokenManager) ParseToken(token string) (User, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return User{}, ErrInvalidToken
	}

	expectedSignature := m.sign(parts[0])
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
		return User{}, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return User{}, ErrInvalidToken
	}

	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return User{}, ErrInvalidToken
	}

	if claims.Expires < time.Now().Unix() {
		return User{}, ErrExpiredToken
	}

	return User{
		ID:       claims.UserID,
		Role:     claims.Role,
		FullName: claims.FullName,
	}, nil
}

func CheckPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (m *TokenManager) sign(value string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
