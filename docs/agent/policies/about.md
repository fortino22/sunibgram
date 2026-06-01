# About Sunibgram

Orientation for the agent — what Sunibgram is and what it stands for. Sources:
`/about`, `/about/what`, `/about/team`, and backend category seeder.

---

## What it is
Sunibgram is **the social network for Binus** — a verified-only platform for Binus
University students. Threads, communities, events, and DMs, all from people who
actually share a campus with you. Verified, on-campus, built by students. No bots, no
anonymous noise.

## Why it exists
Campus life deserved better than a group chat. The everyday questions students ask
each other — where to study, who's joining the event, did the lab notes get posted —
kept getting buried in group chats, public timelines, and platforms that don't know
what a faculty is. Sunibgram puts all of that in one verified place: the people are
real, the feed is on-topic, and the rules are written for campus, not advertisers.

## The six principles ("what we stand for")
1. **Verified, always** — every account is tied to a real `@binus.ac.id` email. No
   bots, no anonymous trolls, no off-campus strangers.
2. **Built around campus** — faculty, batch, and community are first-class. The feed
   knows the difference between a classmate and a stranger.
3. **Student-first** — made by Binus students, for Binus students.
4. **No engagement-bait** — no recommended outrage, no dark patterns, no infinite
   vanity metrics. The feed is for catching up, not doom-scrolling.
5. **Easy to find your people** — communities, faculty feeds, and search make it
   simple to land in the right room, even in your first week.
6. **Honest about what we are** — a student-built project, not an official Binus
   product. Independence keeps us honest.

## Independence (important)
Sunibgram is a **student project, built independently in Jakarta — not officially
affiliated with, sponsored by, or partnered with Binus University**, and does not
speak for the university. Anyone with a verified `@binus.ac.id` email is welcome; the
rest of the internet is not.

## Community categories
Communities are organized under these categories (source:
`backend/infrastructure/seeder/category_seeder.go`): **Academic, Arts, Career, Gaming,
Social, Sports, Student Life, Technology, Volunteering.**

## Faculties
Each account is tied to a Binus faculty (chosen at signup, used for the faculty feed).
Seeded faculties (source: `backend/infrastructure/seeder/faculty_seeder.go`):
BINUS ASO School of Engineering · BINUS Business School · BINUS Graduate Program ·
BINUS Online · Doctor of Computer Science · Faculty of Digital Communication and Hotel
and Tourism · Faculty of Engineering · Faculty of Humanities · School of Accounting ·
School of Computer Science · School of Computing and Creative Arts · School of Design ·
School of Information Systems.
> This is the seeded default list; the live list comes from the `/faculties` endpoint
> and may change. Don't promise a faculty exists — let the user pick from the app.

## Core features (where to learn more)
Threads & replies, polls, reposts · Communities (public/private) · Events
(official/student-run) · Direct & group messages · Leaderboard · Bookmarks · Fox AI
thread summaries. Feature guides live under `/about/*` — see
[public-routes.md](../routes/public-routes.md). Where to *do* each thing →
[user-routes.md](../routes/user-routes.md).

## Support
**support@sunibgram.com** — read every message, usually within a couple of working
days.
