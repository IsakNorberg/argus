package cvr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

// Client for Danmarks CVR / Virksomhedsguiden (v2 API)
// Open data, no API key required.
type Client struct {
	client   *http.Client
	lastCall time.Time
}

func New() *Client {
	return &Client{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "cvr" }

// cvrResponse from search endpoint
type cvrResponse struct {
	Hits         int           `json:"hits"`
	Virksomheder []virksomhed  `json:"virksomheder"`
}

type virksomhed struct {
	CVRNummer        int            `json:"CVRNummer"`
	Virksomhedsnavn  string         `json:"Virksomhedsnavn"`
	Virksomhedsstatus string        `json:"virksomhedsstatus"`
	Adresse          *cvrAdresse    `json:"Adresse,omitempty"`
	Produktionsenhed []any          `json:"Produktionsenhed"`
}

type cvrAdresse struct {
	Vejnavn    string `json:"vejnavn"`
	Husnummer  string `json:"husnummer"`
	Postnummer struct {
		Nummer string `json:"nummer"`
	} `json:"postnummer,omitempty"`
}

// FetchCompanies searches CVR by common business form terms.
// CVR does not have "get all companies" — we paginate broad searches.
func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	var allCompanies []sources.Company

	// Broad searches for Danish company types
	for start := 0; start < 10000; start += 100 {
		c.rateLimit()
		urlStr := fmt.Sprintf("https://datacvr.virk.dk/api/v2/virksomhed?q=%s&sortering=navn&start=%d", "Virksomhedsstatus:Normal", start)
		companies, hasMore, err := c.fetchPage(ctx, urlStr)
		if err != nil {
			return allCompanies, err
		}
		allCompanies = append(allCompanies, companies...)
		if !hasMore || len(companies) == 0 {
			break
		}
	}

	return allCompanies, nil
}

func (c *Client) fetchPage(ctx context.Context, urlStr string) ([]sources.Company, bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, false, fmt.Errorf("bygg request: %w", err)
	}
	req.Header.Set("User-Agent", "Argus (argus@isaknorberg.dev)")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("hämta CVR: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("läs body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("cvr http %d: %s", resp.StatusCode, body)
	}

	var data cvrResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, false, fmt.Errorf("decode CVR: %w", err)
	}

	var companies []sources.Company
	for _, v := range data.Virksomheder {
		if v.Virksomhedsnavn == "" || v.CVRNummer == 0 {
			continue
		}
		companies = append(companies, sources.Company{
			ExternalID: fmt.Sprintf("%d", v.CVRNummer),
			Name:       v.Virksomhedsnavn,
			Country:    "DK",
			Exchange:   "CSE", // Copenhagen Stock Exchange
		})
	}

	return companies, len(data.Virksomheder) >= 100, nil
}

// FetchProfile gets company profile by CVR number
func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	c.rateLimit()
	// Search for specific CVR number
	urlStr := fmt.Sprintf("https://datacvr.virk.dk/api/v2/virksomhed?cvrNummer=%s", companyID)

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("bygg request: %w", err)
	}
	req.Header.Set("User-Agent", "Argus (argus@isaknorberg.dev)")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hämta profil: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cvr profil HTTP %d", resp.StatusCode)
	}

	var data cvrResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode profil: %w", err)
	}

	if len(data.Virksomheder) == 0 {
		return nil, fmt.Errorf("ingen profil hittad för CVR %s", companyID)
	}

	v := data.Virksomheder[0]
	return &sources.Profile{
		CompanyID:   fmt.Sprintf("%d", v.CVRNummer),
		Description: fmt.Sprintf("%s — Danish company", v.Virksomhedsnavn),
		Country:     "DK",
		Exchange:    "CSE",
	}, nil
}

// FetchFinancials — CVR embeds financials in company metadata.
func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// Get profile first (financials are embedded)
	c.rateLimit()
	urlStr := fmt.Sprintf("https://datacvr.virk.dk/api/v2/virksomhed?cvrNummer=%s", companyID)

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("bygg request: %w", err)
	}
	req.Header.Set("User-Agent", "Argus (argus@isaknorberg.dev)")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hämta finansiell: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cvr finansiell HTTP %d", resp.StatusCode)
	}

	var data cvrResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode finansiell: %w", err)
	}

	if len(data.Virksomheder) == 0 {
		return nil, fmt.Errorf("ingen data för CVR %s", companyID)
	}

	v := data.Virksomheder[0]
	// Financials are in metadata — parse from regnskab field
	// TODO: Parse actual CVR financial metadata structure
	_ = v // Used when implementing full CVR financial metadata parsing

	return nil, fmt.Errorf("finansiell data för CVR kräver djupare parsing — källa finns")
}

func (c *Client) rateLimit() {
	delay := time.Since(c.lastCall)
	if delay < 100*time.Millisecond {
		time.Sleep(100*time.Millisecond - delay)
	}
	c.lastCall = time.Now()
}
