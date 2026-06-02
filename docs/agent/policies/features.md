# Feature Capabilities (how things work)

How each feature behaves, so the agent can answer "how do I use X?" and "can I …?".
Sources: `/about/*` feature guides + backend models. For where to click, see
[user-routes.md](../routes/user-routes.md); for hard limits, see [limits.md](./limits.md).

---

## Threads & posts
Source: `/about/threads`, `backend/domain/model/thread/thread.go`.
- **Thread types:** `POST`, `REPLY`, `REPOST`.
- **Reply permissions** — the author chooses who can reply to a thread:
  - `EVERYONE` (default) · `FOLLOWING` (only your followers) · `MUTUAL` (only mutual
    follows).
- **Reposts & quotes:** one-tap repost, or quote with your own comment (quoted thread
  embeds inline). Undo with a tap; original credit stays with the author.
- **Polls:** add a poll to a post (limits in [limits.md](./limits.md)); live results.
- **@mention autocomplete** while composing.
- **Edit / delete / pin** your own threads any time.
- **Feeds:** switch between **For You**, **Following**, and **Faculty** in one tap.
  The faculty feed shows posts only from verified students in your faculty.

## Communities
Source: `/about/communities`, community models.
- **Public vs private:** public communities — any verified Binus student joins
  instantly and sees all threads; private communities — the owner/admin approves each
  join request, and only members can read the threads. Visibility can be changed later.
- **Roles:** members can be promoted to **Moderator**. Owners and moderators manage
  rules, pin, remove members, and change roles.
- **Creating a community requires premium.** Categories: see [about.md](./about.md).

## Events
Source: `/about/events`, event models.
- **Official events:** submitted by faculties/student councils, require a document
  attachment (`.doc`/`.docx`), go through **admin approval**, and get a verified badge.
- **Unofficial (student-run) events:** quick form (title, time, location, audience);
  **publish instantly**, no admin gate, no verified badge.
- **RSVP:** live attendance counter; see who's going; DM attendees from the event page.
- **Event threads:** official announcements + student commentary; filterable to
  "official only".
- **Creating an event requires premium.**

## Messages (DMs & groups)
Source: `/about/messages`.
- **Who you can DM:** anyone with a verified Binus identity — no friend requests, no
  follower gates.
- **Group rooms:** owner can rename, set a description, and remove members
  (limits in [limits.md](./limits.md)).
- **Attachments:** multiple files per message; images preview inline; files are
  searchable in-chat.
- **Message states:** Sending · Sent · Delivered · Seen.
- **Edit / unsend:** any message can be edited; unsending removes it for everyone.
- **Real-time** delivery over Socket.io.

## Leaderboard
Source: `/about/leaderboard`, `/about/help`.
- **Reset:** rankings refresh **weekly**, resetting **Sunday at 00:00** (local time).
- **Scoring:** replies you receive weigh heaviest, then reposts and likes from other
  verified students. Quality interactions outweigh raw post volume. Bots and
  self-engagement are filtered out.
- **Faculty boards:** a separate ranking per faculty, plus your own week-over-week stats.

## Bookmarks
Source: bookmark models, `user-routes.md`.
- Save any thread; organize into **collections** (limits in [limits.md](./limits.md)).

## Fox AI summaries
- Open a thread → generate a **Fox** summary of the post + its replies. Your summary
  history lives at `/summaries` (Fox-agent tab).

## Account controls
Source: `/about/privacy`, settings routes.
- **Private account** (approve new followers), **block** / **unfollow**, **unsend**
  messages, **delete account** — all from settings. See [privacy.md](./privacy.md).
