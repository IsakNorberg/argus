package yahoo

import (
	"context"
	"fmt"
)

type PriceQuote struct {
	Ticker        string
	Price         float64
	Currency      string
	MarketCap     int64
	Volume        int64
	ChangePercent float64
}

type Client struct {
	baseURL string
}

func New() *Client {
	return &Client{
		baseURL: "https://query1.finance.yahoo.com/v8/finance/chart",
	}
}

func (c *Client) FetchQuote(ctx context.Context, ticker string) (*PriceQuote, error) {
	// TODO: Hämta från Yahoo Finance API
	return nil, fmt.Errorf("ej implementerad ännen")
}
