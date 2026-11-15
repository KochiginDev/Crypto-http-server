package provider

import "context"

// Resolver сопоставляет тикер монеты с её идентификатором в CoinGecko.
type Resolver interface {
	// Resolve по символу возвращает CoinGecko id и название монеты.
	// Символ приходит уже нормализованным (в верхнем регистре),
	// выбор из нескольких кандидатов с одним тикером остаётся на реализации.
	// Возвращает domain.ErrUnknownSymbol, если такого символа в CoinGecko нет,
	// и domain.ErrProviderUnavailable при сетевой ошибке, таймауте,
	// ответе 5xx или превышении лимита запросов.
	Resolve(ctx context.Context, symbol string) (coingeckoID string, name string, err error)
}

// PriceSource отдаёт текущие цены монет.
type PriceSource interface {
	// GetPrice возвращает текущую цену монеты в USD по её CoinGecko id.
	// Время получения цены фиксирует вызывающая сторона.
	// Возвращает domain.ErrUnknownSymbol, если для этого id цены нет в ответе
	// (CoinGecko в таком случае отдаёт пустой объект, а не ошибку),
	// и domain.ErrProviderUnavailable при сетевой ошибке, таймауте,
	// ответе 5xx или превышении лимита запросов.
	GetPrice(ctx context.Context, coingeckoID string) (price float64, err error)
}
