# EODHD — Historical & Fundamental Data

## API:er
- **Base URL:** `https://eodhd.com/api/`
- **Fundamentals:** `/fundamentals/{symbol}?api_token=TOKEN`
- **End-of-Day:** `/eod/{symbol}?api_token=TOKEN`

## Data typer
| Endpoint | Data |
|---|---|
| `/fundamentals/` | P/E, EPS, revenue, margins, balance sheet |
| `/eod/` | Dagliga kurser |
| `/screener/` | filtrera marknad, sektor, market cap |

## Rate limit
Beroende på plan. Standard: ~300 req/min.

## Auth
API-nyckel i query param `?api_token=TOKEN`

## Nästa steg
- [ ] Använd samma nyckel som Pythia
- [ ] Hämta fundamentals för alla tickers
- [ ] Mappa mot Argus schema
