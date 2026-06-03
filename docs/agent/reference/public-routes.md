---
summary: Catalog of no-login routes — auth pages (/login, /signup, /forgot-password) and the public /about info section, with "suggest when" cues.
last_synced: 2026-06-03
source:
  - frontend/app/login/page.tsx
  - frontend/app/signup/page.tsx
  - frontend/app/about/page.tsx
tags: [reference, routes, public]
---

# Public Routes (no login)

Routes reachable without signing in: authentication pages and the `/about` info
section. Safe to share with anyone, including prospective students.

> Signed-in, user-facing routes are in [[user-routes]].

---

## Auth

### `/login`
**Log in.** Two-step: enter credentials, then the OTP code emailed to you. On success,
goes to the home feed `/`.
**Suggest when:** "sign in", "log in", "I can't access my account".

### `/signup`
**Create an account.** Three steps: account info → profile details → email
verification. Requires a valid `@binus.ac.id` email; one account per student.
**Suggest when:** "sign up", "create an account", "how do I join".

### `/forgot-password?email=`
**Reset password.** Reset via an email OTP. Optional `?email=` pre-fills the address.
**Suggest when:** "forgot my password", "reset password", "can't remember password".

### `/login/expired`
**Sign-in link expired.** Landing shown when a "was this you?" sign-in confirmation
link from the alert email is no longer valid. Offers Back to login and Reset password.
**Suggest when:** "the email link doesn't work", "sign-in link expired", "link expired".

---

## About / Info

The `/about` section explains the product and policies. All pages are public.

| Route | Page | Suggest when |
|---|---|---|
| `/about` | About hub / landing — what Sunibgram is | "what is Sunibgram", "about the app" |
| `/about/company` | What it is + independence (not official Binus) | "what is Sunibgram", "is this official Binus", "who runs it" |
| `/about/our-story` | Why it exists + what the team believes | "why does this exist", "why build it", "what do you stand for" |
| `/about/threads` | Threads guide — posting, replies, polls, reposts | "how do threads work", "how to post / poll" |
| `/about/communities` | Communities guide — public vs private, moderation | "how communities work", "public vs private" |
| `/about/events` | Events guide — official vs student-run, RSVPs | "how events work", "official vs unofficial events" |
| `/about/messages` | Messages guide — DMs, group rooms, read receipts | "how DMs work", "how does chat work" |
| `/about/leaderboard` | Leaderboard guide — weekly scoring, faculty boards | "how is the leaderboard scored" |
| `/about/algorithm` | For You feed transparency — ranking, weights, fairness | "how does the feed work", "how is For You ranked", "why this order" |
| `/about/fox` | Fox AI analyst — event review, daily/weekly reports | "what is Fox", "how does the AI work", "who approves events" |
| `/about/hermes` | Hermes — operational toolset powering automation + Fox | "how do you operate", "what runs the automation" |
| `/about/use-cases` | Real scenarios (first-years, clubs, study groups…) | "how would I use this", "examples" |
| `/about/news` | News — announcements + release notes (index) | "what's new", "release notes", "any announcements" |
| `/about/news/:slug` | A single news article (live: `sunibgram-enters-beta`) | "read the beta announcement", "open that news post" |
| `/about/help` | Help Center + FAQ + support email | "help", "FAQ", "I need support" |
| `/about/privacy` | Privacy Center | "privacy policy", "how is my data used" |
| `/about/terms` | Terms & Conditions | "terms", "rules for using the app" |

For the policy text itself (so the agent can answer without sending the user away), see
the source pages [[terms]], [[privacy]], [[faq]], and [[rules]]. Feature behaviour lives
in the [[about|entities and concepts]].
