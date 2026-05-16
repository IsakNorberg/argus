# ESEF — European Single Electronic Format

## Beskrivning
ESEF är EU:s gemensamma XBRL-standard för årsredovisningar. Alla bolag på reglerade marknader måste rapportera i ESEF-format sedan 2020.

## Källor
- **ESMA:** `https://www.esma.europa.eu`
- **ESEF Taxonomy:** `https://www.esma.europa.eu/esef-taxonomy`
- **Filer:** `.xhtml` med XBRL-taggning

## Data typer
| Typ | Data |
|---|---|
| iXBRL | Inline XBRL — läsbar + maskinläsbar |
| Taxonomi | Gemensamma fält (revenue, assets, etc) |
| Extension | Bolagsspecifika tags (sällan, men finns) |

## Viktigt
ESEF är **nyckeln** för att standardisera alla europeiska datakällor.
Alla europeiska registrer börjar använda ESEF — vi kan parsera med gemensam logik.

## Nästa steg
- [ ] Hitta ESEF-filer från nationella register
- [ ] Bygg XBRL-parser
- [ ] Mappa ESEF taxonomy → Argus schema
