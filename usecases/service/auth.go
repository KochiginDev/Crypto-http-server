package service

import (
	"context"
	"crypthttp/domain"
	"crypthttp/repository"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	User repository.User
}

func NewAuthService(user repository.User) *AuthService {
	return &AuthService{User: user}
}
func (service *AuthService) generateToken() (string, error) {
	temp := make([]byte, 32)
	_, err := rand.Read(temp)
	if err != nil {
		return "", fmt.Errorf("create token: %w", err)
	}
	token := hex.EncodeToString(temp)
	return token, nil
}
func (service *AuthService) Register(ctx context.Context, username string, password string) (string, error) {
	if len(password) > 50 {
		return "", domain.ErrInvalidInput
	}
	if len(password) < 6 {
		return "", domain.ErrInvalidInput
	}
	username = strings.TrimSpace(username)
	if len(username) == 0 {
		return "", domain.ErrInvalidInput
	}
	if len(username) > 50 {
		return "", domain.ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}

	us, err := service.User.Create(ctx, username, string(hash))
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	token, err := service.generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	err = service.User.SaveNewToken(ctx, us.Id, token)
	if err != nil {
		return "", fmt.Errorf("save new token: %w", err)
	}
	return token, nil
}

func (service *AuthService) Login(ctx context.Context, username string, password string) (string, error) {
	username = strings.TrimSpace(username)
	if len(username) == 0 {
		return "", domain.ErrInvalidInput
	}
	if len(password) == 0 {
		return "", domain.ErrInvalidInput
	}
	us, err := service.User.GetByName(ctx, username)

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", fmt.Errorf("get user by name: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(us.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", domain.ErrInvalidCredentials
		}
		return "", fmt.Errorf("compare password: %w", err)
	}
	token, err := service.generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	err = service.User.SaveNewToken(ctx, us.Id, token)
	if err != nil {
		return "", fmt.Errorf("save new token: %w", err)
	}
	return token, nil
}

func (service *AuthService) CheckToken(ctx context.Context, token string) (domain.User, error) {
	if len(token) == 0 {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	us, err := service.User.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, domain.ErrInvalidCredentials
		}
		return domain.User{}, fmt.Errorf("check token: %w", err)
	}
	return us, nil
}
