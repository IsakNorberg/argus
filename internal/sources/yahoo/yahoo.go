package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"time"
)

// PriceQuote är en prisnotering.
type PriceQuote struct {
	Ticker        string
	Price         float64
	Currency      string
	MarketCap     int64
	Volume        int64
	ChangePercent float64
	Source        string
}

// Client hämtar kursdata från Yahoo Finance.
type Client struct {
	client   *http.Client
	crumb    string
	lastCall time.Time
}

func New() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

type quoteResponse struct {
	QuoteResponse struct {
		Result []struct {
			Symbol                 string  `json:"symbol"`
			RegularMarketPrice     float64 `json:"regularMarketPrice"`
			Currency               string  `json:"currency"`
			MarketCap              float64 `json:"marketCap"`
			RegularMarketVolume    float64 `json:"regularMarketVolume"`
			RegularMarketChangePct float64 `json:"regularMarketChangePercent"`
		} `json:"result"`
		Error any `json:"error"`
	} `json:"quoteResponse"`
}

// getCrumb hämtar en Yahoo crumb för API-autentisering.
func (c *Client) getCrumb(ctx context.Context) error {
	if c.crumb != "" {
		return nil
	}

	// Steg 1: Besök Yahoo Finance för att få cookies
	req, err := http.NewRequestWithContext(ctx, "GET", "https://finance.yahoo.com/quote/AAPL/", nil)
	if err != nil {
		return fmt.Errorf("bygg request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("hämta cookies: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("yahoo cookies: http %d", resp.StatusCode)
	}

	// Steg 2: Hämta crumb
	req2, err := http.NewRequestWithContext(ctx, "GET", "https://query2.finance.yahoo.com/v1/test/getcrumb", nil)
	if err != nil {
		return fmt.Errorf("bygg crumb request: %w", err)
	}
	req2.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp2, err := c.client.Do(req2)
	if err != nil {
		return fmt.Errorf("hämta crumb: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return fmt.Errorf("crumb: http %d", resp2.StatusCode)
	}

	crumb, err := io.ReadAll(resp2.Body)
	if err != nil {
		return fmt.Errorf("läs crumb: %w", err)
	}

	c.crumb = strings.TrimSpace(string(crumb))
	if c.crumb == "" {
		return fmt.Errorf("tom crumb")
	}

	return nil
}

// FetchQuote hämtar realtidsdata för en eller flera tickers.
func (c *Client) FetchQuote(ctx context.Context, tickers []string) ([]*PriceQuote, error) {
	if len(tickers) == 0 {
		return nil, fmt.Errorf("inga tickers angivna")
	}

	c.rateLimit()

	if err := c.getCrumb(ctx); err != nil {
		// Fallback: försök utan crumb
		return c.fetchWithoutCrumb(ctx, tickers)
	}

	symbols := strings.Join(tickers, ",")
	url := fmt.Sprintf("https://query2.finance.yahoo.com/v6/finance/quote?symbols=%s&crumb=%s", symbols, c.crumb)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("bygg request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hämta quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Om crumb är ogiltig, försök förnya
		if resp.StatusCode == 401 {
			c.crumb = ""
			if err2 := c.getCrumb(ctx); err2 == nil {
				return c.FetchQuote(ctx, tickers) // retry med ny crumb
			}
		}
		return nil, fmt.Errorf("yahoo http %d: %s", resp.StatusCode, body)
	}

	var data quoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	if len(data.QuoteResponse.Result) == 0 {
		if data.QuoteResponse.Error != nil {
			return nil, fmt.Errorf("yahoo fel: %v", data.QuoteResponse.Error)
		}
		return nil, fmt.Errorf("inga resultat från Yahoo")
	}

	var quotes []*PriceQuote
	for _, r := range data.QuoteResponse.Result {
		q := &PriceQuote{
			Ticker:        r.Symbol,
			Price:         r.RegularMarketPrice,
			Currency:      r.Currency,
			MarketCap:     int64(r.MarketCap),
			Volume:        int64(r.RegularMarketVolume),
			ChangePercent: r.RegularMarketChangePct,
			Source:        "yahoo",
		}
		quotes = append(quotes, q)
	}

	return quotes, nil
}

// fetchWithoutCrumb - fallback utan crumb-autentisering.
func (c *Client) fetchWithoutCrumb(ctx context.Context, tickers []string) ([]*PriceQuote, error) {
	symbols := strings.Join(tickers, ",")
	url := fmt.Sprintf("https://query2.finance.yahoo.com/v7/finance/quote?symbols=%s", symbols)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("bygg request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hämta quote: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo http %d: %s", resp.StatusCode, bodyStr)
	}

	// Försök parse JSON
	var data quoteResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		// Försök hitta embedded JSON i HTML
		return c.parseFromHTML(ctx, bodyStr, tickers)
	}

	if len(data.QuoteResponse.Result) == 0 {
		return nil, fmt.Errorf("inga resultat")
	}

	var quotes []*PriceQuote
	for _, r := range data.QuoteResponse.Result {
		q := &PriceQuote{
			Ticker:        r.Symbol,
			Price:         r.RegularMarketPrice,
			Currency:      r.Currency,
			MarketCap:     int64(r.MarketCap),
			Volume:        int64(r.RegularMarketVolume),
			ChangePercent: r.RegularMarketChangePct,
			Source:        "yahoo",
		}
		quotes = append(quotes, q)
	}

	return quotes, nil
}

// parseFromHTML extraherar prisdata från Yahoo Finance HTML-sidor.
func (c *Client) parseFromHTML(ctx context.Context, html string, tickers []string) ([]*PriceQuote, error) {
	// Försök hitta "regularMarketPrice":{"raw":X.XX
	re := regexp.MustCompile(`"regularMarketPrice":\{"raw":([0-9.]+)`)
	
	var quotes []*PriceQuote
	for _, ticker := range tickers {
		match := re.FindStringSubmatch(html)
		if len(match) >= 2 {
			var price float64
			fmt.Sscanf(match[1], "%f", &price)
			quotes = append(quotes, &PriceQuote{
				Ticker: ticker,
				Price:  price,
				Source: "yahoo",
			})
		}
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("kunde inte extrahera pris från HTML")
	}

	return quotes, nil
}

func (c *Client) rateLimit() {
	delay := time.Since(c.lastCall)
	if delay < 1*time.Second {
		time.Sleep(1*time.Second - delay)
	}
	c.lastCall = time.Now()
}
