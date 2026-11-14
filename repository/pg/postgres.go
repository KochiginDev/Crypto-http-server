package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateAuthTable(ctx context.Context, db *pgx.Conn) error {
	query := `
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    model VARCHAR(100) NOT NULL,
    price NUMERIC(10, 2) NOT NULL
);`
	_, err := db.Exec(ctx, query)
	if err != nil {
		return err
	}
	return nil
}
