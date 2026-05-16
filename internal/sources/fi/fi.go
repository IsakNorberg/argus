package fi

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// FIClient hämtar data från Finansinspeksen.
type FIClient struct {
	baseURL string
}

func New() *FIClient {
	return &FIClient{
		baseURL: "https://fi.se/oppnadata/",
	}
}

func (c *FIClient) Name() string { return "fi" }

func (c *FIClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// TODO: Registrerade bolag på svenska börsen
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *FIClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *FIClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	return nil, fmt.Errorf("inte implementerad ännen")
}
