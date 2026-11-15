package usecases

import (
	"context"
	"crypthttp/domain"
)

type Crypto interface {
	//Add
	//Пустой символ  -ErrInvalidInput
	//Дубликат - ErrAlreadyExists
	//Неизвестный символ - ErrUnknownSymbol
	//сбой провайдера - ErrProviderUnavailable
	Add(ctx context.Context, symbol string) (domain.Crypto, error)
	//GetList
	//Пустой список - не ошибка
	GetList(ctx context.Context) ([]domain.Crypto, error)
	//Get
	//нет монеты -ErrNotFound
	Get(ctx context.Context, symbol string) (domain.Crypto, error)
	//Refresh
	//нет монеты - ErrNotFound
	Refresh(ctx context.Context, symbol string) (domain.Crypto, error)
	//GetHistory - GetStats - Delete
	//несуществующий символ - ErrNotFound
	GetHistory(ctx context.Context, symbol string) ([]domain.PricePoint, error)
	//GetStats вызывает только статистику, ответ хэндлер собирает из него и из Get
	GetStats(ctx context.Context, symbol string) (domain.Statistic, error)
	Delete(ctx context.Context, symbol string) error
}
