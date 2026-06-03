---
summary: Public or private rooms organized by category; public join instantly, private need approval. Roles, moderation, and premium-gated creation.
last_synced: 2026-06-03
source:
  - frontend/app/about/communities/page.tsx
tags: [entity, communities]
---

# Communities

Rooms organized by category. Find them at `/communities` and `/communities/discover`;
open one at `/communities/:id` (see [[user-routes]]).

## Public vs private
- **Public** — any verified Binus student joins instantly and sees all threads.
- **Private** — the owner/admin approves each join request, and only members can read the
  threads.
Visibility can be changed later. Who-sees-what detail is in [[privacy-model]].

## Roles & moderation
Members can be promoted to **Moderator**. Owners and moderators manage rules, pin, remove
members, and change roles. Community owners and their admins enforce rules within their
own rooms — Sunibgram steps in only if a community breaks platform rules. See
[[moderation-and-reports]].

## Creating
**Creating a community requires [[premium]].** Categories: **Academic, Arts, Career,
Gaming, Social, Student Life, Technology, Volunteering** (see [[about]]).

Name / description / rules caps are in [[limits]].
