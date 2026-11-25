package main

import (
	"context"
	"crypthttp/usecases/service"
	"errors"
	"fmt"
	"time"

	"crypthttp/domain"
	"crypthttp/repository/pg"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testAuth(ctx context.Context, pool *pgxpool.Pool) {
	svc := service.NewAuthService(pg.NewUserRepo(pool), []byte("test-secret"), time.Hour)
	name := fmt.Sprintf("auth_%d", time.Now().UnixNano())

	tok, err := svc.Register(ctx, name, "secret123")
	fmt.Println("a1 register: token len", len(tok), "err:", err)

	_, err = svc.Register(ctx, name, "secret123")
	fmt.Println("a2 duplicate -> ErrAlreadyExists:", errors.Is(err, domain.ErrAlreadyExists))

	_, err = svc.Register(ctx, "x"+name, "123")
	fmt.Println("a3 short password -> ErrInvalidInput:", errors.Is(err, domain.ErrInvalidInput))

	_, err = svc.Register(ctx, "   ", "secret123")
	fmt.Println("a4 blank name -> ErrInvalidInput:", errors.Is(err, domain.ErrInvalidInput))

	u, err := svc.CheckToken(ctx, tok)
	fmt.Printf("a5 check token: id=%d name=%s err=%v\n", u.Id, u.Username, err)

	tok2, err := svc.Login(ctx, name, "secret123")
	fmt.Println("a6 login: err:", err)

	_, err = svc.CheckToken(ctx, tok2)
	fmt.Println("a8 new token err:", err)

	_, err = svc.Login(ctx, name, "wrongpass")
	fmt.Println("a9 wrong password -> ErrInvalidCredentials:", errors.Is(err, domain.ErrInvalidCredentials))
	_, err = svc.Login(ctx, "no_such_"+name, "secret123")
	fmt.Println("a10 unknown user -> ErrInvalidCredentials:", errors.Is(err, domain.ErrInvalidCredentials))
	_, err = svc.Login(ctx, name, "")
	fmt.Println("a11 empty password -> ErrInvalidInput:", errors.Is(err, domain.ErrInvalidInput))
	_, err = svc.CheckToken(ctx, "")
	fmt.Println("a12 empty token -> ErrInvalidCredentials:", errors.Is(err, domain.ErrInvalidCredentials))
	_, err = svc.CheckToken(ctx, "abc")
	fmt.Println("a13:", errors.Is(err, domain.ErrInvalidCredentials))
	
}
func testCrypto(ctx context.Context, pool *pgxpool.Pool) {
	repo := pg.NewCryptoRepo(pool)
	sym := "TST"
	now := time.Now().UTC()

	// убираем хвосты от прошлых запусков
	_ = repo.Delete(ctx, sym)

	c := domain.Crypto{Symbol: sym, Name: "Test Coin", CurrentPrice: 100, LastUpdate: now, CoinGeckoId: "test-coin"}
	err := repo.Create(ctx, c)
	fmt.Println("c1 create err:", err)

	err = repo.Create(ctx, c)
	fmt.Println("c2 duplicate -> ErrAlreadyExists:", errors.Is(err, domain.ErrAlreadyExists))

	got, err := repo.Get(ctx, sym)
	fmt.Printf("c3 get: %+v err=%v\n", got, err)

	_, err = repo.Get(ctx, "NOPE")
	fmt.Println("c4 unknown -> ErrNotFound:", errors.Is(err, domain.ErrNotFound))

	err = repo.UpdatePrice(ctx, sym, 110, now.Add(time.Second))
	fmt.Println("c5 update 110 err:", err)
	err = repo.UpdatePrice(ctx, sym, 120, now.Add(2*time.Second))
	fmt.Println("c6 update 120 err:", err)

	err = repo.UpdatePrice(ctx, "NOPE", 1, now)
	fmt.Println("c7 update unknown -> ErrNotFound:", errors.Is(err, domain.ErrNotFound))

	hist, err := repo.GetHistory(ctx, sym)
	fmt.Printf("c8 history (%d points) err=%v\n", len(hist), err)
	for _, p := range hist {
		fmt.Printf("   %v  %v\n", p.Price, p.Time)
	}

	stats, err := repo.GetStats(ctx, sym)
	fmt.Printf("c9 stats: %+v err=%v\n", stats, err)

	list, err := repo.GetList(ctx)
	fmt.Printf("c10 list: %d items err=%v\n", len(list), err)

	err = repo.Delete(ctx, sym)
	fmt.Println("c11 delete err:", err)
	err = repo.Delete(ctx, sym)
	fmt.Println("c12 delete again -> ErrNotFound:", errors.Is(err, domain.ErrNotFound))

	_, err = repo.Get(ctx, sym)
	fmt.Println("c13 get after delete -> ErrNotFound:", errors.Is(err, domain.ErrNotFound))
	hist, err = repo.GetHistory(ctx, sym)
	fmt.Printf("c14 history after delete: %d points (должно быть 0, это каскад) err=%v\n", len(hist), err)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := "postgres://postgres:mypassword@localhost:5432/productdb"
	pool, err := pg.CreatePool(ctx, dsn)
	if err != nil {
		fmt.Println("pool:", err)
		return
	}
	defer pool.Close()

	repo := pg.NewUserRepo(pool)
	name := fmt.Sprintf("scratch_%d", time.Now().UnixNano())

	u, err := repo.Create(ctx, name, "hash")
	fmt.Printf("1 create: %+v err=%v\n", u, err)

	_, err = repo.Create(ctx, name, "hash")
	fmt.Println("2 duplicate -> ErrAlreadyExists:", errors.Is(err, domain.ErrAlreadyExists))

	u, err = repo.GetByName(ctx, name)
	fmt.Printf("3 get by name: %+v err=%v", u, err)

	_, err = repo.GetByName(ctx, "no_such_user")
	fmt.Println("4 unknown name -> ErrNotFound:", errors.Is(err, domain.ErrNotFound))

	testCrypto(ctx, pool)
	testAuth(ctx, pool)
}
