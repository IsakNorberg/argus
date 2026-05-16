package yahoo

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// YahooClient hämtar data från Yahoo Finance.
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
	// TODO: Yahoo har inget sök-API direkt
	// Använd symbols från en extern lista
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *YahooClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// TODO: /v10/finance/quoteSummary/{symbol}?modules=incomeStatementHistory,balanceSheetHistory,cashflowStatementHistory
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *YahooClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	// TODO: /v10/finance/quoteSummary/{symbol}?modules=assetProfile
	return nil, fmt.Errorf("inte implementerad ännen")
}
