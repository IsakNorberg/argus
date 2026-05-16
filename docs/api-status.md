# API Status — Verifierade Endpoints

> Uppdaterad: 2026-05-16
> Metod: Subagents + curl-testning

## ✅ FUNGERAR (implementera dessa först)

| Register | Status | Endpoint | Auth | Finansiell data |
|---|---|---|---|---|
| 🇺🇸 **SEC EDGAR** | ✅ Full | `https://data.sec.gov/submissions/CIK{cik}.json` | Ingen (User-Agent) | ✅ XBRL, 3.7MB/bolag |
| 🇳🇴 **Brreg** | ✅ Full | `https://data.brreg.no/enhetsregisteret/api/enheter` | Ingen | ⚠️ Begränsad |
| 🇫🇮 **PRH** | ✅ Full | `https://avoindata.prh.fi/opendata-ytj-api/v3` | **Ingen** | ✅ XBRL separat: `/opendata-xbrl-api/v3` |
| 🇨🇿 **ARES** | ✅ Bolag | `POST https://ares.gov.cz/ekonomicke-subjekty-v-be/rest/ekonomicke-subjekty/vyhledat` | Ingen | ❌ Bara bolagsinfo |
| 🇺🇸 **Yahoo** | ✅ Kurser | `https://query1.finance.yahoo.com/v8/finance/chart/{TICKER}` | Ingen (User-Agent) | ✅ Dagliga kurser |

### 🇫🇮 PRH Detail (Ny Upptäckt!)
```
Bolag:    GET https://avoindata.prh.fi/opendata-ytj-api/v3/companies?name=test
XBRL:     GET https://avoindata.prh.fi/opendata-xbrl-api/v3/{y-tunnus}/xbrl
Notiser:  GET https://avoindata.prh.fi/opendata-registerednotices-api/v3
Pagination: &page=1&results=85
```

### 🇺🇸 SEC EDGAR Detail (Verifierad!)
```
Submissions: GET https://data.sec.gov/submissions/CIK0000320193.json
Company Facts: GET https://data.sec.gov/api/xbrl/companyfacts/CIK0000320193.json
CIK List: GET https://www.sec.gov/Archives/edgar/cik-lookup-data.txt
Namespaces: us-gaap, dei
Size: ~3.7MB per bolag (Apple)
```

## ⚠️ KRÄVER SETUP (men fungerar)

| Register | Status | Vad krävs |
|---|---|---|
| 🇬🇧 **Companies House** | 401 | Gratis API-nyckel (registrera: developer.company-information.service.gov.uk) |
| 🇸🇪 **Bolagsverket** | Okänd | API-nyckel (ansökan krävs) |
| 🇯🇵 **EDINET** | 403 | Kan kräva specifik User-Agent eller headers |
| 🇳🇱 **KVK** | 403 | API-nyckel (kostar för bulk) |

## ❌ EJ FUNGERANDE

| Register | Problem | Lösning |
|---|---|---|
| 🇩🇰 **CVR** | Cloudflare CAPTCHA | Bulk-dump eller partner-åtkomst |
| 🇫🇷 **INPI** | Cloudflare CAPTCHA | INPI-konto krävs |
| 🇨🇭 **Zefix** | 000 (nås inte) | Okänd |
| 🇮🇪 **CRO** | 403 | Okänd |
| 🇨🇦 **SEDAR+** | Bara hemsida (301) | Inget API hittat |
| 🇵🇱 **KRS** | Omdirigerar | Okänd |

## Implementations-prioritet

1. **SEC EDGAR** — Bäst API, mest data ✅
2. **PRH 🇫🇮** — Nytt! Fungerar, 3 API:er ✅
3. **Brreg 🇳🇴** — Fungerar ✅
4. **ARES 🇨🇿** — Bolagsinfo (ej finansiell) ✅
5. **Yahoo** — Kurser ✅
6. **Companies House 🇬🇧** — Skaffa API-nyckel
7. Övriga ❌ — Lägg på is

## Nyckel-fält per API

### SEC us-gaap → Vårt schema
| SEC us-gaap | Vårt Fält |
|---|---|
| Assets | total_assets |
| Liabilities | total_debt |
| StockholdersEquity | equity |
| RevenueFromContractWithCustomer | revenue |
| NetIncomeLoss | net_income |
| EarningsPerShareBasic | eps |

### PRH XBRL → Vårt schema
PRH använder IFRS/EU-taxonomi. Samma taggar som ESEF:
- `Assets` → total_assets
- `Liabilities` → total_debt
- `Equity` → equity

### Brreg → Vårt schema
- `navn` → name
- `organisasjonsnummer` → org_number
- `naeringskode1.kode` → industry
