# INPI — Institut National de la Propriété Industrielle (Frankrike)

## API:er
- **Base:** `https://data.inpi.fr`
- **BODACC:** Bolagskungörelser
- **Sirene (INSEE):** `https://sirene.insee.fr`

## Data typer
| Typ | Data |
|---|---|
| SIRET/SIREN | 9-siffrig orgnummer |
| Comptes annuels | Årsredovisningar |
| BODACC | Kungörelser om bolagsändringar |

## Auth
Publikt. Krav: User-Agent header.

## Nästa steg
- [ ] Hämta SIREN-lista
- [ ] Parsa årsredovisningar
- [ ] Mappa mot Argus schema
