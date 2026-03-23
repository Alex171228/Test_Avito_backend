package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	AdminUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	DummyUserID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

type AuthService struct {
	jwtSecret []byte
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{jwtSecret: []byte(secret)}
}

type TokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateToken(userID uuid.UUID, role string) (string, error) {
	claims := TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ParseToken(tokenStr string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return uuid.Nil, "", err
	}
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return uuid.Nil, "", fmt.Errorf("invalid token claims")
	}
	return claims.UserID, claims.Role, nil
}

func (s *AuthService) DummyLogin(role string) (string, error) {
	var userID uuid.UUID
	switch role {
	case "admin":
		userID = AdminUserID
	case "user":
		userID = DummyUserID
	default:
		return "", fmt.Errorf("invalid role: %s", role)
	}
	return s.GenerateToken(userID, role)
}
