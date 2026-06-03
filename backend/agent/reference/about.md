---
summary: Orientation hub — what Sunibgram is, why it exists, and links out to the principles, independence, faculties, features, and AI layer.
last_synced: 2026-06-03
source:
  - frontend/app/about/page.tsx
  - frontend/app/about/company/page.tsx
  - frontend/app/about/our-story/page.tsx
  - backend/infrastructure/seeder/category_seeder.go
tags: [reference, overview, hub]
---

# About Sunibgram

Orientation hub for the agent — what Sunibgram is, with links to the canonical pages for
each topic.

## What it is
Sunibgram is **the social network for Binus** — a verified-only platform for Binus
University students. [[threads|Threads]], [[communities]], [[events]], and
[[messages|DMs]], all from people who actually share a campus with you. Verified,
on-campus, built by students. No bots, no anonymous noise.

## Why it exists
Campus life deserved better than a group chat. The everyday questions students ask each
other — where to study, who's joining the event, did the lab notes get posted — kept
getting buried in group chats, public timelines, and platforms that don't know what a
faculty is. Sunibgram puts all of that in one verified place: the people are real, the
feed is on-topic, and the rules are written for campus, not advertisers.

## What it stands for
The six **[[principles]]** (verified-always, campus-built, student-first, no
engagement-bait, easy to find your people, honest about what it is). Crucially,
Sunibgram is a student project — see [[independence]].

## Community categories
[[communities|Communities]] are organized under these categories (source:
`backend/infrastructure/seeder/category_seeder.go`): **Academic, Arts, Career, Gaming,
Social, Student Life, Technology, Volunteering.**

## Faculties
Each account is tied to a Binus faculty, used for the faculty feed. The seeded list and
the "don't promise a faculty exists" caveat live in [[faculties]].

## Core features
[[threads|Threads & replies]], polls, reposts · [[communities]] (public/private) ·
[[events]] (official/student-run) · [[messages|Direct & group messages]] ·
[[leaderboard]] · [[bookmarks]] · [[fox|Fox AI]] thread summaries. Feature guides live
under `/about/*` — see [[public-routes]]. Where to *do* each thing → [[user-routes]].

## AI / automation layer
[[fox|Fox]] is the on-platform AI analyst (thread summaries + official-event review +
daily/weekly reports); [[hermes]] is the operational toolset it runs on.

## Support
**support@sunibgram.com** — read every message, usually within a couple of working days.
