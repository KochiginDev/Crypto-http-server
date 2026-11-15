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

	// SaveNewToken записывает пользователю новый токен.
	// Возвращает domain.ErrNotFound, если пользователя с таким id нет.
	SaveNewToken(ctx context.Context, id int64, token string) error

	// GetByToken возвращает пользователя по токену.
	// Возвращает domain.ErrNotFound, если токен неизвестен.
	GetByToken(ctx context.Context, token string) (domain.User, error)
}
