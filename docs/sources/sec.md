# SEC — U.S. Securities and Exchange Commission

## API:er
- **Company Facts:** `https://data.sec.gov/submissions/CIK{cik}.json`
- **Full-Text Search:** `https://efts.sec.gov/LATEST/search-index`
- **Browser EDGAR:** `https://www.sec.gov/cgi-bin/browse-edgar`

## Data typer
| Typ | Beskrivning | Frekvens |
|---|---|---|
| 10-K | Årsredovisning | Årligen |
| 10-Q | Kvartalsrapport | Kvartalsvis |
| 8-K | Händelserapport | Vid behov |
| DEF 14A | Proxy statement | Årligen |

## Rate limit
- 10 req/s rekommenderat
- User-Agent krävs (med email)

## Auth
Ingen API-nyckel. Publikt API.

## Nästa steg
- [ ] Hämta CIK-lista
- [ ] Fetch company facts per ticker
- [ ] XBRL → gemensamt schema
