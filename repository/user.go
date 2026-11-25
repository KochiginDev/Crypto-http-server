package repository

import (
	"context"
	"crypthttp/domain"
)

type User interface {
	// Create сохраняет нового пользователя с уже готовым хешем пароля.
	// Возвращает domain.ErrAlreadyExists, если такое имя уже занято.
	Create(ctx context.Context, username string, passwordHash string) (domain.User, error)

	// GetByName возвращает пользователя по имени.
	// Возвращает domain.ErrNotFound, если такого пользователя нет.
	GetByName(ctx context.Context, username string) (domain.User, error)
}
