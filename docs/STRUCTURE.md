# Argus — Projektstruktur

> Den med 100 ögon — samlar, standardiserar och lagrar bolagsdata från alla källor.

## Översikt

Argus hämtar bolagsdata från **14 officiella register**, standardiserar till ett gemensamt format och lagrar i SQLite (migrerbart till PostgreSQL).

## Källor (14 register)

### 🟢 Helt Gratis (11)
| Register | Land | Format | API-typ |
|---|---|---|---|
| SEC EDGAR | 🇺🇸 USA | XBRL/JSON | REST, helt öppet |
| CVR | 🇩🇰 Danmark | XBRL/JSON | REST, 250 req/min |
| Companies House | 🇬🇧 UK | JSON/iXBRL | Basic Auth (gratis) |
| PRH | 🇫🇮 Finland | JSON | REST, öppet |
| EDINET | 🇯🇵 Japan | XBRL | REST/SOAP, öppet |
| Zefix | 🇨🇭 Schweiz | JSON | REST, öppet |
| CRO | 🇮🇪 Irland | PDF/JSON | REST, öppet |
| SEDAR+ | 🇨🇦 Kanada | XBRL/PDF | REST, öppet |
| ARES | 🇨🇿 Tjeckien | JSON | REST, öppet |
| KRS | 🇵🇱 Polen | JSON/XML | REST, öppet |
| INPI | 🇫🇷 Frankrike | JSON | REST, öppet (data.inpi.fr) |

### 🟡 Kräver Registrering/Nyckel (3)
| Register | Land | Format | API-typ |
|---|---|---|---|
| Bolagsverket | 🇸🇪 Sverige | JSON | REST, API-nyckel |
| Brreg | 🇳🇴 Norge | JSON/XML | REST, grundläggande auth |
| KVK | 🇳🇱 Nederländerna | JSON | REST, API-nyckel (+ betald för bulk) |

### 📈 Kompletterande
| Källa | Typ |
|---|---|
| Yahoo Finance | Dagliga kurser, historisk prisdata (oofficiellt) |

## Arkitektur

```
┌─────────────────────────────────────────────────────────────┐
│                       Argus CLI                              │
│  cmd/argus/main.go                                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐    ┌──────────────┐    ┌────────────┐      │
│  │  Sources    │───▶│ Standardizer │───▶│  Database  │      │
│  │  (14 st)    │    │              │    │  (SQLite)  │      │
│  └─────────────┘    └──────────────┘    └────────────┘      │
│       │                    │                   │             │
│       │ FetchCompanies     │ ToCompany         │ Upsert      │
│       │ FetchFinancials    │ ToFinancials      │ Dedup       │
│       │ FetchProfile       │                   │ Query       │
│                                                              │
│  ┌──────────────┐                                           │
│  │ De-duplicator│  ← ISIN match, fallback: ticker+country+namn   │
│  └──────────────┘                                           │
│                                                              │
│  ┌──────────────┐                                           │
│  │ Price Service│  ← Yahoo Finance (dagliga kurser)          │
│  └──────────────┘                                           │
└─────────────────────────────────────────────────────────────┘
```

## Deduplicerings-strategi

Samma bolag kan finnas i flera register (ex: Apple i SEC + SEDAR+, Nokia i PRH + Companies House).

### Matchningsordning

| Steg | Metod | Confidence |
|---|---|---|
| 1 | **ISIN** match | 100% ✅ |
| 2 | **Ticker + Country + Namn** (fuzzy, Levenshtein) | >90% match → 80% ⚠️ |
| 3 | **Manuell mapping** (`source_mappings`-tabellen) | 100% ✅ |
| 4 | Lägre än 90% → flagga för granskning | 🔴 |

### Mapping-tabell
```sql
CREATE TABLE source_mappings (
    id INTEGER PRIMARY KEY,
    source TEXT NOT NULL,        -- "sec", "cvr", "companieshouse"...
    external_id TEXT NOT NULL,   -- "0000320193" (SEC CIK)
    argus_id INTEGER NOT NULL,   -- FK → companies.id
    confidence REAL DEFAULT 1.0, -- 1.0 (ISIN), 0.8 (fuzzy), 0.5 (manuell)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source, external_id)
);
```

## argus_id — ID-strategi

Unikt ID per bolag. 10 siffror:
- Första 2 = börs-prefix
- Nästa 8 = sekventiellt nummer

```
[EX] [XXXXXXXX]
 ↑          ↑
 │          └── Sekventiellt (00000001 - 99999999)
 └── Börs-prefix:
     10 = US (SEC/NASDAQ/NYSE)
     11 = DK (CPR/Nasdaq Copenhagen)
     12 = UK (Companies House/LSE)
     13 = FI (PRH/Nasdaq Helsinki)
     14 = NO (Brreg/Oslo Børs)
     15 = SE (Bolagsverket/Nasdaq Stockholm)
     16 = FR (INPI/Euronext Paris)
     17 = NL (KVK/Euronext Amsterdam)
     18 = CA (SEDAR+/TSX)
     19 = JP (EDINET/Tokyo)
     20 = CH (Zefix/SIX Swiss)
     21 = IE (CRO/Irish Stock Exchange)
     22 = CZ (ARES/Prague)
     23 = PL (KRS/Warsaw)
```

Exempel:
- Apple (US): `10 00000001` → `1000000001`
- Volvo (SE): `15 00000001` → `1500000001`
- Nokia (FI): `13 00000001` → `1300000001`

## Valutahantering

| Princip | Detalj |
|---|---|
| **Spara original-valuta** | Varje rad har sin valuta (SEK, USD, EUR, JPY, CAD, etc) |
| **Ingen automatisk konvertering** | Vi sparar rådata |
| **Konvertering sker i query** | `SELECT revenue * rate FROM financials JOIN exchange_rates ...` |
| **Växelkurs-tabell** | Dagliga kurser sparas separat för historisk korrekthet |

```sql
CREATE TABLE exchange_rates (
    from_currency TEXT,
    to_currency TEXT DEFAULT 'EUR',
    rate REAL,
    date DATE,
    source TEXT DEFAULT 'ECB',
    PRIMARY KEY (from_currency, date)
);
```

## Uppdateringsfrekvens

Varje källa har konfigurerbar frekvens:

| Källa-typ | Frekvens | Varför |
|---|---|---|
| **Fundamentals** (10-K, 10-Q, årsredovisningar) | Kvartalsvis/Årligen | Rapporterna ändras sällan |
| **Bolagsprofiler** (namn, sector, VD) | Månatligen | Ändras ibland |
| **Kurser** (Yahoo Finance) | Dagligen | Marknadsdata, ändras dagligen |

Konfigurerbart via `source_schedule`-tabell:
```sql
CREATE TABLE source_schedule (
    source TEXT PRIMARY KEY,
    fundamentals_interval TEXT DEFAULT 'quarterly',
    profile_interval TEXT DEFAULT 'monthly',
    prices_interval TEXT DEFAULT 'daily',
    enabled BOOLEAN DEFAULT true
);
```

## Pagination-strategi

Alla API:er returnerar paginerade resultat. Standardiserad hantering:

```go
type PaginatedResponse struct {
    Items      []Company
    TotalCount int
    NextCursor string   // eller NextPage
    HasMore    bool
}

// Fetch med retry + pagination
func (c *SECClient) FetchAll(ctx context.Context) ([]Company, error) {
    var all []Company
    cursor := ""
    for {
        resp, err := c.fetchPage(ctx, cursor)
        if err != nil { return nil, err }
        all = append(all, resp.Items...)
        if !resp.HasMore { break }
        cursor = resp.NextCursor
    }
    return all, nil
}
```

## Rate Limit & Retry

| Källa | Limit | Retry-strategi |
|---|---|---|
| SEC | 10 req/s | Exponential backoff, max 3 retries |
| CVR | 250 req/min | Sleep mellan requests |
| Companies House | 600 req/5min | Bucket-baserad |
| Alla övriga | Okänt | Conservative: 2 req/s, backoff |

```go
type RateLimiter struct {
    mu       sync.Mutex
    requests map[string]time.Time
    limit    time.Duration
}

func (l *RateLimiter) Wait(key string) {
    // Sleep until rate limit allows
}
```

## Mappstruktur

```
argus/
├── cmd/argus/
│   └── main.go                 ← CLI entry point
│
├── internal/
│   ├── sources/                ← 14 källor + Yahoo
│   │   ├── source.go           ← Interface
│   │   ├── sec/                ← 🇺🇸 SEC EDGAR
│   │   ├── bolagsverket/       ← 🇸🇪 Bolagsverket
│   │   ├── brreg/              ← 🇳🇴 Brønnøysundregistrene
│   │   ├── cvr/                ← 🇩🇰 CVR
│   │   ├── prh/                ← 🇫🇮 PRH
│   │   ├── companieshouse/     ← 🇬🇧 Companies House
│   │   ├── inpi/               ← 🇫🇷 INPI
│   │   ├── kvk/                ← 🇳🇱 KVK
│   │   ├── edinet/             ← 🇯🇵 EDINET (Japan)
│   │   ├── zefix/              ← 🇨🇭 Schweiz
│   │   ├── cro/                ← 🇮🇪 Irland
│   │   ├── sedar/              ← 🇨🇦 SEDAR+ (Kanada)
│   │   ├── ares/               ← 🇨🇿 Tjeckien
│   │   ├── krs/                ← 🇵🇱 Polen
│   │   └── yahoo/              ← 📈 Yahoo Finance (kurser)
│   │
│   ├── xbrl/
│   │   └── parser.go           ← Gemensam XBRL-parser
│   │   └── mappings.go         ← Fält-mappning per taxonomy
│   │
│   ├── standardizer/
│   │   └── standardizer.go     ← Källdata → gemensamt format
│   │
│   ├── dedup/
│   │   └── dedup.go            ← ISIN match + fuzzy + mapping
│   │
│   ├── prices/
│   │   └── prices.go           ← Yahoo fetcher + exchange rates
│   │
│   └── database/
│       └── database.go         ← SQLite/PostgreSQL
│
├── pkg/models/
│   └── models.go               ← Company, Financials, PriceQuote
│
├── testdata/                   ← Fixtures för unit tests
│   ├── sec/
│   │   └── sample_cik_0000320193.json    ← Apple SEC facts
│   ├── cvr/
│   │   └── sample_company.json
│   └── ...
│
├── docs/
│   ├── STRUCTURE.md            ← Detta dokument
│   ├── IMPLEMENTATION.md       ← Implementeringsplan per källa
│   ├── migrate.md              ← SQLite → PostgreSQL
│   └── sources/                ← MD per källa (API-info)
│
├── data/                       ← argus.db (gitignored)
├── go.mod
├── go.sum
└── README.md
```

## Datamodeller

### Company (bolag)
| Fält | Typ | Beskrivning |
|---|---|---|
| `argus_id` | TEXT (10 siffror) | Vårt unika ID: `[EX][seq]` |
| `name` | TEXT | Bolagsnamn |
| `ticker` | TEXT | Börsförkortning |
| `cik` | TEXT | SEC CIK |
| `org_number` | TEXT | Organisationsnummer |
| `isin` | TEXT | Internationellt ID |
| `country` | TEXT | Land |
| `industry` | TEXT | Bransch |
| `sector` | TEXT | Sektor |
| `exchange` | TEXT | Börs |
| `description` | TEXT | Beskrivning |
| `website` | TEXT | Webbplats |
| `employees` | INT | Antal anställda |
| `ceo` | TEXT | VD/CEO |

### Financials (balansräkning, resultat)
| Fält | Typ | Beskrivning |
|---|---|---|
| `argus_id` | TEXT | Unikt ID per row |
| `company_id` | INT | FK → companies.id |
| `period` | TEXT | "2024Q4", "2024FY" |
| `currency` | TEXT | SEK, USD, EUR, JPY... (original) |
| `revenue` | REAL | Intäkter |
| `net_income` | REAL | Nettoresultat |
| `total_assets` | REAL | Tillgångar |
| `total_debt` | REAL | Skulder |
| `equity` | REAL | Eget kapital |
| `eps` | REAL | Resultat per aktie |
| `ebitda` | REAL | EBITDA |
| `free_cash_flow` | REAL | Fritt kassaflöde |
| `report_date` | DATETIME | Rapportdatum |
| `filing_date` | DATETIME | Inlämningsdatum |
| `source` | TEXT | Källa ("sec", "cvr", ...) |

### Price Quotes (dagliga kurser)
| Fält | Typ | Beskrivning |
|---|---|---|
| `argus_id` | TEXT | Unikt ID per row |
| `company_id` | INT | FK → companies.id |
| `ticker` | TEXT | Börsförkortning |
| `price` | REAL | Kurs |
| `currency` | TEXT | Valuta |
| `market_cap` | INT | Börsvärde |
| `volume` | INT | Dagsvolym |
| `change_percent` | REAL | Förändring % |
| `source` | TEXT | Alltid "yahoo" |
| `quote_date` | DATETIME | Datum |

## Teknisk Stack

| Komponent | Val | Varför |
|---|---|---|
| Språk | Go 1.23 | Concurrency, prestanda, binärer |
| Databas (nu) | SQLite (`modernc.org/sqlite`) | Pure Go, ingen CGO |
| Databas (sen) | PostgreSQL | Migration via `database/sql` |
| HTTP | Standard `net/http` | Inget external dependency |
| JSON | Standard `encoding/json` | Inbyggd |
| XML/XBRL | `encoding/xml` | Inbyggd |
| Fuzzy match | `github.com/hbollon/go-edlib` | Levenshtein distance |
| Rate limiting | `golang.org/x/time/rate` | Token bucket |
| CLI | `github.com/spf13/cobra` | Robust CLI framework |
| Test | Go `testing`, `httptest` | Standard + mocks |
