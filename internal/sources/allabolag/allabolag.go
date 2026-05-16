package allabolag

import (
	"context"
	"fmt"

	"github.com/IsakNorberg/argus/internal/sources"
)

// AllabolagClient scrapar Allabolag.se.
type AllabolagClient struct {
	baseURL string
}

func New() *AllabolagClient {
	return &AllabolagClient{
		baseURL: "https://www.allabolag.se",
	}
}

func (c *AllabolagClient) Name() string { return "allabolag" }

func (c *AllabolagClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// TODO: Scrapa bolagslistor eller sökfunktion
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *AllabolagClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// TODO: Scrapa årsredovisningar från bolagssidor
	return nil, fmt.Errorf("inte implementerad ännen")
}

func (c *AllabolagClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	// TODO: Scrapa bolagsfakta (VD, styrelse, adress)
	return nil, fmt.Errorf("inte implementerad ännen")
}
