---
name: session-handoff
description: >
  Generate a structured session handoff document when the user types /session-handoff
  or asks to "hand off", "end session", "wrap up the session", "create a handoff", or
  "summarize for next session". This skill captures everything the next Claude session
  needs to continue work seamlessly — goals, decisions, progress, open threads, and
  next steps. Always trigger this skill when the user wants to preserve context across
  sessions, switch to a new chat, or hand off work to another instance of Claude.
  Also trigger if the user says things like "save my progress", "I need to start a new
  chat but continue this", or "write a handoff prompt".
---

# Session Handoff Skill

## Purpose

When a user invokes `/session-handoff` (or equivalent phrasing), generate a comprehensive
handoff document they can paste into a new Claude session to resume work with zero context
loss. The output should be a **ready-to-paste prompt** — not a summary for the human to
read, but a briefing *for Claude* to read.

---

## Step 1 — Scan the Conversation

Before writing anything, mentally scan the full conversation for:

- **The original goal / problem statement** — what were we trying to do?
- **Decisions made** — choices, trade-offs, things ruled out and why
- **Work completed** — files created, code written, tasks done, outputs produced
- **Current state** — where exactly things stand right now
- **Open questions / blockers** — anything unresolved or debated
- **Next actions** — concrete next steps that were planned or implied
- **Important constraints** — preferences, rules, technical stack, style choices the user stated

If any of these are ambiguous or missing, make a reasonable inference and note it.

---

## Step 2 — Write the Handoff Document

Output a fenced code block (so the user can easily copy it) containing the handoff prompt.
Use this structure:

```
## Session Handoff

### Context
[1-3 sentences: what this project/task is about and why]

### Goal
[What we are ultimately trying to achieve]

### What's Been Done
- [Completed item 1]
- [Completed item 2]
- ...

### Decisions Made
- [Decision]: [Brief rationale or what was ruled out]
- ...

### Current State
[Clear description of exactly where things stand right now — what exists, what's half-done, what the last action was]

### Open Questions / Blockers
- [Anything unresolved, debated, or waiting on input]

### Next Steps
1. [First concrete action to take]
2. [Second action]
3. ...

### Key Constraints & Preferences
- [Tech stack, style, rules, preferences the user stated]
- ...

### Files / Artifacts
- [filename or description]: [what it is / its current state]
- ...

---
*Pick up from "Current State" above. Ask me if anything needs clarification.*
```

Omit any section that is genuinely empty (e.g., no files if nothing was created).
Keep each section tight — this is a brief for Claude, not a report for humans.

---

## Step 3 — Deliver

1. Present the fenced handoff block so the user can copy it with one click.
2. After the block, add a short human-readable note (2-3 sentences max) explaining:
   - What to do with it ("Paste this at the start of your next chat")
   - Any section you felt uncertain about and why
3. Optionally offer to refine any section if the user wants to add detail.

---

## Tone & Style

- Write the handoff **in second person addressed to Claude** ("You were helping the user...", "The user wants...", "Next you should...")
- Be specific and concrete — vague summaries lose context
- Preserve exact names: file names, variable names, decisions, technologies
- Don't editorialize — report what happened, not your opinion of it
- Keep the whole document under ~400 words unless the session was very complex

---

## Example Output Format

````
## Session Handoff

### Context
The user is building a personal finance dashboard in Next.js that reads from a CSV export from their bank.

### Goal
Create a working dashboard with charts showing monthly spend by category, deployable to Vercel.

### What's Been Done
- Scaffolded a Next.js 14 app with Tailwind CSS
- Built a CSV parser (`/lib/parseCSV.ts`) that maps columns to a normalized Transaction type
- Created a `/dashboard` page with a placeholder bar chart using Recharts

### Decisions Made
- **Recharts over Chart.js**: preferred because it's React-native and smaller bundle
- **No database**: just reading CSV from `/public/data/` for now — user doesn't want backend complexity
- **Categories hardcoded**: user said they'll refine the list later

### Current State
The bar chart renders but shows dummy data. The next step is wiring `parseCSV.ts` into the page so the chart displays real transactions grouped by category and month.

### Open Questions / Blockers
- Should the CSV be uploaded by the user at runtime, or is it a static file checked into the repo? (assumed static for now)

### Next Steps
1. Import and call `parseCSV` in the dashboard page's `getStaticProps`
2. Aggregate transactions by `category` and `month`
3. Pass the aggregated data into the Recharts `<BarChart>`

### Key Constraints & Preferences
- TypeScript throughout
- No backend / API routes
- Keep it simple — user is not a professional developer

### Files / Artifacts
- `/lib/parseCSV.ts`: complete, parses CSV rows into `Transaction[]`
- `/app/dashboard/page.tsx`: scaffolded with dummy chart data, needs real data wired in

---
*Pick up from "Current State" above. Ask me if anything needs clarification.*
````

---

## Edge Cases

- **Short sessions**: If the conversation is only a few turns, still produce the handoff — even a short brief is useful. Trim empty sections.
- **Multiple unrelated threads**: Capture all of them under separate sub-headings within the relevant sections.
- **Sensitive info**: Don't include passwords, API keys, or PII in the handoff. Note their existence ("API key was provided") without echoing the value.
- **Ambiguous next steps**: If next steps weren't explicitly discussed, infer the most logical continuation from context and label it "(inferred)".