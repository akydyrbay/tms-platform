package auth

import (
	"context"
	"errors"

	"tms-platform/internal/model"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserRepository interface {
	CreateWithPassword(ctx context.Context, name, email, passwordHash string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type AuthService struct {
	repo *Repository
	jwt  *JWTManager
}

func NewAuthService(repo *Repository, jwt *JWTManager) *AuthService {
	return &AuthService{repo: repo, jwt: jwt}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*model.User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("name, email and password are required")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateWithPassword(ctx, name, email, hash)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if !CheckPassword(password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}
	return s.jwt.Generate(user.ID, user.Email, user.Role)
}
