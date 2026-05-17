package sedar

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IsakNorberg/argus/internal/sources"
)

const (
	baseURL        = "https://www.sedarplus.ca"
	xmlURL         = "https://www.sedar.com/issuers/issuers_en.xml"
	userAgent      = "Argus (argus@isaknorberg.dev)"
	rateDelay      = 500 * time.Millisecond // 2 req/s - konservativ för SEDAR
)

// SEDARClient hämtar kanadensiska bolag och filings från SEDAR/SEDAR+.
type SEDARClient struct {
	client   *http.Client
	lastCall time.Time
}

func New() *SEDARClient {
	return &SEDARClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *SEDARClient) Name() string { return "sedar" }

// ─── HTTP helper ─────────────────────────────────────────────────────

func (c *SEDARClient) do(ctx context.Context, url string, out any, xmlParse bool) error {
	if err := c.rateLimit(ctx); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("sedar: skapa request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/xml, text/xml, */*")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("sedar: hämta %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sedar: %d %s: %s", resp.StatusCode, resp.Status, string(body))
	}

	if xmlParse && out != nil {
		if err := xml.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("sedar: parse XML: %w", err)
		}
	}
	return nil
}

// ─── Issuer XML-strukturer ───────────────────────────────────────────

// issuerFile representerar rot-elementet i issuer XML-filen.
type issuerFile struct {
	XMLName xml.Name `xml:"issuers"`
	Issuers []issuer `xml:"issuer"`
}

type issuer struct {
	CompanyName     string `xml:"company_name"`
	CDSNumber       string `xml:"cds_number"`
	SEDARSymbol     string `xml:"sedar_symbol"`
	SEDARNumbers    []string `xml:"sedar_number"`
	LanguageOption  string `xml:"language_option"`
}

// ─── Bolagssökning ───────────────────────────────────────────────────

// cachedIssuers håller den parsade XML-listan i minnet.
var cachedIssuers []sources.Company

// FetchCompanies hämtar alla kanadensiska bolag från SEDARs XML-issuers-lista.
// SEDAR har inget offentligt REST-API — data distribueras som en XML-fil med alla
// registrerade emittenter.
func (c *SEDARClient) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// Om vi redan har cachat, returnera
	if len(cachedIssuers) > 0 {
		return cachedIssuers, nil
	}

	var issuers issuerFile
	if err := c.do(ctx, xmlURL, &issuers, true); err != nil {
		return nil, fmt.Errorf("sedar: hämta issuer XML: %w", err)
	}

	var companies []sources.Company
	for _, iss := range issuers.Issuers {
		symbol := iss.SEDARSymbol
		if symbol == "" && len(iss.SEDARNumbers) > 0 {
			symbol = iss.SEDARNumbers[0]
		}

		cdNum := iss.CDSNumber
		if cdNum == "" && len(iss.SEDARNumbers) > 0 {
			cdNum = iss.SEDARNumbers[0]
		}

		companies = append(companies, sources.Company{
			Name:       iss.CompanyName,
			ExternalID: cdNum,
			Ticker:     symbol,
			Country:    "CA",
			Exchange:   "TSX",
			Industry:   "", // SEDAR XML innehåller ingen branschinfo
		})
	}

	cachedIssuers = companies
	return companies, nil
}

// ─── Bolagsprofil ────────────────────────────────────────────────────

// SEDAR har ingen specifik profil-endpoint. Profiler extraheras från
// XML-filen och kompletteras med filing-metadata.
func (c *SEDARClient) FetchProfile(ctx context.Context, companyID string) (*sources.Profile, error) {
	// Sök i cached data
	if len(cachedIssuers) == 0 {
		_, err := c.FetchCompanies(ctx)
		if err != nil {
			return nil, err
		}
	}

	var found *sources.Company
	for i := range cachedIssuers {
		if cachedIssuers[i].ExternalID == companyID || cachedIssuers[i].Ticker == companyID {
			found = &cachedIssuers[i]
			break
		}
	}

	if found == nil {
		return nil, fmt.Errorf("sedar: bolag %s inte funnet i issuer-listan", companyID)
	}

	return &sources.Profile{
		CompanyID:   found.ExternalID,
		Description: fmt.Sprintf("%s (%s) — Kanadensiskt bolag registrerat i SEDAR", found.Name, found.Ticker),
		Sector:      "",
		Industry:    found.Industry,
		Country:     "CA",
		Exchange:    "TSX",
	}, nil
}

// ─── Finansiell data ─────────────────────────────────────────────────

// SEDAR har inte något offentligt REST-API för att hämta finansiella siffror
// direkt. Filings måste hämtas via SEDAR+ webbgränssnittet som kräver
// JavaScript-rendering (PerimeterX bot-skydd).
//
// Denna implementation returnerar en instruktion om att använda SEDAR+
// webbgränssnitt manuellt eller via deras betalda API-tjänst.
func (c *SEDARClient) FetchFinancials(ctx context.Context, companyID string) ([]sources.Financials, error) {
	// SEDAR+ använder PerimeterX bot-skydd som blockerar automatiserade anrop.
	// De finansiella filingarna finns på:
	// https://www.sedarplus.ca/filing-profile?lang=en&company_id={id}
	// Men kräver JavaScript-rendering.
	//
	// Alternativ: Använd SEDARs betalda datafeed eller hämta XBRL filings
	// från SEDAR+ manuellt.
	return nil, fmt.Errorf(
		"sedar: finansiell data kräver SEDAR+ webgränssnitt " +
			"(PerimeterX-bot-skydd). Besök %s/filing-profile för manuella " +
			"uppslag, eller använd SEDARs betalda datafeed", baseURL)
}

// ─── Rate limiting ───────────────────────────────────────────────────

func (c *SEDARClient) rateLimit(ctx context.Context) error {
	now := time.Now()
	elapsed := now.Sub(c.lastCall)
	if elapsed < rateDelay {
		select {
		case <-time.After(rateDelay - elapsed):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	c.lastCall = time.Now()
	return nil
}
