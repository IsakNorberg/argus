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
	Produktionsenhed []any          `json:"Produktionsenhed"`
}

// FetchCompanies searches CVR by common business form terms.
func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	var allCompanies []sources.Company

	for start := 0; start < 10000; start += 100 {
		companies, hasMore, err := c.fetchPage(ctx, start)
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

func (c *Client) fetchPage(ctx context.Context, start int) ([]sources.Company, bool, error) {
	c.rateLimit()
	urlStr := fmt.Sprintf("https://datacvr.virk.dk/api/v2/virksomhed?q=%s&sortering=navn&start=%d", "Virksomhedsstatus:Normal", start)

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, false, fmt.Errorf("bygge request: %w", err)
	}
	req.Header.Set("User-Agent", "Argus")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("hente CVR: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("læse body: %w", err)
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
			Exchange:   "CSE",
		})
	}

	return companies, len(data.Virksomheder) >= 100, nil
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	c.rateLimit()
	urlStr := fmt.Sprintf("https://datacvr.virk.dk/api/v2/virksomhed?cvrNummer=%s", companyID)

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("bygge request: %w", err)
	}
	req.Header.Set("User-Agent", "Argus")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hente profil: %w", err)
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
		return nil, fmt.Errorf("ingen profil fundet for CVR %s", companyID)
	}

	v := data.Virksomheder[0]
	return &sources.Profile{
		CompanyID:   fmt.Sprintf("%d", v.CVRNummer),
		Description: fmt.Sprintf("%s — Danish company", v.Virksomhedsnavn),
		Country:     "DK",
		Exchange:    "CSE",
	}, nil
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// Get profile first (financials are embedded)
	c.rateLimit()
	urlStr := fmt.Sprintf("https://datacvr.virk.dk/api/v2/virksomhed?cvrNummer=%s", companyID)

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("bygge request: %w", err)
	}
	req.Header.Set("User-Agent", "Argus")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hente finansiel: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cvr finansiel HTTP %d", resp.StatusCode)
	}

	var data cvrResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode finansiel: %w", err)
	}

	if len(data.Virksomheder) == 0 {
		return nil, fmt.Errorf("ingen data for CVR %s", companyID)
	}

	v := data.Virksomheder[0]
	// Financials are in metadata — would need to parse Regnskab array
	// Complex structure with nested field codes — implement later
	_ = v

	return nil, fmt.Errorf("finansiel data for CVR kræver dybere parsing — kilde findes")
}

func (c *Client) rateLimit() {
	delay := time.Since(c.lastCall)
	if delay < 100*time.Millisecond {
		time.Sleep(100*time.Millisecond - delay)
	}
	c.lastCall = time.Now()
}
