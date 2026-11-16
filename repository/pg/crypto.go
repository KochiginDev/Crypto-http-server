package pg

import (
	"context"
	"crypthttp/domain"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CryptoRepo struct {
	pool *pgxpool.Pool
}

func NewCryptoRepo(pool *pgxpool.Pool) *CryptoRepo {
	return &CryptoRepo{pool: pool}
}
func (r *CryptoRepo) Create(ctx context.Context, crypto domain.Crypto) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Create begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	query := "INSERT INTO cryptos (symbol, name, current_price, last_updated, coingecko_id) VALUES ($1, $2, $3, $4, $5)"
	_, err = tx.Exec(ctx, query, crypto.Symbol, crypto.Name, crypto.CurrentPrice, crypto.LastUpdate, crypto.CoinGeckoId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("crypto already exists: %w", domain.ErrAlreadyExists)
		}
		return fmt.Errorf("insert crypto: %w", err)
	}

	query = "INSERT INTO price_history (symbol, price, update_time) VALUES ($1, $2, $3)"
	_, err = tx.Exec(ctx, query, crypto.Symbol, crypto.CurrentPrice, crypto.LastUpdate)
	if err != nil {
		return fmt.Errorf("insert history: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *CryptoRepo) Get(ctx context.Context, symbol string) (domain.Crypto, error) {
	var crypto domain.Crypto
	query := "SELECT symbol, name, current_price, last_updated, coingecko_id FROM cryptos WHERE symbol = $1"
	err := r.pool.QueryRow(ctx, query, symbol).Scan(&crypto.Symbol, &crypto.Name, &crypto.CurrentPrice, &crypto.LastUpdate, &crypto.CoinGeckoId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Crypto{}, domain.ErrNotFound
		}
		return domain.Crypto{}, fmt.Errorf("get crypto: %w", err)
	}
	crypto.LastUpdate = crypto.LastUpdate.UTC()
	return crypto, nil
}
func (r *CryptoRepo) GetList(ctx context.Context) ([]domain.Crypto, error) {
	crypto := []domain.Crypto{}
	query := "SELECT symbol, name, current_price, last_updated, coingecko_id FROM cryptos ORDER BY symbol"
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get list by opening query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var crypt domain.Crypto
		err = rows.Scan(&crypt.Symbol, &crypt.Name, &crypt.CurrentPrice, &crypt.LastUpdate, &crypt.CoinGeckoId)
		if err != nil {
			return nil, fmt.Errorf("get list while scan: %w", err)
		}
		crypt.LastUpdate = crypt.LastUpdate.UTC()
		crypto = append(crypto, crypt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get list: %w", err)
	}
	return crypto, nil
}

func (r *CryptoRepo) UpdatePrice(ctx context.Context, symbol string, price float64, updated_at time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("updatePrice begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	query := "UPDATE cryptos SET current_price=$1, last_updated=$2 WHERE symbol = $3"
	tag, err := tx.Exec(ctx, query, price, updated_at, symbol)
	if err != nil {
		return fmt.Errorf("update price at cryptos: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update price at cryptos: %w", domain.ErrNotFound)
	}
	query = "INSERT INTO price_history (symbol, price, update_time) VALUES ($1, $2, $3)"
	_, err = tx.Exec(ctx, query, symbol, price, updated_at)
	if err != nil {
		return fmt.Errorf("insert history: %w", err)
	}
	query = "DELETE FROM price_history WHERE symbol = $1 AND id NOT IN (SELECT id FROM price_history WHERE symbol = $1 ORDER BY update_time DESC, id DESC LIMIT 100)"
	_, err = tx.Exec(ctx, query, symbol)
	if err != nil {
		return fmt.Errorf("delete history more than 100: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
func (r *CryptoRepo) GetHistory(ctx context.Context, symbol string) ([]domain.PricePoint, error) {
	history := []domain.PricePoint{}
	query := "SELECT price, update_time FROM price_history WHERE symbol = $1 ORDER BY update_time ASC,id ASC"
	rows, err := r.pool.Query(ctx, query, symbol)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var point domain.PricePoint
		err = rows.Scan(&point.Price, &point.Time)
		if err != nil {
			return nil, fmt.Errorf("while get history: %w", err)
		}
		point.Time = point.Time.UTC()
		history = append(history, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	return history, nil
}
func (r *CryptoRepo) Delete(ctx context.Context, symbol string) error {
	query := "DELETE FROM cryptos WHERE symbol = $1"
	tag, err := r.pool.Exec(ctx, query, symbol)
	if err != nil {
		return fmt.Errorf("delete cryptos: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete cryptos: %w", domain.ErrNotFound)
	}
	return nil
}

func (r *CryptoRepo) GetStats(ctx context.Context, symbol string) (domain.Statistic, error) {
	var stats domain.Statistic
	query := "SELECT " +
		"COALESCE(MIN(price), 0)," +
		"COALESCE(MAX(price), 0), " +
		"COALESCE(AVG(price), 0), " +
		"COUNT(*)," +
		"COALESCE((SELECT price FROM price_history WHERE symbol = $1 ORDER BY update_time ASC, id ASC LIMIT 1), 0), " +
		"COALESCE((SELECT price FROM price_history WHERE symbol =$1 ORDER BY update_time DESC, id DESC LIMIT 1), 0)" +
		"FROM price_history WHERE symbol = $1"
	var first float64
	var last float64
	var count int64

	err := r.pool.QueryRow(ctx, query, symbol).Scan(&stats.MinPrice, &stats.MaxPrice, &stats.AvgPrice, &count, &first, &last)
	if err != nil {
		return domain.Statistic{}, fmt.Errorf("get stats: %w", err)
	}
	stats.RecordsCount = count
	stats.PriceChange = last - first
	if first == 0 {
		stats.PercentDiff = 0
	} else {
		stats.PercentDiff = stats.PriceChange / first * 100
	}
	return stats, nil
}
