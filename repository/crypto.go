package repository

import (
	"context"
	"crypthttp/domain"
	"time"
)

type Crypto interface {
	Create(ctx context.Context, object domain.Crypto) error
	Get(ctx context.Context, symbol string) (domain.Crypto, error)
	GetList(ctx context.Context) ([]domain.Crypto, error)
	UpdatePrice(ctx context.Context, symbol string, newPrice float64, updatedAt time.Time) error
	GetHistory(ctx context.Context, symbol string) ([]domain.PricePoint, error)
	GetStats(ctx context.Context, symbol string) (domain.Statistic, error)
	Delete(ctx context.Context, symbol string) error
}
