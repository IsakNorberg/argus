# Yahoo Finance

## API:er
- **v8:** `https://query2.finance.yahoo.com/v8/finance/chart/{symbol}`
- **V3:** `https://query2.finance.yahoo.com/v3/finance/`

## Data typer
| Typ | Data |
|---|---|
| Quote | Pris, volym, market cap |
| Profile | Sektor, industry, description |
| Income Statement | Revenue, cost, net income |
| Balance Sheet | Assets, liabilities, equity |
| Cash Flow | Operating, investing, financing |

## Auth
Ingen officiell API-nyckel. Används via unofficial libs.

## Nästa steg
- [ ] Välj Go lib (t.ex. `go-yahoo-finance` eller egen HTTP)
- [ ] Hämta per symbol
- [ ] Standardisera format
