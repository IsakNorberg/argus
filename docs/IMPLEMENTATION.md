# IMPLEMENTERINGPLAN — Per Källa

> Detaljerad guide för att implementera varje källa.
> Ordning: Börja med de som har bäst API → lägre prioritet.

## Viktigt innan vi börjar

### API-verifiering
**ALLA endpoints måste verifieras med curl innan implementation.**
Aldrig implementera mot gissad endpoint.

#### Verifierings-process
```bash
# Steg 1: Kolla om API:n är uppe
curl -s -o /dev/null -w "%{http_code}" https://data.sec.gov/submissions/CIK0000320193.json

# Steg 2: Kolla response
curl -s "https://data.sec.gov/submissions/CIK0000320193.json" | jq '.name, .facts'

# Steg 3: Kolla pagination
curl -s "https://datacvr.virk.dk/api/..." | jq '.pagination'

# Steg 4: Kolla rate limit (20 snabba requests)
for i in {1..20}; do curl -s -o /dev/null -w "%{http_code}\n" URL; done
```

### Färgad status-uppdatering
| Register | Status | Verifierad? |
|---|---|---|
| SEC | 🟢 Gratis | ✅ Endpoint verifierad |
| Companies House | 🟡 Basic Auth (gratis) | ❌ Ej verifierad |
| INPI | 🟢 Gratis | ✅ Öppet API |
| SEDAR+ | 🟢? Gratis | ❌ **KRÄVER VERIFIERING** |
| ARES | 🟢? Gratis | ❌ **KRÄVER VERIFIERING** |
| Zefix | 🟢? Gratis | ❌ **KRÄVER VERIFIERING** |
| CRO | 🟢? Gratis | ❌ **KRÄVER VERIFIERING** |
| KRS | 🟢? Gratis | ❌ **KRÄVER VERIFIERING** |
| CVR | 🟢 Gratis | ❌ Ej verifierad |
| PRH | 🟢 Gratis | ❌ Ej verifierad |
| EDINET | 🟢? Gratis | ❌ **KRÄVER VERIFIERING** |
| Brreg | 🟡 Basic Auth | ❌ Ej verifierad |
| Bolagsverket | 🟡 Nyckel | ❌ Ej verifierad |
| KVK | 🔴 Betald | ❌ Ej verifierad |

---

## Prioritering & Implementationsordning

| # | Källa | Svårighet | Nytta | API-status |
|---|---|---|---|---|
| 1 | SEC EDGAR | Låg | Mycket hög | ✅ Verifierad |
| 2 | Companies House | Låg | Hög | ❌ Verifiera |
| 3 | CVR | Låg | Hög | ❌ Verifiera |
| 4 | PRH | Låg | Medium | ❌ Verifiera |
| 5 | INPI | Låg | Medium | ❌ Verifiera |
| 6 | EDINET | Medium | Mycket hög | ❌ **KRÄVER VERIFIERING** |
| 7 | SEDAR+ | Okänt | Hög | ❌ **KRÄVER VERIFIERING** |
| 8 | Brreg | Låg | Medium | ❌ Verifiera |
| 9 | ARES | Låg | Medium | ❌ **KRÄVER VERIFIERING** |
| 10 | Zefix | Låg | Låg | ❌ **KRÄVER VERIFIERING** |
| 11 | KRS | Låg | Medium | ❌ **KRÄVER VERIFIERING** |
| 12 | CRO | Medium | Låg | ❌ **KRÄVER VERIFIERING** |
| 13 | Bolagsverket | Medium | Hög | ❌ Nyckel behövs |
| 14 | KVK | Hög | Medium | 🔴 Betald |

---

## 1. SEC EDGAR 🇺🇸

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://data.sec.gov` |
| Auth | Ingen. User-Agent med email: `User-Agent: Argus-Collector isak@example.com` |
| Rate Limit | 10 req/s rekommenderat |
| Format | JSON + XBRL |
| API-verifierad | ✅ |

### Plan

#### Steg 1: Hämta CIK-lista
```
GET https://www.sec.gov/Archives/edgar/cik-lookup-data.txt
```
- Format: `<company name>|<CIK>`
- Storlek: ~85MB, ~500k entries
- Parsa till `map[string]string`
- Lagra lokalt i `data/cik_list.txt`

#### Steg 2: Fetch Companies
```
GET https://data.sec.gov/submissions/CIK{cik}.json
# Ex: https://data.sec.gov/submissions/CIK0000320193.json (Apple)
```

**Response structure:**
```json
{
  "cik": "0000320193",
  "name": "Apple Inc.",
  "ticker": "AAPL",
  "sic": "3571",
  "sicDescription": "Electronic Computers",
  "ein": "942404110",
  "fiscalYearEnd": "0930",
  "filings": {
    "recent": {
      "accessionNumber": [...],
      "filingDate": [...],
      "reportDate": [...],
      "form": ["10-K", "10-Q", ...],
      "fileNumber": [...],
      "primaryDocument": [...]
    }
  }
}
```

**Fältsmappning → Company:**
| SEC Fält | Vårt Fält |
|---|---|
| `name` | `name` |
| `ticker` | `ticker` |
| `cik` | `cik` |
| `sic` | används inte direkt (behövs SIC→Industry mapping) |
| `ein` | används inte (US tax ID) |

#### Steg 3: Fetch Financials (Company Facts)
```
Samma endpoint! data finns i .facts:
{
  "facts": {
    "us-gaap": {
      "Assets": {
        "label": "Assets",
        "description": "Sum of carrying amounts...",
        "units": {
          "USD": [
            {
              "start": "2023-09-30",
              "end": "2023-09-30",
              "val": 352583000000,
              "accn": "0000320193-23-000077",
              "frame": "CY2023Q3I"
            }
          ]
        }
      }
    }
  }
}
```

**us-gaap tagg-mappning → Financials:**
| us-gaap Tag | Vårt Fält |
|---|---|
| `us-gaap/Assets` | `total_assets` |
| `us-gaap/Liabilities` | `total_debt` |
| `us-gaap/StockholdersEquity` | `equity` |
| `us-gaap/RevenueFromContractWithCustomerExcludingAssessedTax` | `revenue` |
| `us-gaap/RevenueFromContractWithCustomer` | `revenue` (fallback) |
| `us-gaap/NetIncomeLoss` | `net_income` |
| `us-gaap/EarningsPerShareBasic` | `eps` |
| `us-gaap/OperatingIncomeLoss` | `ebitda` (proxy) |
| `us-gaap/NetCashProvidedByUsedInOperatingActivities` | `free_cash_flow` (proxy) |

#### Steg 4: Pagination
Endpoint returnerar alla filings i `recent`-objekt.
För äldre filings → använd pagination med `index.json`.

#### Kodstruktur

```go
// internal/sources/sec/client.go
package sec

type SECClient struct {
    baseURL   string
    userEmail string
    client    *http.Client
    ciKMap    map[string]string  // name → CIK
}

func New() *SECClient {
    return &SECClient{
        baseURL:   "https://data.sec.gov",
        userEmail: os.Getenv("ARGUS_EMAIL"), // default: "argus@example.com"
        client: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        50,
                MaxIdleConnsPerHost: 20,
            },
        },
    }
}

func (c *SECClient) doRequest(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("User-Agent", "Argus Collector "+c.userEmail)
    // Exec med retry (3 ggr, exponential backoff)
    // Exponera på 429, 503
}

// internal/sources/sec/cik.go
func (c *SECClient) LoadCIKList(ctx context.Context) error {
    // Hämta cik-lookup-data.txt
    // Parsa → map[string]string
}

// internal/sources/sec/parse.go
type CompanyFacts struct {
    CIK   string `json:"cik"`
    Name  string `json:"name"`
    Ticker string `json:"ticker"`
    SIC    string `json:"sic"`
    Facts map[string]map[string]FactData `json:"facts"`
}

type FactData struct {
    Label       string     `json:"label"`
    Description string     `json:"description"`
    Units       map[string][]FactValue `json:"units"`
}

type FactValue struct {
    Start string  `json:"start"`
    End   string  `json:"end"`
    Val   float64 `json:"val"`
    Accn  string  `json:"accn"`
}
```

### Testning

**Mock data:** Kopiera från verklig SEC API
```bash
curl -s -H "User-Agent: Argus isak@example.com" \
  "https://data.sec.gov/submissions/CIK0000320193.json" > \
  testdata/sec/apple_facts.json
```

---

## 2. Companies House 🇬🇧

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://api.company-information.service.gov.uk` |
| Auth | Basic Auth (API-nyckel — **GRATIS**) |
| Rate Limit | 600 req/5min |
| Format | JSON + iXBRL |
| API-verifierad | ❌ Behöver verifieras |

### Plan

#### Steg 1: Skaffa API-nyckel
1. Gå till `https://developer.company-information.service.gov.uk`
2. Skapa konto (gratis)
3. Generera API key
4. Spara i env: `CH_API_KEY=xxx`

#### Steg 2: Fetch Companies
```
GET /search/companies?q={query}&items_per_page=20&start_index=0
GET /company/{company_number}
```

**Company-fält:**
| CH Fält | Vårt Fält |
|---|---|
| `company_name` | `name` |
| `company_number` | används som external_id |
| `company_status` | används för att filtrera (active/dissolved) |
| `sic_codes` | används för industry/sector |
| `date_of_creation` | används |
| `jurisdiction` | `country` |
| `accounts.next_due` | används för nästa filing |

#### Steg 3: Fetch Financials
```
GET /company/{number}/filing-history
# Ger lista med filings. Varje filing har document_url:
GET /company/{number}/filing-history/{transaction_id}/document/{hash}/content
# Returnerar PDF eller iXBRL
```

**Notera:** Många accounts är PDF. iXBRL för större bolag.

---

## 3. CVR 🇩🇰

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://datacvr.virk.dk/api` |
| Auth | Ingen |
| Rate Limit | 250 req/min |
| Format | JSON + XBRL |
| API-verifierad | ❌ Behöver verifieras |

### Plan

#### Steg 1: Fetch Companies
```
GET /api/virksomhed?virksomhedStatus=normal
# Ger alla aktiva bolag
```

**Fält:**
| CVR Fält | Vårt Fält |
|---|---|
| `cvrNummer` | external_id |
| `navne` | `name` |
| `adresse` | address |
| `branchekode` | industry/sector (NACE-kod) |
| `regionskode` | country |

#### Steg 2: Fetch Financials
```
GET /api/regnskab/{cvr_nummer}
# XBRL-format (danskt taxonomy)
```

---

## 4. PRH 🇫🇮

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://avoindata.prh.fi` |
| Auth | Ingen |
| Rate Limit | Okänd |
| Format | JSON |
| API-verifierad | ❌ Behöver verifieras |

### Plan

#### Steg 1: Fetch Companies
```
GET /ytj/yritykset
# Y-tunnus-baserat register
```

---

## 5. INPI 🇫🇷

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://data.inpi.fr` |
| Auth | Ingen |
| Rate Limit | Okänd |
| Format | JSON |
| API-verifierad | ❌ Behöver verifieras |

### Plan

#### Steg 1: Fetch Companies
```
GET /api/sirene/v3/etablissements?q={query}
```

---

## 6. EDINET 🇯🇵

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://disclosure2.edinet-fsa.go.jp` |
| Auth | Ingen |
| Type | REST/SOAP |
| Format | XBRL |
| API-verifierad | ❌ **BEHÖVER VERIFIERING** |

### Viktigt

EDINET API är inte så lättillgängligt som SEC. Dokumentationen är på japanska.
Många endpoints kräver SOAP.

#### Plan
1. Kolla om det finns ett REST API: `https://disclosure2.edinet-fsa.go.jp/webe0/`
2. Om bara SOAP: implementera SOAP client
3. XBRL taxonomy är japansk (jp-gaap)

---

## 7. SEDAR+ 🇨🇦

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://sedarplus.ca` |
| API URL | Okänd — behöver hittas |
| Auth | Okänd |
| Format | XBRL/PDF |
| API-verifierad | ❌ **KAN VARA FEL — KANSKE INGET API** |

### ⚠️ Varning

Sedarplus.ca är **huvudsakligen en websida för sökningar**. Det kan vara att det inte finns något publikt API.

**Plan:**
1. Testa med curl om API finns
2. Om inget API: scrapa HTML (sämre lösning)
3. Alternativt: använd `https://www.sedar.com` (gamla versionen)
4. Om inget fungerar: **ta bort som källa**

---

## 8. Brreg 🇳🇴

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://data.brreg.no` |
| Auth | Basic Auth (gratis) |
| Format | JSON/XML |
| API-verifierad | ❌ Behöver verifieras |

### Plan

#### Steg 1: Fetch Companies
```
GET /enhetsregisteret/api/enheter
# Ger alla registrerade enheter
```

---

## 9. ARES 🇨🇿

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://ares.gov.cz` |
| Auth | Okänd |
| Format | JSON |
| API-verifierad | ❌ **BEHÖVER VERIFIERING** |

### ⚠️ Varning

ARES-API:et kan vara annorlunda än förväntat.
**Plan:**
1. Verifiera med curl först
2. Om API är begränsat: överväg att ta bort

---

## 10. Zefix 🇨🇭

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://www.zefix.ch` |
| API URL | `https://www.egiz.admin.ch/rest/zefixapp/api/search/` |
| Auth | Okänd |
| Format | JSON |
| API-verifierad | ❌ **BEHÖVER VERIFIERING** |

### ⚠️ Varning

Zefix har ett REST API via `egiz.admin.ch`. Detta kan vara begränsat.

---

## 11. KRS 🇵🇱

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://ems.ms.gov.pl` |
| Auth | Okänd |
| Format | JSON/XML |
| API-verifierad | ❌ **BEHÖVER VERIFIERING** |

---

## 12. CRO 🇮🇪

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://www.cro.ie` |
| Auth | Okänd |
| Format | PDF/JSON |
| API-verifierad | ❌ **BEHÖVER VERIFIERING** |

---

## 13. Bolagsverket 🇸🇪

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://bolagsverket.se/datamangder/oppnadata` |
| Auth | API-nyckel (ansökas) |
| Format | JSON |
| API-verifierad | ❌ Ej verifierad |

---

## 14. KVK 🇳🇱

### API-info
| Parameter | Värde |
|---|---|
| Base URL | `https://api.kvk.nl` |
| Auth | API-nyckel (+ betald för bulk) |
| Format | JSON |
| API-verifierad | ❌ Ej verifierad |
| Status | 🔴 Troligen betald |

---

## Implementation Workflow

För varje källa (i prioriterad ordning):

1. **Verifiering** (30 min)
   - Testa base URL med curl
   - Kolla om API:n svarar (200/401/403)
   - Identifiera endpoints

2. **Implementering** (2-4h)
   - Skapa klient med rätt auth
   - Implementera FetchCompanies
   - Implementera FetchFinancials
   - Implementera FetchProfile

3. **Testning** (1h)
   - Fixturer från mock data
   - Integrationstest mot live API
   - Validera fältmappning

4. **Deduplicering** (1h)
   - Kolla ISIN-matchning
   - Kör fuzzy match för bolag som finns i tidigare källor
   - Rapportera conflicts

---

## Nästa steg

**Börja med SEC EDGAR:**
1. ✅ Endpoint redan verifierad
2. Nästa: Hämta Apple (CIK0000320193) som test case
3. Implementera FetchCompanies → FetchFinancials → dedup
4. Skapa fixtures

**Efter SEC:** Verifiera Companies House API (grundläggande, gratis).

---

## Okända som behöver svar

| Fråga | Förslag |
|---|---|
| **SEDAR+ har API?** | Testa — om inte, ta bort |
| **EDINET har REST?** | Testa — SOAP kräver extra dependencies |
| **KVK gratis?** | Troligen inte — kanske ta bort tidigt |
| **CRO API existerar?** | Testa — CRO är främst web-baserad |
| **Zefix API är öppet?** | Testa — kan vara begränsat |
| **Vilken User-Agent/email för SEC?** | Sätt default argus email |
| **Vilken filtyp för fixtures?** | JSON (raw API response) |