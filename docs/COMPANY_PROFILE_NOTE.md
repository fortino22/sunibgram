# Sunibgram — Company Profile & Copy Deck

> Source of truth for the public **About / company profile** page. Covers brand identity, voice, marketing-page visual system, and section-by-section copy.
>
> **Reference style:** Composio (composio.dev) — light, clean, monospace-accented UI chrome, large inline-element headlines, brand-tinted use-case grid, floating concept bubbles, multi-column categorized footer.
>
> **Important:** The marketing page intentionally **does not match the in-app UI**. In-app is dark-first; the marketing page is light-first. The accent blue stays consistent across both.

---

## 1. At a Glance

**Sunibgram** is a university-exclusive social platform built for students of Binus University. It combines the short-form posting model of X / Threads with campus-native primitives — faculties, communities, events, verified-only identity, real-time chat, leaderboard, and bookmark collections — inside a single closed ecosystem.

- **Category:** Closed-community social network for higher education
- **Primary audience:** Verified Binus University students (`@binus.ac.id`)
- **Form factor:** Web application (Next.js), real-time backend (Go)
- **Status:** First-cut shipped across all major surfaces; ongoing work focuses on depth, polish, and real-time refinement

---

## 2. Positioning Statement

Sunibgram is the campus social layer for Binus University: a verified-only space where students discuss, organize, and discover — without the noise, anonymity, and trust gap of open social networks.

Every account is tied to a real, current student. Every community, event, and thread happens in a known context. The result is a higher-signal, lower-risk environment purpose-built for university life.

---

## 3. Audience

| Segment | Description |
|---|---|
| Binus students | Undergraduates and graduates with a valid `@binus.ac.id` email |
| Community organizers | Students running clubs, study groups, and interest circles |
| Event creators | Students publishing official or unofficial campus events |
| University administrators | Staff approving official events and moderating reports |

---

## 4. Brand Identity

### Name
**Sunibgram** — a composite blending a student-facing handle (*Suni*), the host institution (*Binus*), and the *-gram* suffix familiar from social-media vernacular.

### Voice & Tone
- **Confident, not loud.** The platform's value is restraint — verified-only, focused, signal-rich.
- **Plain-spoken.** Short sentences. No jargon. No exclamation marks.
- **Student-first.** Talks about campus life, not "the platform" or "the community."
- **Trustworthy by default.** Identity, safety, and moderation are framed as features, not policies.

### Marketing-Page Visual System (light-first)

> Distinct from the in-app dark theme. Apply only inside `/about` (and any other public marketing surfaces).

**Palette**

| Token | Role | Value |
|---|---|---|
| `--mkt-bg` | Page background | `#ffffff` |
| `--mkt-bg-grid` | Subtle hero mesh background | `radial-gradient(...)` cyan → white |
| `--mkt-card` | Card / panel surface | `oklch(98% 0.005 220)` |
| `--mkt-card-tint` | Use-case card tint | `oklch(95% 0.04 200)` — light cyan |
| `--mkt-accent` | Primary brand color (shared with app) | `oklch(62% 0.2 260)` |
| `--mkt-accent-soft` | Eyebrow background, soft highlights | `oklch(95% 0.05 260)` |
| `--mkt-text` | Primary text | `oklch(15% 0.01 260)` — near-black |
| `--mkt-text-muted` | Secondary text | `oklch(45% 0.01 260)` — medium gray |
| `--mkt-text-tertiary` | Fine print, captions | `oklch(60% 0.008 260)` |
| `--mkt-border` | Hairline borders | `oklch(92% 0.005 260)` |
| `--mkt-button` | Primary CTA background | `#000000` (Composio uses near-black for primary CTAs) |
| `--mkt-button-text` | Primary CTA text | `#ffffff` |

**Typography**
- **Display & body:** DM Sans (already loaded project-wide)
- **UI chrome (nav, eyebrows, CTAs, footer labels):** system monospace — `ui-monospace, 'SF Mono', 'Cascadia Mono', 'Roboto Mono', monospace`. No new font dependency; same loaded fonts as the rest of the project.
- **Headline scale:** clamp from ~3rem on mobile to ~5.5rem on desktop
- **Headline weight:** regular to medium (not bold) — clean and roomy, not aggressive

**Iconography**
- Lucide line icons — clean stroke
- Real product UI fragments where they communicate better than icons (e.g., a thread card, an avatar chip, a community card)

---

## 5. Section Composition Rules (the Composio pattern)

Every marketing block follows the same four-beat structure:

1. **Eyebrow** — 2-5 words, mono caps, accent color, prefixed with a small `■` marker. States the *category* of the block (e.g., `■ ONE CAMPUS, ONE PLACE.`).
2. **Headline** — ≤ 6 words. DM Sans, large. Allow inline visual substitutions (a logo, avatar chip, or label badge inside the headline) where it adds clarity.
3. **Subhead** — exactly two short sentences, ≤ 240 characters total. `--mkt-text-muted`.
4. **CTA** — one verb-led monospace button per block. Always the primary action (no decorative buttons).

**Visual half:** every section pairs the copy half with a *right-side visual half* — a product UI fragment, a constellation of floating bubbles, or a card stack. Never decorative-only.

---

## 6. Page Composition — Section-by-Section Copy Deck

The `/about` page composes the blocks below in order. Each block follows the rules in §5.

### 6.1 Hero

- **Eyebrow:** (none — the hero is the eyebrow for the page)
- **Headline (inline element):** *Make your* `[binus account]` *do more.*
  - Inline element: a small profile-chip mockup (avatar + `@username` + faculty badge) inserted mid-headline, like Composio drops `CODEX` into theirs.
- **Subhead:** Skip the noise. Sunibgram is the verified social layer for Binus students — threads, communities, events, chat, all inside one campus.
- **CTA:** `JOIN WITH YOUR BINUS EMAIL`
- **Visual:** soft cyan-to-white radial mesh covering the hero band; no orbs, no glow rings.

### 6.2 "Never just a chat" — primary value

Mirrors Composio's "Never leave the chat."

- **Eyebrow:** `■ ONE CAMPUS. ONE PLACE.`
- **Headline:** Threads, communities, events — without leaving campus.
- **Subhead:** Sunibgram pulls every campus conversation into one feed, verified by your `@binus.ac.id` and organized by the faculty you actually belong to.
- **CTA:** `Get started for free →`
- **Visual (right):** a stacked card list mockup — a thread card on top, a community card behind it, an event card behind that, each subtly tinted. Two cards have a small checkmark "joined" indicator.

### 6.3 "One identity, every surface" — connection point

Mirrors Composio's "One connection, every app."

- **Eyebrow:** `■ VERIFIED IDENTITY`
- **Headline:** One sign-in, every campus surface.
- **Subhead:** Threads, chats, communities, leaderboards, events, and bookmarks all share one verified Binus identity. No second account, no second password.
- **CTA:** `Get started for free →`
- **Visual (right):** a centered avatar with floating labelled chips orbiting around it — `Threads`, `Chats`, `Events`, `Communities`, `Leaderboard`, `Bookmarks`, `Profile`, `Faculty`. Mirrors Composio's app-sphere visual.

### 6.4 "Built for how Binus students live" — use-case grid

Mirrors Composio's "Built for how you work."

- **Eyebrow:** `■ USE CASES`
- **Headline:** Built for how Binus students live.
- **Subhead:** Sunibgram works for every campus role — with the surfaces, communities, and prompts already tuned for how you study, organize, and connect.
- **CTA:** `All use cases →`
- **Card grid (4 brand-tinted cards in a row, scrollable on mobile):**

| Card | Title | Description | Inline icons |
|---|---|---|---|
| 1 | Everyday students | Post, follow, save what matters — and see what your faculty is talking about. | thread, repost, bookmark, like |
| 2 | Community organizers | Run clubs, study groups, and interest circles with built-in roles and request flows. | communities, members, rules, threads |
| 3 | Event creators | Publish official and unofficial campus events with attachments and document previews. | event, attachment, threads, schedule |
| 4 | Faculty discovery | See what your faculty is talking about, find peers, and build connection from day one. | faculty, connect, leaderboard, mention |

Card background uses `--mkt-card-tint`. Each card has an `EXPLORE →` link at the bottom.

### 6.5 "Recognition that's real" — stat moment

Mirrors Composio's "Accelerate your AI productivity" + the `100,000+` stat panel.

- **Eyebrow (left):** `■ FEED THAT GETS YOU`
- **Headline (left):** Recognition that's real.
- **Subhead (left):** The For You feed mixes fresh posts with engagement-weighted ranking. The Leaderboard rotates trending students every refresh window. Participation isn't lost — it's recognized.
- **CTA (left):** `Get started for free →`
- **Stat block (right):**
  - Eyebrow above number: `STUDENTS ON SUNIBGRAM`
  - Number: `1.000+` (placeholder — replace with real number at publish time)
  - Caption below: `Shipping faster with their voices.`

### 6.6 "Multiple roles, one student" — depth moment

Mirrors Composio's "Multiple accounts per app."

- **Eyebrow:** `■ ROLES THAT FOLLOW YOU`
- **Headline:** Multiple roles, one student.
- **Subhead:** Be a member in one community, moderator in another, owner of your own. Sunibgram lets you wear every campus hat — without juggling logins or profiles.
- **CTA:** `Get started for free →`
- **Visual (right):** a single avatar chip branching out to three role badges — `OWNER`, `MODERATOR`, `MEMBER` — each labelled with the community it belongs to.

### 6.7 "Floating campus actions" — capability sweep (optional, drops between §6.5 and §6.6 if vertical space allows)

Mirrors Composio's floating-bubbles section around "YOUR AI AGENT."

- Center label: `YOUR CAMPUS PROFILE`
- Floating bubbles around it:
  - `Reply to your faculty thread`
  - `Pin your best post`
  - `Save to a collection`
  - `Trend on the leaderboard`
  - `Join a community`
  - `RSVP to an event`
- No CTA — it's a capability poster.

### 6.8 Final CTA

Mirrors Composio's "Your agents decide. We make it happen."

- **Headline (single bold line, mono accent):** *Your campus, your voice. We connect the wires.*
- **CTAs (two buttons, side by side):**
  - Secondary outline: `BOOK A DEMO`
  - Primary filled: `JOIN SUNIBGRAM TODAY`
- **Background:** dark — same near-black as the primary CTA button. Sets a hard visual closure before the footer.

### 6.9 Footer

Mirrors Composio's multi-column categorized footer. Background `--sidebar` (the dark sidebar token).

- **Left:** logo mark + `SUNIBGRAM` wordmark in mono caps.
- **Five columns:**

| PRODUCT | FOR STUDENTS | FOR ORGANIZERS | RESOURCES | COMPANY |
|---|---|---|---|---|
| Threads | Sign up | Create community | Docs | About |
| Communities | Sign in | Host an event | Roadmap | Trust & safety |
| Events | Faculty feed | Become moderator | Changelog | Contact |
| Chat | Discover | Admin handbook | Brand kit | Privacy |
| Leaderboard | | | | Terms |
| Connect | | | | |
| Bookmarks | | | | |
| Premium | | | | |

- **Bottom row:** social icons (X, GitHub, LinkedIn, YouTube placeholders) on the left, `© Sunibgram 2026` on the right.

---

## 7. Working Headlines & Taglines (Composio voice)

- *Make your Binus account do more.*
- *Verified students. Real conversations.*
- *Threads. Communities. Events. One campus.*
- *Your campus, your voice.*
- *Skip the noise. Connect with your campus.*

Final tagline selection should follow stakeholder review.

---

## 8. Maturity & Roadmap

| Phase | Scope | Status |
|---|---|---|
| 1 | Authentication, registration, OTP, password recovery | Shipped |
| 2 | Threads — feed, create, reply, polls, like, bookmark, repost, pin | Shipped |
| 3 | Profiles, follow graph, image cropping | Shipped |
| 4 | Thread context menu, chain view, AI summary | Shipped |
| 5 | Communities (full lifecycle) | Shipped |
| 6 | Notifications and real-time delivery | Shipped (real-time refinement ongoing) |
| 7 | Explore and unified search | Shipped |
| 8 | Chats (private and group rooms) | Shipped |
| 9 | Events with admin approval workflow, document preview | Shipped |
| 10 | Premium subscriptions and payments | Shipped |
| 11 | Reports and violations | Shipped |
| 12 | Trending hashtags, mention autocomplete | Shipped |
| 13 | User settings — privacy, appearance (light/dark), subscription | Shipped |
| 14 | Leaderboard and trending users | Shipped |
| 15 | Connect — mutual and faculty-based suggestions | Shipped |
| 16 | Bookmark collections | Shipped |
| 17 | In-product feedback (bugs and feature requests) with admin triage | Shipped |
| 18 | Admin trending analytics with visualized reason breakdowns | Shipped |

---

## 9. Glossary

- **Thread** — A post. Text, media, polls, and hashtags. Can be replied to, reposted, liked, bookmarked, and pinned.
- **Chain** — A thread and its full descendant reply tree, rendered as a single conversation.
- **Faculty** — An academic division within Binus University. Used as a profile attribute and as a feed filter.
- **Community** — A user-formed group, public or private, with its own roles, rules, and posting surface.
- **Event** — A scheduled campus happening. Official (admin-approved with attachments) or unofficial (student-published).
- **For You** — The engagement-ranked default feed.
- **Premium** — A paid subscription tier with profile recognition.
- **Connect** — The dedicated people-discovery surface (mutual + faculty suggestions).
- **Leaderboard** — The trending-users ranking surface with podium and personal stats.
- **Collection** — A named group of bookmarked threads for curation.
