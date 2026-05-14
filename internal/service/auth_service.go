package service

import (
	"context"
	"errors"
	"time"

	"shopapi/internal/model"
	"shopapi/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	users   *repository.UserRepository
	secret  []byte
	expiryH int
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, expiryHours int) *AuthService {
	if expiryHours <= 0 {
		expiryHours = 72
	}
	return &AuthService{
		users:   users,
		secret:  []byte(jwtSecret),
		expiryH: expiryHours,
	}
}

type RegisterInput struct {
	Username string
	Password string
	Role     string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*model.User, error) {
	if len(in.Password) < 8 {
		return nil, errors.New("password too short")
	}
	if in.Role == "" {
		in.Role = "user"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Username:     in.Username,
		PasswordHash: string(hash),
		Role:         in.Role,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

type LoginInput struct {
	Username string
	Password string
}

type TokenPair struct {
	Token     string
	ExpiresAt time.Time
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*model.User, *TokenPair, error) {
	u, err := s.users.GetByUsername(ctx, in.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New("invalid credentials")
		}
		return nil, nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)) != nil {
		return nil, nil, errors.New("invalid credentials")
	}
	exp := time.Now().Add(time.Duration(s.expiryH) * time.Hour)
	tok, err := s.signJWT(u.ID, u.Username, u.Role, exp)
	if err != nil {
		return nil, nil, err
	}
	return u, &TokenPair{Token: tok, ExpiresAt: exp}, nil
}

type jwtClaims struct {
	UserID   uint64 `json:"uid"`
	Username string `json:"sub"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) signJWT(userID uint64, username, role string, exp time.Time) (string, error) {
	claims := jwtClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.secret)
}

func (s *AuthService) ParseToken(token string) (*jwtClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
