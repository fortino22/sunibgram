# Sunibgram Agent Knowledge Base

Reference docs the **Fox** AI agent reads to guide users around Sunibgram — answering
"how do I access X?", "how do I use Y?", and "what are the rules?".

Sunibgram is a verified-only campus social network for **Binus University students**
(`@binus.ac.id` email required). It has threads, replies, reposts, communities,
events, direct messages, bookmarks, a leaderboard, and Fox thread summaries.

## How the agent should use these files

- **"Where / how do I …?"** → look up `routes/` and return the matching path as a
  clickable link. Paths are Next.js routes on the frontend origin (e.g.
  `https://<app-host>/explore`). The agent should output the path; the UI renders it
  clickable for the user.
- **"What are the rules / can I … / how is my data used?"** → cite `policies/`.
- Dynamic segments are shown as `:param` (e.g. `/thread/:id`, `/profile/:userId`).
  The agent fills these in with a real ID when it knows one; otherwise it tells the
  user how to reach the page (e.g. "tap a post to open its thread").
- **Premium-gated** actions are flagged. The agent should mention the premium
  requirement instead of promising the action will work for everyone.

## Layout

```
docs/agent/
  README.md              ← this file
  routes/
    user-routes.md       ← signed-in, user-facing routes (grouped by category)
    public-routes.md     ← auth + /about pages (no login required)
  policies/
    faq.md               ← top questions + answers
    terms.md             ← terms & conditions
    privacy.md           ← privacy policy
    rules.md             ← eligibility + acceptable use + report categories
    about.md             ← what Sunibgram is, principles, faculties, support
    limits.md            ← numeric/format limits (char/size/count caps, password, OTP)
    features.md          ← feature capabilities & how-to (reply perms, roles, RSVP…)
```

When a user asks "how long / how big / how many / what file types?" → cite
`limits.md`. For "how do I use X / can I …?" → cite `features.md`.

## Scope & maintenance

- Covers **user-facing and public** routes only. Admin routes (`/admin/*`) and
  internal auth API handlers (`/auth/*`) are intentionally excluded.
- **Source of truth** for this content:
  - Routes → `frontend/app/**/page.tsx` (folder path = URL path)
  - FAQ → `frontend/app/about/help/page.tsx`
  - Terms → `frontend/app/about/terms/page.tsx`
  - Privacy → `frontend/app/about/privacy/page.tsx`
  - Purpose / principles → `frontend/app/about/what/page.tsx`, `.../team/page.tsx`
  - Report categories → `backend/infrastructure/seeder/violation_seeder.go`
  - Community categories → `backend/infrastructure/seeder/category_seeder.go`
  - Faculties → `backend/infrastructure/seeder/faculty_seeder.go`
  - Trend-report reasons → `backend/infrastructure/seeder/trend_report_reason_seeder.go`
  - Limits → domain models (`thread_poll`, `media`, `thread_media`, `user`) + frontend
    services/components (`threads.ts`, `bookmarkCollections.ts`, `feedback.ts`,
    community/chat modals); premium → `usecase/order` + `usecase/premium_subscription`
- When the live `/about/*` copy, routes, or these limits change, re-sync these files.
- **Last synced:** 2026-06-01.
- **Support:** `support@sunibgram.com`.
