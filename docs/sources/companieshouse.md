# Companies House — Storbritannien

## API
- **Base:** `https://api.company-information.service.gov.uk`
- **Companies:** `/company/{company_number}`
- **Filing History:** `/company/{company_number}/filing-history`
- **Officers:** `/company/{company_number}/officers`
- **Filing Documents:** `/company/{company_number}/filing-history/{transaction_id}/document/{content_hash}/content`

## Data typer
| Typ | Data |
|---|---|
| Company Number | 8-siffrig ( Companies House number) |
| Accounts | Full accounts, micro accounts |
| Confirmation Statement | Årlig statusrapport |
| Charges | Pantbelåningar |

## Auth | OAuth2 eller Basic Auth med API-nyckel.
- Registrera app: `https://developer.company-information.service.gov.uk`

## Nästa steg
- [ ] Skaffa API-nyckel
- [ ] Hämta company list
- [ ] Parsa accounts (många är PDF eller iXBRL)
