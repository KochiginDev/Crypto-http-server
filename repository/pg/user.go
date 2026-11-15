package pg

import (
	"context"
	"crypthttp/domain"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}
func (r *UserRepo) Create(ctx context.Context, username string, passwordHash string) (domain.User, error) {
	query := "INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id"
	var id int64
	err := r.pool.QueryRow(ctx, query, username, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return domain.User{}, fmt.Errorf("user already exists: %w", domain.ErrAlreadyExists)
			}
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return domain.User{Id: id, Username: username, PasswordHash: passwordHash}, nil
}

func (r *UserRepo) GetByName(ctx context.Context, username string) (domain.User, error) {
	query := "SELECT id, username, password_hash, COALESCE(token, '') FROM users WHERE username = $1"
	var user domain.User
	err := r.pool.QueryRow(ctx, query, username).Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by name: %w", err)
	}
	return user, nil
}

func (r *UserRepo) GetByToken(ctx context.Context, token string) (domain.User, error) {
	query := "SELECT id, username, password_hash, COALESCE(token, '') FROM users WHERE token = $1"
	var user domain.User
	err := r.pool.QueryRow(ctx, query, token).Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by token: %w", err)
	}
	return user, nil
}

func (r *UserRepo) SaveNewToken(ctx context.Context, id int64, token string) error {
	query := "UPDATE users SET token = $1 WHERE id = $2"
	tag, err := r.pool.Exec(ctx, query, token, id)
	if err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
