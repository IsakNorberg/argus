# SEC — U.S. Securities and Exchange Commission

## Källor
- **EDGAR Full-Text Search:** https://efts.sec.gov/LATEST/search-index
- **Filings API:** https://www.sec.gov/cgi-bin/browse-edgar?action=getcompany
- **Company Facts:** https://data.sec.gov/submissions/CIK{cik}.json

## Data typer
| Typ | Beskrivning | Frekvens |
|---|---|---|
| 10-K | Årsredovisning | Årligen |
| 10-Q | Kvartalsrapport | Kvartalsvis |
| 8-K | Händelserapport | Vid behov |
| DEF 14A | Proxy statement | Årligen |

## Rate limit
- 10 requests/second (rekommenderat)
- User-Agent header krävs (med email)

## Authentication
Ingen API-nyckel krävs för publika endpoints.

## XBRL
SEC filings använder XBRL-format. Facts API:et ger strukturerad data direkt.
- Base: `https://data.sec.gov/submissions/`
- Fil: `CIK{cik}.json` (cik padad till 10 siffror, t.ex. CIK0000320193 för Apple)

## Nästa steg
- [ ] Mappa XBRL → gemensamt schema
- [ ] Bygg parser för filing-dokument
- [ ] Definiera standardiserade fält
