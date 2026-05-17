package brreg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

type Client struct {
	client   *http.Client
	lastCall time.Time
}

func New() *Client {
	return &Client{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Name() string { return "brreg" }

// Enhet från Brreg API
type enhetResponse struct {
	Links struct {
		First struct{ Href string } `json:"first"`
		Last  struct{ Href string } `json:"last"`
		Next  struct{ Href string } `json:"next"`
	} `json:"_links"`
	Page struct {
		Size         int `json:"size"`
		TotalElements int `json:"totalElements"`
		TotalPages   int `json:"totalPages"`
		Number       int `json:"number"`
	} `json:"page"`
	Embedded struct {
		Enheter []enhet `json:"enheter"`
	} `json:"_embedded"`
}

type enhet struct {
	Organisasjonsnummer  string     `json:"organisasjonsnummer"`
	Navn                 string     `json:"navn"`
	Organisasjonsform    orgForm    `json:"organisasjonsform"`
	Forretningsadresse   adresse    `json:"forretningsadresse"`
	Naeringskode         []naering  `json:"naeringskode"`
	Registreringsdato    string     `json:"stiftelsesdato,omitempty"`
	Opphoert             string     `json:"opphoert,omitempty"`
	AntallAnsatte        int        `json:"antallAnsatte"`
}

type orgForm struct {
	Kode        string `json:"kode"`
	Beskrivelse string `json:"beskrivelse"`
}

type adresse struct {
	Kommune string `json:"kommune"`
	Land    string `json:"land"`
}

type naering struct {
	Kode        string `json:"kode"`
	Beskrivelse string `json:"beskrivelse"`
}

// Regnskap från Brreg regnskapsregister
type regnskap struct {
	Organisasjonsnummer string  `json:"organisasjonsnummer"`
	Regnskapsdato       string  `json:"regnskapsdato"`
	Regnskapsart        string  `json:"regnskapsart"`
	Driftsinntekter     float64 `json:"driftsinntekter,omitempty"`
	Balansesum          float64 `json:"balansesum,omitempty"`
}

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	var allCompanies []sources.Company
	page := 0

	for {
		c.rateLimit()
		urlStr := fmt.Sprintf("https://data.brreg.no/enhetsregisteret/api/enheter?page=%d&size=100", page)

		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return allCompanies, fmt.Errorf("bygg request: %w", err)
		}
		req.Header.Set("User-Agent", "Argus (argus@isaknorberg.dev)")

		resp, err := c.client.Do(req)
		if err != nil {
			return allCompanies, fmt.Errorf("hämta Brreg: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return allCompanies, fmt.Errorf("läs body: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			break // Ingen fler sidor
		}

		var data enhetResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return allCompanies, fmt.Errorf("decode Brreg: %w", err)
		}

		for _, e := range data.Embedded.Enheter {
			industry := ""
			if len(e.Naeringskode) > 0 {
				industry = e.Naeringskode[0].Beskrivelse
			}
			company := sources.Company{
				Name:       e.Navn,
				Country:    "NO",
				ExternalID: e.Organisasjonsnummer,
				Industry:   industry,
				Exchange:   "OSE", // Oslo Børs
			}
			allCompanies = append(allCompanies, company)
		}

		if len(data.Embedded.Enheter) < 100 || page >= data.Page.TotalPages-1 {
			break
		}
		page++
	}

	return allCompanies, nil
}

func (c *Client) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	c.rateLimit()
	urlStr := fmt.Sprintf("https://data.brreg.no/enhetsregisteret/api/enheter/%s", companyID)

	resp, err := c.client.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("hämta profil: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Brreg profil HTTP %d", resp.StatusCode)
	}

	var data enhet
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode profil: %w", err)
	}

	sector := ""
	if len(data.Naeringskode) > 0 {
		sector = data.Naeringskode[0].Beskrivelse
	}

	return &sources.Profile{
		CompanyID:   data.Organisasjonsnummer,
		Description: fmt.Sprintf("%s is a Norwegian %s", data.Navn, data.Organisasjonsform.Beskrivelse),
		Sector:      sector,
		Industry:    "",
		Country:     "NO",
		Exchange:    "OSE",
		Employees:   data.AntallAnsatte,
	}, nil
}

func (c *Client) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	c.rateLimit()
	urlStr := fmt.Sprintf("https://data.brreg.no/regnskapsregisteret/api/regnskap?organisasjonsnummer=%s", url.QueryEscape(companyID))

	resp, err := c.client.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("hämta regnskap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Brreg regnskap HTTP %d", resp.StatusCode)
	}

	// Regnskaps API returnerar direkt array
	var regnskapList []regnskap
	if err := json.NewDecoder(resp.Body).Decode(&regnskapList); err != nil {
		return nil, fmt.Errorf("decode regnskap: %w", err)
	}

	var financials []sources.Financials
	for _, r := range regnskapList {
		period := ""
		if r.Regnskapsdato != "" {
			period = r.Regnskapsdato[:4] + "FY"
		}
		fin := sources.Financials{
			CompanyID:  r.Organisasjonsnummer,
			Period:     period,
			Currency:   "NOK",
			Revenue:    r.Driftsinntekter,
			TotalAssets: r.Balansesum,
			ReportDate: r.Regnskapsdato,
			FilingDate: r.Regnskapsdato,
			Source:     "brreg",
		}
		financials = append(financials, fin)
	}

	return financials, nil
}

func (c *Client) rateLimit() {
	delay := time.Since(c.lastCall)
	if delay < 16*time.Millisecond { // ~60/min max
		time.Sleep(16*time.Millisecond - delay)
	}
	c.lastCall = time.Now()
}
