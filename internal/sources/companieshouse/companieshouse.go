package companieshouse

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

type Client struct {
	client   *http.Client
	apiKey   string
	lastCall time.Time
}

func New() *Client {
	return &Client{
		client: &http.Client{Timeout: 30 * time.Second},
		apiKey: os.Getenv("COMPANIES_HOUSE_API_KEY"),
	}
}

func (c *Client) Name() string { return "companieshouse" }

type searchResponse struct {
	TotalResults int `json:"total_results"`
	Items        []struct {
		CompanyNumber string `json:"company_number"`
		Title         string `json:"title"`
		DateOfCreation string `json:"date_of_creation"`
		CompanyStatus string `json:"company_status"`
		Address       struct {
			Locality      string `json:"locality"`
			Premises      string `json:"premises"`
			PostalCode    string `json:"postal_code"`
			AddressLine1  string `json:"address_line_1"`
			Region        string `json:"region"`
		} `json:"address"`
	} `json:"items"`
}

type companyProfile struct {
	CompanyName       string `json:"company_name"`
	CompanyNumber     string `json:"company_number"`
	Category          string `json:"type"`
	DateOfCreation    string `json:"date_of_creation"`
	CompanyStatus     string `json:"company_status"`
	SicCodes          []string `json:"sic_codes"`
	Accounts          struct {
		NextDue     string `json:"next_due"`
		NextMadeUpTo string `json:"next_made_up_to"`
	} `json:"accounts"`
	RegisteredOfficeAddress struct {
		Locality   string `json:"locality"`
		Country    string `json:"country"`
		PostalCode string `json:"postal_code"`
	} `json:"registered_office_address"`
}

type filingHistory struct {
	TotalCount int `json:"total_count"`
	FilingHistory []struct {
		Category   string `json:"category"`
		Date       string `json:"date"`
		Description string `json:"description"`
		Resolutions []string `json:"resolutions"`
		Type       string `json:"type"`
	} `json:"items"`
}

func (c *Client) authHeader() string {
	if c.apiKey != "" {
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(c.apiKey+":"))
	}
	return ""
}

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// Companies House har ingen "alla bolag"-endpoint.
	// Söker med breda PLC/LTD termer
	searches := []string{"plc", "ltd", "public limited", "group"}
	var allCompanies []sources.Company

	for _, query := range searches {
		for start := 0; start < 20; start++ {
			c.rateLimit()
			url := fmt.Sprintf("https://api.company-information.service.gov.uk/search/companies?q=%s&start_index=%d&items_per_page=20&order=incorporation_date", query, start*20)

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return allCompanies, fmt.Errorf("bygg request: %w", err)
			}
			req.Header.Set("Accept", "application/json")
			if auth := c.authHeader(); auth != "" {
				req.Header.Set("Authorization", auth)
			}

			resp, err := c.client.Do(req)
			if err != nil {
				return allCompanies, fmt.Errorf("hämta CompaniesHouse: %w", err)
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return allCompanies, fmt.Errorf("läs body: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				return allCompanies, fmt.Errorf("companieshouse HTTP %d: %s", resp.StatusCode, body)
			}

			var data searchResponse
			if err := json.Unmarshal(body, &data); err != nil {
				return allCompanies, fmt.Errorf("decode CompaniesHouse: %w", err)
			}

			for _, item := range data.Items {
				if item.Title == "" || item.CompanyNumber == "" {
					continue
				}
				allCompanies = append(allCompanies, sources.Company{
					Name:       item.Title,
					Country:    "GB",
					ExternalID: item.CompanyNumber,
					Exchange:   "LSE",
					Industry:   "",
				})
			}

			if len(data.Items) < 20 {
				break
			}
		}
	}

	return allCompanies, nil
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	c.rateLimit()
	url := fmt.Sprintf("https://api.company-information.service.gov.uk/company/%s", companyID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("bygg request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if auth := c.authHeader(); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hämta profil: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("companieshouse profil HTTP %d", resp.StatusCode)
	}

	var data companyProfile
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode profil: %w", err)
	}

	sector := "Unknown"
	if len(data.SicCodes) > 0 {
		sector = data.SicCodes[0]
	}

	category := "Unknown"
	if data.Category != "" {
		category = data.Category
	}

	return &sources.Profile{
		CompanyID:   data.CompanyNumber,
		Description: fmt.Sprintf("%s is a UK %s company", data.CompanyName, category),
		Sector:      sector,
		Country:     "GB",
		Exchange:    "LSE",
	}, nil
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	c.rateLimit()
	url := fmt.Sprintf("https://api.company-information.service.gov.uk/company/%s/filing-history?type=AA&category=accounts&items_per_page=50", companyID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("bygg request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if auth := c.authHeader(); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hämta filing-history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("companieshouse filing HTTP %d", resp.StatusCode)
	}

	var data filingHistory
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode filing: %w", err)
	}

	var financials []sources.Financials
	for _, f := range data.FilingHistory {
		if f.Date == "" {
			continue
		}
		fin := sources.Financials{
			CompanyID:  companyID,
			Period:     f.Date[:4] + "FY",
			Currency:   "GBP",
			ReportDate: f.Date,
			FilingDate: f.Date,
			Source:     "companieshouse",
		}
		financials = append(financials, fin)
	}

	if len(financials) == 0 {
		return nil, fmt.Errorf("ingen finansiell data hittad")
	}

	return financials, nil
}

func (c *Client) rateLimit() {
	delay := time.Since(c.lastCall)
	// API limit: 600 req/min = 10/sec → 100ms mellan requests
	if delay < 100*time.Millisecond {
		time.Sleep(100*time.Millisecond - delay)
	}
	c.lastCall = time.Now()
}
