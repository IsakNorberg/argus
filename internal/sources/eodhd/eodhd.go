package eodhd

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// EODHDClient hämtar data från EODHD API.
type EODHDClient struct {
	apiKey  string
	baseURL string
}

func New(apiKey string) *EODHDClient {
	return &EODHDClient{
		apiKey:  apiKey,
		baseURL: "https://eodhd.com/api",
	}
}

func (c *EODHDClient) Name() string { return "eodhd" }

func (c *EODHDClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// TODO: /api/screener?api_token=TOKEN
	// Returnerar alla bolag med filtrering
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *EODHDClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// TODO: /api/fundamentals/{symbol}?api_token=TOKEN
	// Innehåller income_statement, balance_sheet, cash_flow
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *EODHDClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	// TODO: Samma fundamentals-endpoint har General-fältet
	return nil, fmt.Errorf("inte implementerad ännen")
}
