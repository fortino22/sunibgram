---
summary: Append-only chronological record of wiki activity — ingests, queries, and lint passes. Newest at the bottom.
last_synced: 2026-06-03
source:
  - docs/agent/SCHEMA.md
tags: [reference, log, meta]
---

# Log

Append-only. Each entry starts with `## [YYYY-MM-DD] <type> | <title>` so the log is
greppable (`grep "^## \[" log.md | tail -5`). Entry types: `ingest`, `query`, `lint`,
`restructure`. See [[SCHEMA]] §5.

## [2026-06-03] restructure | Reshaped agent KB into LLM-wiki layout
Reshaped the flat `README + routes/ + policies/` tree into an LLM-wiki:
`entities/` + `concepts/` + `sources/` + `reference/`, plus `index.md`, `log.md`, and
`SCHEMA.md`. Added `summary` / `last_synced` / `source` / `tags` frontmatter to every
page and converted all internal links to `[[wikilinks]]`. Content (facts, limits,
categories, routes) carried over unchanged; canonical facts deduplicated to a single
page each (numeric caps → [[limits]], report categories → [[moderation-and-reports]]).
