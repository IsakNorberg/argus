package prh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (c *Client) Name() string { return "prh" }

// prhResponse är gemensam respons från PRH API.
type prhResponse struct {
	Version      string    `json:"version"`
	TotalResults int       `json:"totalResults"`
	ResultCount  int       `json:"resultCount"`
	Results      []prhItem `json:"results"`
}

type prhItem struct {
	Name           string `json:"name"`
	BusinessID     string `json:"businessId"`
	CompanyForm    string `json:"companyForm"`
	RegistrationDate string `json:"registrationDate"`
	DeletionDate   string `json:"deletionDate,omitempty"`
	Address        struct {
		Street   string `json:"street"`
		City     string `json:"city"`
		PostCode string `json:"postCode"`
	} `json:"address"`
	BusinessLines []struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"businessLines"`
}

type fd1Item struct {
	BusinessID   string `json:"businessId"`
	ClosingDate  string `json:"closingDate"`
	Currency     string `json:"currency"`
	Consolidated bool   `json:"consolidated"`
	FilingDate   string `json:"filingDate"`
	Period       int    `json:"period"`
	Status       string `json:"status"`
}

func (c *Client) FetchCompanies(ctx context.Context) ([]sources.Company, error) {
	// PRH har ingen "hämta alla"-endpoint. Vi söker med breda termer.
	// Använder pagination för att hämta så många som möjligt.
	var allCompanies []sources.Company

	// Sök med vanliga bolagsformer för att hitta flest möjliga
	searches := []string{"Oyj", "Oy", "Ab", "Ky"}
	for _, term := range searches {
		page := 0
		for {
			c.rateLimit()
			url := fmt.Sprintf("https://avoindata.prh.fi/avoindata/tr/pr?totalResults=true&name=%s&results=100&orderBy=name", term)
			if page > 0 {
				url += fmt.Sprintf("&results=100&offset=%d", page*100)
			}

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return allCompanies, fmt.Errorf("bygg request: %w", err)
			}
			req.Header.Set("User-Agent", "Argus (argus@isaknorberg.dev)")

			resp, err := c.client.Do(req)
			if err != nil {
				return allCompanies, fmt.Errorf("hämta PRH %s: %w", term, err)
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return allCompanies, fmt.Errorf("läs body: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				return allCompanies, fmt.Errorf("PRH HTTP %d: %s", resp.StatusCode, body)
			}

			var data prhResponse
			if err := json.Unmarshal(body, &data); err != nil {
				return allCompanies, fmt.Errorf("decode PRH: %w", err)
			}

			for _, item := range data.Results {
				company := sources.Company{
					Name:       item.Name,
					Ticker:     "", // PRH ger inte tickrar direkt
					Country:    "FI",
					ExternalID: item.BusinessID,
					Industry:   "",
					Exchange:   "HEX",
				}
				if len(item.BusinessLines) > 0 {
					company.Industry = item.BusinessLines[0].Name
				}
				allCompanies = append(allCompanies, company)
			}

			if len(data.Results) < 100 || page > 40 { // max ~4000 per search
				break
			}
			page++
		}
	}

	return allCompanies, nil
}

func (c *Client) FetchProfile(_ context.Context, companyID string) (*sources.Profile, error) {
	c.rateLimit()
	url := fmt.Sprintf("https://avoindata.prh.fi/avoindata/tr/pr?businessId=%s&totalResults=false&results=1", companyID)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("hämta profil: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PRH profil HTTP %d", resp.StatusCode)
	}

	var data prhResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode profil: %w", err)
	}

	if len(data.Results) == 0 {
		return nil, fmt.Errorf("ingen profil hittad")
	}

	item := data.Results[0]
	// Försök hitta company form från namnet (Oyj = Plc, Oy = Ltd)
	sector := "Ltd"
	if item.CompanyForm == "Plc" {
		sector = "Plc"
	}

	return &sources.Profile{
		CompanyID:   item.BusinessID,
		Description: fmt.Sprintf("%s is a Finnish %s", item.Name, item.CompanyForm),
		Sector:      sector,
		Industry:    "",
		Country:     "FI",
		Exchange:    "HEX",
	}, nil
}

func (c *Client) FetchFinancials(_ context.Context, companyID string) ([]sources.Financials, error) {
	c.rateLimit()
	var allFinancials []sources.Financials

	// Hämta FD1 (basic) och FD2 (detailed)
	for _, endpoint := range []string{"fd1", "fd2"} {
		url := fmt.Sprintf("https://avoindata.prh.fi/avoindata/fd/%s?businessId=%s&totalResults=false&results=100", endpoint, companyID)

		resp, err := c.client.Get(url)
		if err != nil {
			continue // Hoppa över om endpoint inte finns
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			continue
		}

		var data prhResponse
		if err := json.Unmarshal(body, &data); err != nil {
			continue
		}

		for _, item := range data.Results {
			var fd fd1Item
			jsonBytes, _ := json.Marshal(item)
			json.Unmarshal(jsonBytes, &fd)

			period := "FY"
			if fd.ClosingDate != "" {
				period = fd.ClosingDate[:4] + "FY"
			}

			fin := sources.Financials{
				CompanyID:  fd.BusinessID,
				Period:     period,
				Currency:   fd.Currency,
				ReportDate: fd.ClosingDate,
				FilingDate: fd.FilingDate,
				Source:     "prh",
			}
			allFinancials = append(allFinancials, fin)
		}
	}

	if len(allFinancials) == 0 {
		return nil, fmt.Errorf("ingen finansiell data hittad")
	}

	return allFinancials, nil
}

func (c *Client) rateLimit() {
	delay := time.Since(c.lastCall)
	if delay < 250*time.Millisecond {
		time.Sleep(250*time.Millisecond - delay)
	}
	c.lastCall = time.Now()
}
