package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/gupta/leetcode-judge/internal/config"
	"github.com/gupta/leetcode-judge/internal/models"
	"github.com/gupta/leetcode-judge/internal/repository"
)

type Service struct {
	userRepo *repository.UserRepository
	jwtCfg   *config.JWTConfig
}

func NewService(userRepo *repository.UserRepository, jwtCfg *config.JWTConfig) *Service {
	return &Service{userRepo: userRepo, jwtCfg: jwtCfg}
}

func (s *Service) Register(ctx context.Context, username, email, password string) (*models.User, string, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     "user",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := GenerateToken(user.ID, user.Username, user.Role, s.jwtCfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if !CheckPassword(user.Password, password) {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := GenerateToken(user.ID, user.Username, user.Role, s.jwtCfg)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}
