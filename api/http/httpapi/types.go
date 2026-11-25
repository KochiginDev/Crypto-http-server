package httpapi

import "crypthttp/domain"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type TokenResponse struct {
	Token string `json:"token"`
}
type AddCryptoRequest struct {
	Symbol string `json:"symbol"`
}
type StatsResponse struct {
	Symbol string           `json:"symbol"`
	Price  float64          `json:"current_price"`
	Stats  domain.Statistic `json:"stats"`
}

type CryptoResponse struct {
	Crypto domain.Crypto `json:"crypto"`
}
type CryptoListResponse struct {
	Cryptos []domain.Crypto `json:"cryptos"`
}
type HistoryResponse struct {
	Symbol  string              `json:"symbol"`
	History []domain.PricePoint `json:"history"`
}
