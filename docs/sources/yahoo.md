# Yahoo Finance — DAGENS KURS

## Syfte
Kompletterande källa för **dagens aktiekurs** och **historisk prisdata**.
Används INTE för balansräkningar eller fundamentals.

## API:er
- **Quote:** `GET /v6/finance/quote?symbols=TICKER`
- **Chart:** `GET /v8/finance/chart/{ticker}?range=1d&interval=1d`

## Data
| Fält | Beskrivning |
|---|---|
| regularMarketPrice | Dagens kurs |
| currency | Valuta (USD, SEK, etc) |
| marketCap | Börsvärde |
| volume | Dagsvolym |
| regularMarketChangePercent | Dagens förändring % |

## Auth
Ingen. Oofficiellt API — kan ändras utan avisering.

## Nästa steg
- [ ] Implementera FetchQuote
- [ ] Implementera FetchHistoricalPrices
- [ ] Lägg till price_quotes-tabell i databasen
