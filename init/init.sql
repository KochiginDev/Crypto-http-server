CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    token TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cryptos (
    symbol VARCHAR(50) PRIMARY KEY,
    current_price DOUBLE PRECISION NOT NULL,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    coingecko_id TEXT NOT NULL,
    name TEXT NOT NULL
);

CREATE TABLE price_history (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL REFERENCES cryptos(symbol)  ON DELETE CASCADE,
    price DOUBLE PRECISION NOT NULL,
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx
ON price_history(symbol, update_time);