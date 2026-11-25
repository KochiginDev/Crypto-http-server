package service

import (
	"context"
	"crypthttp/domain"
	"crypthttp/provider"
	"crypthttp/repository"
	"errors"
	"strings"
	"time"
)

type CryptoService struct {
	repo     repository.Crypto
	resolver provider.Resolver
	price    provider.PriceSource
}

func NewCryptoService(repo repository.Crypto, resolver provider.Resolver, price provider.PriceSource) *CryptoService {
	return &CryptoService{repo: repo, resolver: resolver, price: price}
}

func normalize(symbol string) string {
	symbol = strings.TrimSpace(symbol)
	symbol = strings.ToUpper(symbol)
	return symbol
}
func (s *CryptoService) Add(ctx context.Context, symbol string) (domain.Crypto, error) {
	symbol = normalize(symbol)
	if symbol == "" {
		return domain.Crypto{}, domain.ErrInvalidInput
	}
	_, err := s.repo.Get(ctx, symbol)
	if err == nil {
		return domain.Crypto{}, domain.ErrAlreadyExists
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.Crypto{}, err
	}
	id, name, err := s.resolver.Resolve(ctx, symbol)
	if err != nil {
		return domain.Crypto{}, err
	}
	price, err := s.price.GetPrice(ctx, id)
	if err != nil {
		return domain.Crypto{}, err
	}
	crypt := domain.Crypto{
		Symbol:       symbol,
		Name:         name,
		CurrentPrice: price,
		LastUpdate:   time.Now(),
		CoinGeckoId:  id,
	}
	err = s.repo.Create(ctx, crypt)
	if err != nil {
		return domain.Crypto{}, err
	}
	return crypt, nil
}

func (s *CryptoService) Refresh(ctx context.Context, symbol string) (domain.Crypto, error) {
	symbol = normalize(symbol)
	if symbol == "" {
		return domain.Crypto{}, domain.ErrInvalidInput
	}
	v, err := s.repo.Get(ctx, symbol)
	if err != nil {
		return domain.Crypto{}, err
	}
	price, err := s.price.GetPrice(ctx, v.CoinGeckoId)
	if err != nil {
		return domain.Crypto{}, err
	}
	now := time.Now()
	err = s.repo.UpdatePrice(ctx, symbol, price, now)
	if err != nil {
		return domain.Crypto{}, err
	}
	v.LastUpdate = now
	v.CurrentPrice = price
	return v, nil
}
func (s *CryptoService) GetHistory(ctx context.Context, symbol string) ([]domain.PricePoint, error) {
	symbol = normalize(symbol)
	if symbol == "" {
		return nil, domain.ErrInvalidInput
	}
	_, err := s.repo.Get(ctx, symbol)
	if err != nil {
		return nil, err
	}
	prices, err := s.repo.GetHistory(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return prices, nil
}

func (s *CryptoService) GetStats(ctx context.Context, symbol string) (domain.Statistic, error) {
	symbol = normalize(symbol)
	if symbol == "" {
		return domain.Statistic{}, domain.ErrInvalidInput
	}
	_, err := s.repo.Get(ctx, symbol)
	if err != nil {
		return domain.Statistic{}, err
	}
	stats, err := s.repo.GetStats(ctx, symbol)
	if err != nil {
		return domain.Statistic{}, err
	}
	return stats, nil
}

func (s *CryptoService) Get(ctx context.Context, symbol string) (domain.Crypto, error) {
	symbol = normalize(symbol)
	if symbol == "" {
		return domain.Crypto{}, domain.ErrInvalidInput
	}
	crypt, err := s.repo.Get(ctx, symbol)
	if err != nil {
		return domain.Crypto{}, err
	}
	return crypt, nil
}
func (s *CryptoService) GetList(ctx context.Context) ([]domain.Crypto, error) {
	return s.repo.GetList(ctx)
}
func (s *CryptoService) Delete(ctx context.Context, symbol string) error {
	symbol = normalize(symbol)
	if symbol == "" {
		return domain.ErrInvalidInput
	}
	return s.repo.Delete(ctx, symbol)
}
