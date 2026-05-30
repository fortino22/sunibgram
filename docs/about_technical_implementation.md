# About Page — Technical Implementation Plan

> Public, full-width marketing page for Sunibgram. Accessible without login.
> **Design reference:** Composio (composio.dev) — **light**, clean, monospace-accented UI chrome, large inline-element headlines, brand-tinted use-case cards, floating-bubble visuals, multi-column footer.
>
> This document is the build spec. Copy lives in `COMPANY_PROFILE_NOTE.md` (§6 — Page Composition).

---

## 0. What is changing vs. the previous spec

The previous version of this doc described a **dark, Orbitron-heavy, gradient-glow** about page — and the corresponding components were built (`AboutHero`, `AboutFeatures`, `AboutMission`, etc., plus effect components `ScrollReveal`, `AnimatedCounter`, `SpotlightCard`, `HeroMouseGlow`).

The new direction is the **Composio aesthetic**:

| Old | New |
|---|---|
| Dark page background | Light page background (white) |
| Orbitron display font | DM Sans display + JetBrains Mono for UI chrome |
| Gradient text effects | Flat text; one accent color only |
| Glow radials and ambient orbs | Subtle cyan-to-white mesh in the hero only |
| 6-card feature grid with icons | 4 brand-tinted use-case cards with app-icon rows |
| Centered mission with blockquote | Stat-block moment with one large number |
| 3-step "How it works" numbered flow | Removed — collapsed into the value-prop split sections |
| Minimal one-line footer | Multi-column categorized footer (5 columns) |

Existing components are **kept but rewritten in place** (one PR per component is fine). Effect components (`ScrollReveal`, `AnimatedCounter`, `SpotlightCard`, `HeroMouseGlow`) are **removed** — the new design doesn't use them.

---

## 1. Routing & Layout

`/about` already bypasses the sidebar via `AUTH_PATHS` in `components/layout/MainWrapper.tsx`. No router change needed.

The page must opt out of the in-app dark theme. Strategy:

1. Wrap the page in a single `<div className="about-root">` that carries the marketing-palette CSS variables (light tokens), overriding the in-app dark tokens for the subtree.
2. All marketing components consume `--mkt-*` tokens, not the global `--bg`/`--text` tokens.

---

## 2. File Structure

```
app/
  about/
    page.tsx                       ← route entry, metadata, composition

components/
  about/
    AboutNav.tsx                   ← sticky top bar — logo + product/solutions/toolkits/blog/docs + GET STARTED
    AboutHero.tsx                  ← inline-element headline ("Make your [chip] do more"), CTA, mesh bg
    AboutSplitValue.tsx            ← "Never just a chat" — left copy + right card-stack mockup
    AboutSplitConnection.tsx       ← "One identity, every surface" — left copy + right orbiting-chips visual
    AboutUseCases.tsx              ← "Built for how Binus students live" — 4 brand-tinted cards
    AboutStatMoment.tsx            ← "Recognition that's real" — left copy + right big-stat block
    AboutFloatingBubbles.tsx       ← optional: "YOUR CAMPUS PROFILE" centerpiece with floating capability chips
    AboutSplitRoles.tsx            ← "Multiple roles, one student" — left copy + right role-branch visual
    AboutFinalCTA.tsx              ← dark band: "Your campus, your voice. We connect the wires." + 2 buttons
    AboutFooter.tsx                ← 5-column categorized footer
```

**Removed (delete after migration):**
- `AboutStats.tsx` (replaced by `AboutStatMoment.tsx` — one big number, not four)
- `AboutFeatures.tsx` (replaced by `AboutUseCases.tsx` — 4 tinted cards, not 6 icon cards)
- `AboutHowItWorks.tsx` (cut from the page)
- `AboutMission.tsx` (cut — its purpose is absorbed into the split sections)
- `AboutCTA.tsx` (replaced by `AboutFinalCTA.tsx`)
- `ScrollReveal.tsx`, `AnimatedCounter.tsx`, `SpotlightCard.tsx`, `HeroMouseGlow.tsx` (no longer needed)

All marketing components are **Server Components** (no `'use client'`) unless a section needs interactivity. The new design has no interactive sections.

---

## 3. globals.css Additions

Append a marketing-palette block scoped to `.about-root`. This prevents the marketing tokens from leaking into the rest of the app.

```css
/* About page — marketing palette (light-first, distinct from in-app dark theme) */
.about-root {
  --mkt-bg:               #ffffff;
  --mkt-card:             oklch(98% 0.005 220);
  --mkt-card-tint:        oklch(95% 0.04 200);
  --mkt-accent:           var(--accent);           /* shared brand blue */
  --mkt-accent-soft:      oklch(95% 0.05 260);
  --mkt-text:             oklch(15% 0.01 260);
  --mkt-text-muted:       oklch(45% 0.01 260);
  --mkt-text-tertiary:    oklch(60% 0.008 260);
  --mkt-border:           oklch(92% 0.005 260);
  --mkt-button:           #000000;
  --mkt-button-text:      #ffffff;
  --font-mkt-mono:        ui-monospace, 'SF Mono', 'Cascadia Mono', 'Roboto Mono', monospace;

  background: var(--mkt-bg);
  color: var(--mkt-text);
  font-family: var(--font-dm-sans);
}

/* Hero mesh background (used by AboutHero only) */
.about-hero-mesh {
  background:
    radial-gradient(60% 50% at 30% 40%, oklch(95% 0.08 220 / 0.6) 0%, transparent 60%),
    radial-gradient(50% 40% at 70% 50%, oklch(95% 0.06 260 / 0.55) 0%, transparent 65%),
    linear-gradient(0deg, var(--mkt-bg), var(--mkt-bg));
  background-size: cover;
}

/* Monospace UI chrome — nav, eyebrows, CTAs, footer labels */
.about-root .mono {
  font-family: var(--font-mkt-mono);
  letter-spacing: 0.02em;
}

/* Eyebrow tag: ■ EYEBROW TEXT */
.about-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: var(--font-mkt-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--mkt-accent);
}
.about-eyebrow::before {
  content: '';
  display: inline-block;
  width: 6px;
  height: 6px;
  background: var(--mkt-accent);
  border-radius: 1px;
}

/* Primary marketing button */
.about-btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  background: var(--mkt-button);
  color: var(--mkt-button-text);
  font-family: var(--font-mkt-mono);
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-radius: 4px;
  border: 1px solid var(--mkt-button);
  transition: opacity 0.15s ease;
}
.about-btn-primary:hover { opacity: 0.85; }

/* Secondary marketing button (outline on light) */
.about-btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  background: transparent;
  color: var(--mkt-text);
  font-family: var(--font-mkt-mono);
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-radius: 4px;
  border: 1px solid var(--mkt-border);
  transition: background-color 0.15s ease;
}
.about-btn-secondary:hover { background: var(--mkt-card); }

/* Use-case card — brand-tinted, no hover glow */
.about-usecase-card {
  background: var(--mkt-card-tint);
  border-radius: 12px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 240px;
}
```

(Tailwind handles all layout/spacing. Colors and typography come from these variables.)

---

## 4. Page Sections

### 4.1 `app/about/page.tsx`

```tsx
import type { Metadata } from 'next';
import { AboutNav } from '@/components/about/AboutNav';
import { AboutHero } from '@/components/about/AboutHero';
import { AboutSplitValue } from '@/components/about/AboutSplitValue';
import { AboutSplitConnection } from '@/components/about/AboutSplitConnection';
import { AboutUseCases } from '@/components/about/AboutUseCases';
import { AboutStatMoment } from '@/components/about/AboutStatMoment';
import { AboutSplitRoles } from '@/components/about/AboutSplitRoles';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';
import { AboutFooter } from '@/components/about/AboutFooter';

export const metadata: Metadata = {
  title: 'About — Sunibgram',
  description:
    'Sunibgram is the verified social layer for Binus University students — threads, communities, events, and chat inside one campus.',
};

export default function AboutPage() {
  return (
    <div className="about-root">
      <AboutNav />
      <AboutHero />
      <AboutSplitValue />
      <AboutSplitConnection />
      <AboutUseCases />
      <AboutStatMoment />
      <AboutSplitRoles />
      <AboutFinalCTA />
      <AboutFooter />
    </div>
  );
}
```

### 4.2 `AboutNav`

Sticky top bar. Mirrors Composio's nav layout.

- **Left:** logo mark + `Composio`-style wordmark (here: `Sunibgram` in DM Sans medium) + a small mono pill tag `FOR YOU` ← (decorative, matches reference).
- **Center / right:** mono caps links — `PRODUCT`, `COMMUNITIES`, `EVENTS`, `BLOG`, `DOCS` — and a primary `GET STARTED` filled mono button on the far right.
- **Background:** `var(--mkt-bg)`, hairline bottom border `var(--mkt-border)`, `position: sticky; top: 0; z-index: 50`.

### 4.3 `AboutHero`

Full-width hero. The page's primary impression.

- **Background:** `about-hero-mesh` class on a relative-positioned outer div.
- **Headline (inline element pattern):** centered, DM Sans, regular weight, clamp(2.5rem, 6vw, 5rem).
  > `Make your` `[ binus account ]` `do more.`
  - The `[ binus account ]` chip is a small inline component: avatar circle + `@username` + a tiny mono `BINUS` badge, all inside a rounded white card with a 1px border. Renders inline with the text via `display: inline-flex; vertical-align: middle`.
- **Subhead:** ~`1.05rem`, `var(--mkt-text-muted)`, max 36rem wide, centered.
- **CTA:** single primary `about-btn-primary` — `GET STARTED FOR FREE`.
- **Padding:** `~120px` top, `~80px` bottom; full-bleed background.

### 4.4 `AboutSplitValue` — "Never just a chat"

Two-column section, copy on left, mockup on right.

- **Left column:**
  - Eyebrow: `■ ONE CAMPUS. ONE PLACE.`
  - Headline: `Threads, communities, events — without leaving campus.`
  - Subhead (2 sentences).
  - CTA: `Get started for free →` — uses `about-btn-secondary` styled as a text link with mono caps.
- **Right column:** a small product mockup card titled `AGENT WORKING` (matches Composio's pattern) — but renamed `CAMPUS ACTIVITY`:
  - Header row: `CAMPUS ACTIVITY` (mono caps, muted) + counter `02 / 06` on the right.
  - Six rows, each: a status dot (filled / hollow), a tiny app-icon stack (thread / community / event icons), and a one-line task description in muted text. First two rows have completed checkmarks; rest are pending.

### 4.5 `AboutSplitConnection` — "One identity, every surface"

Two-column section, copy on left, "constellation" visual on right.

- **Left column:**
  - Eyebrow: `■ VERIFIED IDENTITY`
  - Headline: `One sign-in, every campus surface.`
  - Subhead.
  - CTA: `Get started for free →`
- **Right column:** dotted-sphere graphic with floating labelled chips placed at various positions — each chip is a rounded white card with a Lucide icon + label:
  - `Threads`, `Chats`, `Events`, `Communities`, `Leaderboard`, `Bookmarks`, `Profile`, `Faculty`.
- The dotted sphere is a static SVG. No JS animation.

### 4.6 `AboutUseCases` — "Built for how Binus students live"

Brand-tinted card grid section.

- **Section header (top):**
  - Eyebrow: `■ USE CASES`
  - Headline (left-aligned): `Built for how Binus students live.`
  - Subhead.
  - Right side of header: a mono link `All use cases ↗` + left/right pager arrows (decorative — no carousel JS).
- **Cards (4 in a row, horizontally scrollable on mobile):**

```tsx
const USE_CASES = [
  {
    title: 'Everyday students',
    description: 'Post, follow, save what matters — and see what your faculty is talking about.',
    apps: ['MessageSquare', 'Repeat2', 'Bookmark', 'Heart'],
  },
  {
    title: 'Community organizers',
    description: 'Run clubs, study groups, and interest circles with built-in roles and request flows.',
    apps: ['Users', 'UserCheck', 'BookOpen', 'MessageSquare'],
  },
  {
    title: 'Event creators',
    description: 'Publish official and unofficial campus events with attachments and document previews.',
    apps: ['Calendar', 'Paperclip', 'MessageSquare', 'Clock'],
  },
  {
    title: 'Faculty discovery',
    description: 'See what your faculty is talking about, find peers, and build connection from day one.',
    apps: ['GraduationCap', 'UserPlus', 'Trophy', 'AtSign'],
  },
];
```

Each card uses `about-usecase-card`. Footer row of each card: small Lucide icons (one per `apps[]`) in white circular wells + a mono `EXPLORE →` link.

### 4.7 `AboutStatMoment` — "Recognition that's real"

Two-column section.

- **Left column:**
  - Eyebrow: `■ FEED THAT GETS YOU`
  - Headline: `Recognition that's real.`
  - Subhead (mentions For You feed + Leaderboard).
  - CTA: `Get started for free →`
- **Right column (stat block):**
  - Small mono eyebrow on top: `STUDENTS ON SUNIBGRAM`
  - Number: `1.000+` (DM Sans, ~5rem, regular). Placeholder for now; replace at publish time.
  - Caption below: `Shipping faster with their voices.` (muted)

### 4.8 `AboutFloatingBubbles` — optional capability poster

(Compose between `AboutStatMoment` and `AboutSplitRoles` if vertical rhythm needs it.)

- Center label panel: `YOUR CAMPUS PROFILE` (mono caps, muted, with a small gradient-cube logo glyph).
- Six rounded-pill bubbles positioned around the center, each with a small Lucide icon + label:
  - `Reply to your faculty thread`
  - `Pin your best post`
  - `Save to a collection`
  - `Trend on the leaderboard`
  - `Join a community`
  - `RSVP to an event`
- All static. No animation. No CTA.

### 4.9 `AboutSplitRoles` — "Multiple roles, one student"

Two-column section.

- **Left column:**
  - Eyebrow: `■ ROLES THAT FOLLOW YOU`
  - Headline: `Multiple roles, one student.`
  - Subhead.
  - CTA: `Get started for free →`
- **Right column:** an avatar chip on the left connected by hairline branches to three role badges on the right:
  - `OWNER` — `@bnpc-club`
  - `MODERATOR` — `@design-binus`
  - `MEMBER` — `@stein-binus`
- Branches are simple SVG curves at `var(--mkt-border)` 1px.

### 4.10 `AboutFinalCTA` — dark closure band

Dark band (background `#000` or `var(--sidebar)`) immediately before the footer. Mirrors Composio's "Your agents decide. We make it happen." block.

- **Headline (left-aligned, large, white):** `Your campus, your voice. We connect the wires.`
- **CTAs (right-aligned, two mono buttons side by side):**
  - Secondary outline (white border on dark): `BOOK A DEMO`
  - Primary filled (white background, black text): `JOIN SUNIBGRAM TODAY`

### 4.11 `AboutFooter` — multi-column categorized

Background `var(--sidebar)` (dark) — matches the final-CTA band, no seam between them.

- **Top row:** small logo + `Sunibgram` wordmark on the far left; nothing on the right of the top row.
- **Five columns** with mono uppercase headers and DM Sans body links:

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

- **Bottom row:** social icons (X, GitHub, LinkedIn, YouTube — placeholder hrefs) on the left, `© Sunibgram 2026` on the right (mono caps, muted).

---

## 5. Design Tokens — quick reference

| Token | Used by |
|---|---|
| `--mkt-bg` | Page bg, nav bg |
| `--mkt-card` | Mockup cards, pills |
| `--mkt-card-tint` | Use-case card bg |
| `--mkt-accent` | Eyebrow text, eyebrow marker, accent links |
| `--mkt-accent-soft` | Eyebrow pill bg (if used) |
| `--mkt-text` | Headlines, primary body |
| `--mkt-text-muted` | Subheads, captions |
| `--mkt-text-tertiary` | Fine print, mono helper text |
| `--mkt-border` | Hairlines, card borders, secondary button border |
| `--mkt-button` / `--mkt-button-text` | Primary CTA bg / text |
| `--sidebar` | Dark closure band + footer |
| `--font-dm-sans` | Headlines, body |
| `--font-mkt-mono` | Nav, eyebrows, CTAs, footer column headers |

**Inline-element headline pattern:**

```tsx
<h1 className="text-5xl md:text-7xl text-center" style={{ color: 'var(--mkt-text)' }}>
  Make your{' '}
  <BinusAccountChip />{' '}
  do more.
</h1>
```

Where `<BinusAccountChip />` is a small inline `<span style={{ display: 'inline-flex', verticalAlign: 'middle' }}>` containing an avatar + `@username` + a mono `BINUS` badge inside a white rounded card with a 1px border.

---

## 6. Implementation Checklist

**Setup**
- [ ] Add marketing palette + utility classes to `app/globals.css` (§3)
- [ ] Confirm `/about` is in `AUTH_PATHS` (already done — verify with one grep)
- [x] Mono font: using the system `ui-monospace` fallback chain — no new font import. Project keeps DM Sans + Orbitron as today.

**New components**
- [ ] `components/about/AboutSplitValue.tsx`
- [ ] `components/about/AboutSplitConnection.tsx`
- [ ] `components/about/AboutUseCases.tsx`
- [ ] `components/about/AboutStatMoment.tsx`
- [ ] `components/about/AboutFloatingBubbles.tsx` (optional)
- [ ] `components/about/AboutSplitRoles.tsx`
- [ ] `components/about/AboutFinalCTA.tsx`

**Rewritten in place**
- [ ] `components/about/AboutNav.tsx` — Composio-style mono nav
- [ ] `components/about/AboutHero.tsx` — light mesh + inline-chip headline
- [ ] `components/about/AboutFooter.tsx` — 5-column categorized

**Composition**
- [ ] Rewrite `app/about/page.tsx` to the section order in §4.1

**Removed**
- [ ] Delete `AboutStats.tsx`, `AboutFeatures.tsx`, `AboutHowItWorks.tsx`, `AboutMission.tsx`, `AboutCTA.tsx`
- [ ] Delete `ScrollReveal.tsx`, `AnimatedCounter.tsx`, `SpotlightCard.tsx`, `HeroMouseGlow.tsx`
- [ ] Remove the old `.about-card` hover rule from `globals.css`

**Validation**
- [ ] Open `http://localhost:3000/about` — page is light, sticky mono nav, hero shows inline-chip headline
- [ ] No dark sidebar or in-app chrome leaks in
- [ ] Mobile responsive — use-case cards become a horizontal scroll, split sections stack
- [ ] Tab through the page — every CTA is keyboard reachable
- [ ] Lighthouse pass — no client JS for non-interactive sections; all components are Server Components
