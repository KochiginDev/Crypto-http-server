package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"crypthttp/domain"
	"crypthttp/repository/pg"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	token := "tok_" + name

	u, err := repo.Create(ctx, name, "hash")
	fmt.Printf("1 create: %+v err=%v\n", u, err)

	_, err = repo.Create(ctx, name, "hash")
	fmt.Println("2 duplicate -> ErrAlreadyExists:", errors.Is(err, domain.ErrAlreadyExists))

	u, err = repo.GetByName(ctx, name)
	fmt.Printf("3 get by name: %+v err=%v, token empty: %v\n", u, err, u.Token == "")

	_, err = repo.GetByName(ctx, "no_such_user")
	fmt.Println("4 unknown name -> ErrNotFound:", errors.Is(err, domain.ErrNotFound))

	err = repo.SaveNewToken(ctx, u.Id, token)
	fmt.Println("5 save token err:", err)

	u, err = repo.GetByToken(ctx, token)
	fmt.Printf("6 get by token: %+v err=%v\n", u, err)

	_, err = repo.GetByToken(ctx, "wrong")
	fmt.Println("7 unknown token -> ErrNotFound:", errors.Is(err, domain.ErrNotFound))

	err = repo.SaveNewToken(ctx, 999999, "x")
	fmt.Println("8 unknown id -> ErrNotFound:", errors.Is(err, domain.ErrNotFound))
}
