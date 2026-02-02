package services

import (
	"context"
	"time"

	"github.com/lucas/go-rest-api-mongo/internal/dto"
	"github.com/lucas/go-rest-api-mongo/internal/models"
	"github.com/lucas/go-rest-api-mongo/internal/repositories"
)

type UserService struct {
	repo        *repositories.UserRepository
	authService *AuthService
}

func NewUserService(repo *repositories.UserRepository, authService *AuthService) *UserService {
	return &UserService{
		repo:        repo,
		authService: authService,
	}
}

func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*models.User, error) {
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := s.authService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:     req.Email,
		Password:  hashedPassword,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := s.authService.ComparePassword(user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.authService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID.Hex(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.String(),
		},
	}, nil
}
