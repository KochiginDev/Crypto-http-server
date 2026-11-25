package service

import (
	"context"
	"crypthttp/domain"
	"crypthttp/repository"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	User   repository.User
	secret []byte
	Living time.Duration
}

func NewAuthService(user repository.User, secret []byte, life time.Duration) *AuthService {
	return &AuthService{User: user, secret: secret, Living: life}
}
func (service *AuthService) generateToken(us domain.User) (string, error) {
	now := time.Now()
	mp := jwt.MapClaims{
		"sub":      strconv.FormatInt(us.Id, 10),
		"username": us.Username,
		"iat":      now.Unix(),
		"exp":      now.Add(service.Living).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mp)
	str, err := token.SignedString(service.secret)
	if err != nil {
		return "", fmt.Errorf("signed string: %w", err)
	}
	return str, nil
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

	token, err := service.generateToken(us)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
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
	token, err := service.generateToken(us)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}

func (service *AuthService) CheckToken(ctx context.Context, token string) (domain.User, error) {
	if len(token) == 0 {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	hs, err := jwt.Parse(token, func(t *jwt.Token) (any, error) { return service.secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	claims, ok := hs.Claims.(jwt.MapClaims)
	if !ok {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	username, ok := claims["username"].(string)
	if !ok {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	return domain.User{Id: id, Username: username}, nil

}
