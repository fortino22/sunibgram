---
summary: Human intro to the Sunibgram agent wiki — what it is, how it's organized, and where to start.
last_synced: 2026-06-03
source:
  - docs/agent/SCHEMA.md
tags: [reference, readme, meta]
---

# Sunibgram Agent Wiki

An LLM-maintained knowledge base the **Fox** AI agent reads to guide users around
Sunibgram — answering "how do I access X?", "how do I use Y?", and "what are the rules?".
This is a wiki, not flat docs: the LLM writes and maintains every page; the human curates
the live product (the source of truth) and asks the questions.

Sunibgram is a verified-only campus social network for **Binus University students**
(`@binus.ac.id` email required). It has threads, replies, reposts, communities, events,
direct messages, bookmarks, a leaderboard, and Fox thread summaries.

## Where to start
- **[[index]]** — the content catalog. Every page, grouped, with a one-line summary.
  Read this first when answering a question.
- **[[SCHEMA]]** — the conventions and workflows (how to ingest a source change, answer a
  query, and lint the wiki). Read this before editing any page.
- **[[log]]** — the chronological record of what's changed.

## Layout
```
docs/agent/
  README.md      this file
  SCHEMA.md      conventions + ingest/query/lint workflows
  index.md       content catalog
  log.md         chronological record

  entities/      concrete product things (fox, threads, communities, events, …)
  concepts/      cross-cutting ideas (feed-ranking, privacy-model, eligibility, …)
  sources/       policy-text summaries (terms, privacy, faq, rules)
  reference/     lookup hubs (limits table, route catalogs, about overview)
```

## Scope
Covers **user-facing and public** routes only. Admin routes (`/admin/*`) and internal
auth API handlers (`/auth/*`) are intentionally excluded. The content syncs from the live
app, backend seeders, and domain models — full source-of-truth list is in [[SCHEMA]] §4.

**Support:** `support@sunibgram.com`.
