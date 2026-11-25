package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"crypthttp/domain"
	"crypthttp/provider/coingecko"
)

const baseURL = "https://api.coingecko.com/api/v3"

// kind переводит ошибку в понятную метку, чтобы видеть не только текст.
func kind(err error) string {
	switch {
	case err == nil:
		return "nil"
	case errors.Is(err, domain.ErrUnknownSymbol):
		return "ErrUnknownSymbol"
	case errors.Is(err, domain.ErrProviderUnavailable):
		return "ErrProviderUnavailable"
	default:
		return "другая ошибка"
	}
}

func verdict(ok bool) string {
	if ok {
		return "OK  "
	}
	return "FAIL"
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	key := os.Getenv("COINGECKO_API_KEY")

	c, err := coingecko.NewClient(nil, baseURL, key)
	if err != nil {
		fmt.Println("NewClient:", err)
		os.Exit(1)
	}

	// Бесплатный тариф ограничен по числу запросов в минуту,
	// поэтому между вызовами небольшая пауза.
	pause := func() { time.Sleep(1500 * time.Millisecond) }

	// 1. BTC -> bitcoin (проверяет выбор по рангу)
	id, name, err := c.Resolve(ctx, "BTC")
	fmt.Printf("%s 1 Resolve(BTC): id=%q name=%q err=%v [%s]\n",
		verdict(err == nil && id == "bitcoin"), id, name, err, kind(err))
	pause()

	// 2. ETH -> ethereum
	id, name, err = c.Resolve(ctx, "ETH")
	fmt.Printf("%s 2 Resolve(ETH): id=%q name=%q err=%v [%s]\n",
		verdict(err == nil && id == "ethereum"), id, name, err, kind(err))
	pause()

	// 3. USDT -> tether (у тикера много «двойников»)
	id, name, err = c.Resolve(ctx, "USDT")
	fmt.Printf("%s 3 Resolve(USDT): id=%q name=%q err=%v [%s]\n",
		verdict(err == nil && id == "tether"), id, name, err, kind(err))
	pause()

	// 4. Несуществующий тикер -> ErrUnknownSymbol
	id, name, err = c.Resolve(ctx, "ZZZZZZ")
	fmt.Printf("%s 4 Resolve(ZZZZZZ): id=%q name=%q err=%v [%s]\n",
		verdict(errors.Is(err, domain.ErrUnknownSymbol)), id, name, err, kind(err))
	pause()

	// 5. Цена bitcoin: сравни порядок числа с сайтом CoinGecko
	price, err := c.GetPrice(ctx, "bitcoin")
	fmt.Printf("%s 5 GetPrice(bitcoin): price=%v err=%v [%s]\n",
		verdict(err == nil && price > 0), price, err, kind(err))
	pause()

	// 6. Выдуманный id -> ErrUnknownSymbol, а не цена 0
	price, err = c.GetPrice(ctx, "not-a-real-coin")
	fmt.Printf("%s 6 GetPrice(not-a-real-coin): price=%v err=%v [%s]\n",
		verdict(errors.Is(err, domain.ErrUnknownSymbol)), price, err, kind(err))

	// 7. Адрес, где никто не слушает -> ErrProviderUnavailable
	dead, err := coingecko.NewClient(nil, "http://localhost:1", "")
	if err != nil {
		fmt.Println("NewClient(dead):", err)
		os.Exit(1)
	}
	_, _, err = dead.Resolve(ctx, "BTC")
	fmt.Printf("%s 7 Resolve на мёртвом адресе: err=%v [%s]\n",
		verdict(errors.Is(err, domain.ErrProviderUnavailable)), err, kind(err))

	// 8. Слишком маленький таймаут -> ErrProviderUnavailable
	fast, err := coingecko.NewClient(&http.Client{Timeout: time.Millisecond}, baseURL, key)
	if err != nil {
		fmt.Println("NewClient(fast):", err)
		os.Exit(1)
	}
	_, err = fast.GetPrice(ctx, "bitcoin")
	fmt.Printf("%s 8 GetPrice с таймаутом 1мс: err=%v [%s]\n",
		verdict(errors.Is(err, domain.ErrProviderUnavailable)), err, kind(err))

	// 9. Отменённый контекст -> НЕ ErrProviderUnavailable (это отмена вызывающим)
	cctx, ccancel := context.WithCancel(context.Background())
	ccancel()
	_, err = c.GetPrice(cctx, "bitcoin")
	fmt.Printf("%s 9 GetPrice с отменённым ctx: err=%v [%s]\n",
		verdict(errors.Is(err, context.Canceled) && !errors.Is(err, domain.ErrProviderUnavailable)),
		err, kind(err))
}
