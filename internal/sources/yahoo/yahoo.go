package yahoo

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// YahooClient hämtar DAGENS KURS från Yahoo Finance.
// Används ENDAST för prisdata — inga fundamentals.
type YahooClient struct {
	baseURL string
}

func New() *YahooClient {
	return &YahooClient{
		baseURL: "https://query2.finance.yahoo.com",
	}
}

func (c *YahooClient) Name() string { return "yahoo" }

func (c *YahooClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// Yahoo har inget bolagsregister — hoppa över
	return nil, fmt.Errorf("yahoo har inget bolagsregister")
}

func (c *YahooClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// Yahoo används INTE för fundamentals
	return nil, fmt.Errorf("yahoo används inte för fundamentals")
}

func (c *YahooClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

// FetchQuote hämtar dagens kurs för ett bolag.
// Returnerar: pris, valuta, market cap, volym, dagens förändring.
func (c *YahooClient) FetchQuote(ctx context.Context, ticker string) (*PriceQuote, error) {
	// TODO: GET /v6/finance/quote?symbols={ticker}
	// fields: regularMarketPrice, currency, marketCap, volume, regularMarketChangePercent
	return nil, fmt.Errorf("inte implementerad ännen")
}

// FetchHistoricalPrices hämtar historisk kursdata.
func (c *YahooClient) FetchHistoricalPrices(ctx context.Context, ticker, period string) ([]DailyPrice, error) {
	// TODO: GET /v8/finance/chart/{ticker}?range=1d&interval=1d
	return nil, fmt.Errorf("inte implementerad ännen")
}

// PriceQuote är dagens kursdata.
type PriceQuote struct {
	Ticker         string  `json:"ticker"`
	Price          float64 `json:"price"`
	Currency       string  `json:"currency"`
	MarketCap      int64   `json:"market_cap"`
	Volume         int64   `json:"volume"`
	ChangePercent  float64 `json:"change_percent"`
	LastUpdated    string  `json:"last_updated"`
}

// DailyPrice är daglig kursdata.
type DailyPrice struct {
	Ticker string  `json:"ticker"`
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}
