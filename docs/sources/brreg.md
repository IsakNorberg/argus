# Brønnøysundregistrene (Brreg) — Norge

## API
- **Base:** `https://data.brreg.no`
- **Enhetsregisteret:** `/enhetsregisteret/api/enheter`
- **Regnskapsregisteret:** `/regnskapsregisteret/api/regnskap`
- **Selskapsregisteret:** `/selskapsregisteret/api/selskaper`

## Data typer
| Typ | Data |
|---|---|
| Organisasjonsnummer | 9-siffrig orgnummer |
| Årsregnskap | Balance sheet, resultatsregnskap |
| Styrelse | Ledamöter, roller |
| Eierskap | Ägarstruktur |

## Auth
Publikt API. Grundläggande auth krävs för vissa endpoints.

## Nästa steg
- [ ] Hämta orgnummer-lista
- [ ] Hämta årsregnskap per orgnr
- [ ] NUF = Mappa mot XBRL
