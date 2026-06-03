---
summary: Wiki conventions and ingest/query/lint workflows — the schema layer that turns this folder into an LLM-maintained knowledge base.
last_synced: 2026-06-03
source:
  - docs/agent/README.md
tags: [reference, schema, meta]
---

# Wiki Schema

This file is the **configuration** for the Sunibgram agent knowledge base. It tells the
LLM how the wiki is structured, what the page conventions are, and which workflows to
follow when ingesting source changes, answering questions, or health-checking the wiki.

The wiki is an LLM-maintained artifact: the LLM writes and maintains every page; the
human curates the sources (the live app + backend) and asks the questions. Read
[[index]] first to find pages, then drill in. Treat [[README]] as the human intro.

---

## 1. Layout

```
docs/agent/
  README.md      human intro + pointers
  SCHEMA.md      this file — conventions + workflows
  index.md       content catalog (every page, grouped, one-line summary + link)
  log.md         append-only chronological record

  entities/      concrete product things — one page each (fox, threads, communities, …)
  concepts/      cross-cutting ideas (feed-ranking, reply-permissions, privacy-model, …)
  sources/       policy-text summaries (terms, privacy, faq, rules) — the "source documents"
  reference/     lookup hubs (limits table, route catalogs, about overview)
```

- **entities/** — a noun a user names ("communities", "Fox", "premium"). Behaviour and
  how-to live here.
- **concepts/** — a cross-cutting rule or idea that spans entities (eligibility, feed
  ranking, moderation). When two+ entities share a rule, factor it into a concept page.
- **sources/** — plain-language summaries of the binding policy text (`/about/terms`,
  `/about/privacy`, `/about/help`). These mirror live copy closely.
- **reference/** — high-fan-in hubs: the single numeric [[limits]] table, the
  [[user-routes]] / [[public-routes]] catalogs, and the [[about]] overview. Other pages
  link **into** these instead of restating their contents.

Scope: **user-facing and public** only. Admin routes (`/admin/*`) and internal auth API
handlers (`/auth/*`) are intentionally excluded.

---

## 2. Page conventions

### Frontmatter — required on every page
```yaml
---
summary: One-line description, reused verbatim in index.md.
last_synced: 2026-06-03
source:
  - frontend/app/about/fox/page.tsx
  - backend/domain/model/thread/thread.go
tags: [entity, ai, fox]
---
```
- `summary` — a single line. [[index]] pulls this, so editing it here keeps the catalog
  in sync. Don't let them drift.
- `last_synced` — `YYYY-MM-DD`, the date this page was last reconciled against its source.
  Per-page (no global date).
- `source` — the code/page paths this page syncs from. May also appear inline as a
  `> Source:` line where it aids reading.
- `tags` — **first tag is the kind** (`entity` / `concept` / `source` / `reference`),
  remaining tags are topical (`ai`, `safety`, `messaging`…).

### Links
- Link liberally with `[[slug]]`, where `slug` is the target page's filename without
  `.md` (e.g. `[[communities]]`, `[[feed-ranking]]`). Folder is omitted — slugs are unique.
- A `[[slug]]` with no matching file yet is fine: it marks a page worth writing later.
- **Don't duplicate canonical facts.** Each fact lives on exactly one page; everything
  else links to it. Numeric caps live only in [[limits]]; report categories only in
  [[moderation-and-reports]]; the faculty list only in [[faculties]]; etc.

### Tone
Full sentences, agent-facing prose. These pages are read by the Fox agent and by humans —
not compressed. (Caveman/compressed style does not apply to wiki content.)

---

## 3. Workflows

### Ingest — a source changed
Trigger: live `/about/*` copy, a route, a seeder, or a domain-model limit changed.
1. Read the changed source (paths are in each page's `source` frontmatter + [[README]]).
2. Update the matching entity / concept / source page. If a new noun or rule appeared
   that no page covers, create a page in the right folder with full frontmatter.
3. Refresh that page's `last_synced`.
4. If its `summary` changed, update the matching line in [[index]].
5. Append a [[log]] entry (format in §5).
A single source can touch several pages — e.g. a threads change hits [[threads]],
[[reply-permissions]], and [[limits]]. Update all of them in one pass.

### Query — someone asks a question
1. Read [[index]] first to locate relevant pages.
2. Drill into those pages; follow `[[links]]` to canonical facts.
3. Answer with **route links** (from [[user-routes]] / [[public-routes]]) and **policy
   citations** (from sources/). Flag premium-gated actions instead of promising they work
   for everyone.
4. If the answer is a reusable synthesis (a comparison, a new how-to), file it back as a
   new page so the exploration compounds — don't let it vanish into chat.

### Lint — health check
Periodically scan for:
- **Contradictions** between pages.
- **Stale `last_synced`** — pages older than the live copy they mirror.
- **Orphans** — pages with no inbound `[[links]]` from any other page.
- **Missing pages** — a concept/entity mentioned repeatedly but lacking its own page.
- **Missing cross-refs** — related pages that should link but don't.
- **Dangling links** — `[[slug]]` with no matching file (either create the page or fix
  the link).
Report findings + suggested fixes; append a lint entry to [[log]].

---

## 4. Maintenance sources (source of truth)

The content is derived from the live product; this wiki only restructures it. Re-sync
when these change:
- Routes → `frontend/app/**/page.tsx` (folder path = URL path)
- FAQ / Terms / Privacy → `frontend/app/about/{help,terms,privacy}/page.tsx`
- Purpose / principles → `frontend/app/about/{our-story,company}/page.tsx`
- Fox / Hermes → `frontend/app/about/{fox,hermes}/page.tsx`
- Feed ("For You") ranking → `frontend/app/about/algorithm/page.tsx`
- News → `frontend/lib/news.ts` + `frontend/app/about/news/**`
- Report categories → `backend/infrastructure/seeder/violation_seeder.go`
- Community categories → `backend/infrastructure/seeder/category_seeder.go`
- Faculties → `backend/infrastructure/seeder/faculty_seeder.go`
- Trend-report reasons → `backend/infrastructure/seeder/trend_report_reason_seeder.go`
- Limits → domain models (`thread_poll`, `media`, `thread_media`, `user`) + frontend
  services/components (`threads.ts`, `bookmarkCollections.ts`, `feedback.ts`, community/
  chat modals); premium → `usecase/order` + `usecase/premium_subscription`

**Support:** `support@sunibgram.com`.

---

## 5. Log format

Append-only, newest at the bottom. Each entry starts with a consistent prefix so the log
is greppable:

```
## [YYYY-MM-DD] <ingest|query|lint|restructure> | <title>
```

`grep "^## \[" docs/agent/log.md | tail -5` gives the last five events.
