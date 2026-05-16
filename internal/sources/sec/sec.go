package sec

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// SECClient hämtar data från SEC EDGAR API.
type SECClient struct {
	baseURL string
	email   string // User-Agent krävs av SEC
}

func New(email string) *SECClient {
	return &SECClient{
		baseURL: "https://data.sec.gov",
		email:   email,
	}
}

func (c *SECClient) Name() string { return "sec" }

func (c *SECClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// TODO: Hämta CIK-listan från https://data.sec.gov/submissions/
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *SECClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// TODO: Hämta company facts från https://data.sec.gov/submissions/CIK{cik}.json
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *SECClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	// TODO: Hämta bolagsprofil — finns i SEC filing XML
	return nil, fmt.Errorf("inte implementerad ännen")
}
