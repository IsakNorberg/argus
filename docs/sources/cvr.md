# CVR — Danmark

## API
- **Base:** `https://datacvr.virk.dk`
- **Enheder:** `/api/data?enhedstype=virksomhed`
- **Regnskaber:** `/api/regnskab/{cvr_nr}`

## Data typer
| Typ | Data |
|---|---|
| CVR-nummer | 8-siffrig orgnummer |
| Årsrapporter | XBR/IXBRL format |
| Ledelseserklæringer | Styrelse och ledning |

## Auth
Publikt. Krav: User-Agent header.

## Nästa steg
- [ ] Hämta CVR register
- [ ] Parsa XBRL-regnskaber
- [ ] Mappa mot Argus schema
