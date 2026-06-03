---
summary: Plain-language privacy notice — what's collected, how it's used (no ad selling), who sees your content, your controls, security, and retention.
last_synced: 2026-06-03
source:
  - frontend/app/about/privacy/page.tsx
tags: [source, privacy]
---

# Privacy

Plain-language summary for the agent. Source: `/about/privacy`
(`frontend/app/about/privacy/page.tsx`). **Last updated · January 2026.** Sunibgram is
a student-built campus network, operated independently and **not officially affiliated
with Binus University** (see [[independence]]).

> For the full notice, point users to `/about/privacy` (see [[public-routes]]). The
> who-sees-what model is summarized in [[privacy-model]].

---

## Who this applies to
Everyone who creates an account, browses content, or communicates through Sunibgram.

## Information we collect
Only what's needed to run the service responsibly:
- **Account information** — your verified `@binus.ac.id` email, username, display
  name, faculty, and password (stored as a one-way hash).
- **Content you post** — threads, replies, polls, community posts, direct messages,
  uploaded files/media, and reactions.
- **Usage signals** — basic logs of which features you use, when, and from which
  device (for stability and abuse investigation).
- **Technical metadata** — IP address, browser, and device type, kept only as long as
  needed for security and rate-limiting.

**Not collected:** contact lists, browsing history outside Sunibgram, or location
beyond what's implied by your network address.

## How we use your information
Only to: provide and improve features you sign up for; verify each account belongs to
a current Binus student; show the right content to the right people (faculty feeds,
your communities, your conversations); protect the platform from spam, abuse, and
unauthorized access; and respond to support requests and reports.

**No data selling, no renting to advertisers, no advertising profiles, and no
third-party ad trackers.**

## Who sees your content
Visibility follows where you posted:
- **Public threads** — visible to any verified Sunibgram user.
- **Community threads** — visible only to verified members of that community.
- **Direct messages & group rooms** — visible only to the people you sent them to.
- **Faculty feeds** — posts by people in your faculty; only verified students see them.

Switching to a **private account** means only people you approve see your full profile
and posts.

## Your controls
Edit your profile/avatar/banner; block or unfollow any user; unsend messages (removes
them for everyone); switch to a private account; delete your account (removes your
threads, replies, and messages from the active service). For anything settings don't
cover — including **data export requests** — write to **support@sunibgram.com**.

## How we protect your information
Passwords are hashed before storage (never plain text). Sessions use short-lived
tokens refreshed via a secure, httpOnly cookie. Uploaded files live on managed object
storage with restricted access. OS and dependencies are patched on a regular cadence.
Security reports are prioritized — email **support@sunibgram.com** with subject
*Security report* and allow a reasonable window before public disclosure.

## How long we keep data
Content stays until you delete it or your account. Deleting your account removes active
content from the service. Some operational logs may be retained briefly for safety,
fraud prevention, and legal obligations.

## Changes
The notice is updated when practices change; the "Last updated" date reflects the
latest revision. Substantial changes surface in-app at next sign-in.

## Contact
Questions about how Sunibgram handles your information? **support@sunibgram.com** —
usually answered within a couple of working days.
