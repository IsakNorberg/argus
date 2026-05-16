# Argus 🏛️

> Den med 100 ögon — samlar, standardiserar och lagrar bolagsdata från alla källor.

## Syfte
Samla bolagsdata från flera källor (SEC, EODHD, m.fl.), standardisera till ett gemensamt format och lagra i databas.

## Källor
| Källa | Typ | Status |
|---|---|---|
| SEC (10-K, 10-Q, 8-K) | SEC filings | 📋 Todo |

## Struktur
- `docs/sources/` — MD-filer per datakälla med API-info
- `scripts/` — Scripts för datahämtning
- `data/` — Lokal datalagring (gitignore'd)
