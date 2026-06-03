---
summary: Official (admin/Fox-reviewed, badged) vs student-run (instant) events, with RSVPs, attendee DMs, and event discussion threads.
last_synced: 2026-06-03
source:
  - frontend/app/about/events/page.tsx
tags: [entity, events]
---

# Events

Browse approved events and RSVP at `/events`; open one at `/events/:id` (see
[[user-routes]]).

## Official events
Submitted by faculties / student councils, require a document attachment (`.doc`/`.docx`,
size cap in [[limits]]), go through **admin approval** — reviewed by [[fox]] first — and
get a verified badge.

## Unofficial (student-run) events
Quick form (title, time, location, audience); **publish instantly**, no admin gate, no
verified badge.

## RSVP & discussion
- Live attendance counter; see who's going; DM attendees from the event page (see
  [[messages]]).
- **Event threads:** official announcements + student commentary; filterable to "official
  only".

## Creating
**Creating an event is free** — any student can create one, no [[premium]] needed.
Student-run events publish instantly; official events still need a `.doc`/`.docx`
attachment and admin (Fox) review before they get the verified badge.
