package domain

import (
	"errors"
	"time"
)

type User struct {
	Id           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}
type Crypto struct {
	Symbol       string    `json:"symbol"`
	Name         string    `json:"name"`
	CurrentPrice float64   `json:"current_price"`
	LastUpdate   time.Time `json:"last_updated"`
	CoinGeckoId  string    `json:"-"`
}
type PricePoint struct {
	Price float64   `json:"price"`
	Time  time.Time `json:"timestamp"`
}

type Statistic struct {
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
	AvgPrice     float64 `json:"avg_price"`
	PriceChange  float64 `json:"price_change"`
	PercentDiff  float64 `json:"price_change_percent"`
	RecordsCount int64   `json:"records_count"`
}

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUnknownSymbol       = errors.New("unknown symbol")
	ErrProviderUnavailable = errors.New("provider unavailable")
)
