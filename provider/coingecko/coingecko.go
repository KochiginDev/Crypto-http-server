package coingecko

import (
	"context"
	"crypthttp/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	client  *http.Client
	baseURL *url.URL
	key     string
}

func NewClient(client *http.Client, addr string, key string) (*Client, error) {
	baseURL, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf(`url.Parse(%q): %w`, addr, err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf(`baseURL.Scheme must be "http" or "https"`)
	}
	if baseURL.Host == "" {
		return nil, fmt.Errorf(`baseURL "%q" must have a host`, addr)
	}
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	return &Client{client: client,
		baseURL: baseURL,
		key:     key}, nil
}
func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	u := c.baseURL.JoinPath(path)
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf(`http.NewRequest(%q): %w`, u.String(), err)
	}
	req.Header.Set("Accept", "application/json")
	if c.key != "" {
		req.Header.Set("x-cg-demo-api-key", c.key)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf(`ctx.Err(): %w`, ctx.Err())
		}
		return fmt.Errorf("%w: %w", domain.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()
	rbody := io.LimitReader(resp.Body, 1<<20)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", domain.ErrProviderUnavailable, resp.Status)
	}
	err = json.NewDecoder(rbody).Decode(out)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrProviderUnavailable, err)
	}
	return nil
}

type searchCoin struct {
	Id     string `json:"id"`
	Symbol string `json:"symbol"`
	Rank   *int   `json:"market_cap_rank"`
	Name   string `json:"name"`
}
type searchResponse struct {
	Coins []searchCoin `json:"coins"`
}

func (c *Client) best(c1 searchCoin, c2 searchCoin) bool {
	if c1.Rank == nil && c2.Rank == nil {
		return false
	}
	if c2.Rank == nil {
		return false
	}
	if c1.Rank == nil {
		return true
	}
	if *c2.Rank > 0 && *c2.Rank < *c1.Rank {
		return true
	}
	return false
}
func (c *Client) Resolve(ctx context.Context, symbol string) (string, string, error) {
	path := "search"
	query := url.Values{
		"query": {symbol},
	}
	var out searchResponse
	var bestCoin searchCoin
	err := c.get(ctx, path, query, &out)
	if err != nil {
		return "", "", err
	}
	flag := false
	for _, coin := range out.Coins {
		if !strings.EqualFold(coin.Symbol, symbol) {
			continue
		}

		if c.best(bestCoin, coin) || !flag {
			bestCoin = coin
		}
		flag = true
	}
	if !flag {
		return "", "", domain.ErrUnknownSymbol
	}
	return bestCoin.Id, bestCoin.Name, nil
}

func (c *Client) GetPrice(ctx context.Context, coinGeckoID string) (float64, error) {
	path := "simple/price"
	query := url.Values{
		"ids":           {coinGeckoID},
		"vs_currencies": {"usd"},
	}
	var out map[string]map[string]float64
	err := c.get(ctx, path, query, &out)
	if err != nil {
		return 0, err
	}
	v, ok := out[coinGeckoID]
	if !ok {
		return 0, domain.ErrUnknownSymbol
	}
	price, ok := v["usd"]
	if !ok {
		return 0, domain.ErrUnknownSymbol
	}
	return price, nil
}
