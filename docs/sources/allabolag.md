# Allabolag.se — Svensk Bolagsinfo

## API:er
- **Bas:** `https://www.allabolag.se/`
- **Scraping:** HTML-sidor (inget officiellt API)

## Data typer
| Typ | Data |
|---|---|
| Bolagsfakta | Orgnr, namn, adress, VD |
| Styrelse | Ledamöter, roller |
| Årsredovisningar | Nyckeltal (PDF) |
| Konkurser/Räddningar | Historik |

## Auth
Ingen. Scraping enligt ToS.

## Nästa steg
- [ ] Bygg scraper (colly.go)
- [ ] Extrahera orgnr → bolagsinfo
- [ ] Respektera rate limits
