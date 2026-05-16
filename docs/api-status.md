# API Status — Komplett Rapport (Deep Research)

> Uppdaterad: 2026-05-16
> Metod: 3 subagents + 50+ curl-tester + browser-research

---

## GRATIS och FUNGERAR (implementera dessa först)

| # | Register | Endpoint | Auth | Finansiell data |
|---|---|---|---|---|
| 1 | SEC EDGAR USA | https://data.sec.gov/api/xbrl/companyfacts/CIK{cik}.json | User-Agent | XBRL, komplett |
| 2 | PRH Finland | https://avoindata.prh.fi/opendata-xbrl-api/v3 | Ingen | XBRL |
| 3 | Brreg Norge | https://data.brreg.no/enhetsregisteret/api/enheter | Ingen | Begransad |
| 4 | ARES Tjeckien | POST https://ares.gov.cz/ekonomicke-subjekty-v-be/rest/ekonomicke-subjekty/vyhledat | Ingen | Bara bolagsinfo |
| 5 | Yahoo Finance | https://query1.finance.yahoo.com/v8/finance/chart/{TICKER} | User-Agent | Kurser |

---

## HITTAD - Fungerar med Setup

### KVK Nederlander na - FULLT FUNGERANDE TEST-API!

**Test-API (GRATIS, ingen registrering):**

- Search: `https://api.kvk.nl/test/api/v2/zoeken?naam=test`
- Profil: `https://api.kvk.nl/test/api/v1/basisprofielen/{kvkNummer}`
- Header: `apikey: l7xx1f2691f2520d487b902f4e0b57a0b197`

Verifierad curl:

```bash
curl "https://api.kvk.nl/test/api/v2/zoeken?naam=test" \
  -H "apikey: l7xx1f2691f2520d487b902f4e0b57a0b197" \
  -H "accept: application/json"
```

| Falt | Varde |
|---|---|
| Test-API key | `l7xx1f2691f2520d487b902f4e0b57a0b197` |
| Test-data KVK | 69599084 (test bolag) |
| Production | Betald (developers.kvk.nl/nl/pricing) |
| Developer portal | https://developers.kvk.nl/nl/ |

### Companies House UK - Kraver GRATIS API-nyckel

- Endpoint: `GET https://api.company-information.service.gov.uk/search/companies?q=test`
- Auth: Basic Auth (`curl -u "API_KEY: " ...`)
- Registrera: https://developer.company-information.service.gov.uk (gratis)
- Rate: 600 req/5min

### EDINET Japan - Kraver Registrering

- API finns men kraver Microsoft SSO-inloggning
- Registrering: https://disclosure2.edinet-fsa.go.jp/ -> "EDINET API"
- Production ar gratis efter registrering
- Ingen test-key tillganglig

---

## Alternativa Vagar (research needed)

### CVR Danmark - BLOCKAD av Cloudflare

- https://datafordeler.dk - Officiell bulk-distributor (GraphQL, kraver OAuth)
- https://virksomhedsguiden.dk - CVR data via annan portal
- https://cvrapi.dk - Third-party (ej officiell)

### INPI Frankrike - BLOCKAD av Cloudflare

- api.entreprise.api.gouv.fr - Sirene API (DNS fail fran detta host - fungerar troligen fran fransk/fr EU IP)
- api.insee.net - French national statistics

---

## EJ GENOMFORBAR (inga fungerande API:er)

| Register | Problem | Slutsats |
|---|---|---|
| Zefix Schweiz | Ingen API, bara SPA Angular | TA BORT |
| CRO Irland | Allt Cloudflare-blockerat | TA BORT |
| KRS Polen | Ingen public REST API | TA BORT |

---

## Implementationsprioritet

| # | Kalla | Status | Data |
|---|---|---|---|
| 1 | SEC EDGAR | Verifierad | XBRL (bast!) |
| 2 | PRH Finland | Verifierad | XBRL |
| 3 | KVK | Test-API | Bolagsinfo |
| 4 | Brreg Norge | Verifierad | Bolagsinfo |
| 5 | Companies House | Gratis key | iXBRL |
| 6 | ARES Tjeckien | Verifierad | Bara bolagsinfo |
| 7 | Yahoo | Verifierad | Kurser |
| -- | -- | -- | -- |
| 8 | Bolagsverket | API-nyckel | TBD |
| 9 | EDINET | Registrering | XBRL |
| 10 | CVR Danmark | datafordeler.dk | TBD |
| 11 | INPI Frankrike | Sirene API | TBD |
| -- | -- | -- | -- |
| | Zefix | Ingen API | TA BORT |
| | CRO | Cloudflare | TA BORT |
| | KRS | Ingen API | TA BORT |

---

## Detaljerade Endpoints

### SEC EDAGR - Company Facts

```json
{
  "cik": "320193",
  "entityName": "APPLE INC.",
  "facts": {
    "us-gaap": {
      "Assets": {
        "units": { "USD": [
          { "end": "2023-09-30", "val": 352583000000 }
        ]}
      }
    }
  }
}
```

### PRH - XBRL

```
GET https://avoindata.prh.fi/opendata-xbrl-api/v3/{y-tunnus}/xbrl
-> XBRL data i IFRS/ESEF-format
```

### KVK - Test Search

```
GET https://api.kvk.nl/test/api/v2/zoeken?naam=test&size=10
-> [{ "naam": "...", "kvkNummer": "...", "bezoekadres": {...} }]
```

### ARES - Sok bolag

```json
POST https://ares.gov.cz/ekonomicke-subjekty-v-be/rest/ekonomicke-subjekty/vyhledat
{"ico": ["47114983"]}
// eller
{"obchodniJmeno": "Ceska posta", "pocet": 5, "start": 0}
```

### Yahoo - Kurser

```
GET https://query1.finance.yahoo.com/v8/finance/chart/AAPL?interval=1d&range=5d
-> { chart: { result: [{ meta: { symbol: "AAPL", regularMarketPrice: 300.23 } }] }}
```

---

## Sammanfattning

- **11 register kvar** (8 fungerande + 3 med setup)
- **3 kallor kan tas bort** (Zefix, CRO, KRS)
- **SEC EDGAR och PRH** har bast XBRL-data och fungerar direkt
- **KVK har gratis test-API** som fungerar OMEDELBART
