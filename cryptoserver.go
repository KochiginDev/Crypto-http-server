package main

import (
	"context"
	"crypthttp/api/http/httpapi"
	"crypthttp/pkg/http"
	"crypthttp/provider/coingecko"
	"crypthttp/repository/pg"
	"crypthttp/usecases/service"
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"time"
)

func generateSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}
func main() {
	dbDSN := "postgres://postgres:mypassword@localhost:5432/productdb"
	jwtSecret := generateSecret()
	coingeckoKey := os.Getenv("COINGECKO_KEY")
	ctx := context.Background()
	ttl := 24 * time.Hour
	pool, err := pg.CreatePool(ctx, dbDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	repo := pg.NewUserRepo(pool)
	auth := service.NewAuthService(repo, []byte(jwtSecret), ttl)
	authHandler := httpapi.NewAuthHandler(auth)
	cryptoRepo := pg.NewCryptoRepo(pool)
	client, err := coingecko.NewClient(nil, "https://api.coingecko.com/api/v3", coingeckoKey)
	if err != nil {
		log.Fatal(err)
	}
	cryptoService := service.NewCryptoService(cryptoRepo, client, client)
	cryptoHandler := httpapi.NewCryptoHandler(cryptoService)
	router := httpapi.NewRouter(authHandler, cryptoHandler, auth)

	err = http.CreateAndRunServer(router, ":8080")
	if err != nil {
		log.Fatal(err)
	}
}
