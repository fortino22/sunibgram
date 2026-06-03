---
summary: Catalog of signed-in user-facing routes — feed, profile, threads, communities, events, messages, bookmarks, explore, settings — with "suggest when" cues.
last_synced: 2026-06-03
source:
  - frontend/app/page.tsx
  - frontend/app/communities/page.tsx
  - frontend/app/settings/account/page.tsx
tags: [reference, routes, user]
---

# User Routes (signed-in)

Clickable, user-facing routes for the **user** role. Requires a verified, signed-in
Binus account. Dynamic segments shown as `:param`. Premium-only actions flagged.

> Admin routes (`/admin/*`) are excluded — see scope note in [[SCHEMA]]. Public/no-login
> routes are in [[public-routes]].

---

## Home / Feed

### `/`
**Home feed.** The main timeline. Tabs:
- **For You** — ranked mix of recent + engaging posts; a "New posts" banner appears
  when fresh threads arrive.
- **Following** — posts only from people you follow.
- **Faculty** — posts from students in your faculty.
- **Community** — pinned/joined communities; pick one to see its feed.

Has the create-post box at the top.
**Suggest when:** "open my feed", "go home", "see latest posts", "switch to following".

---

## Profile

### `/profile`
**Your own profile.** Tabs: Threads, Replies, Media, Likes. Edit profile (display
name, bio, avatar, banner) from here.
**Suggest when:** "view my profile", "edit my profile / avatar / bio".

### `/profile/:userId`
**Another user's profile.** Follow / unfollow, view their threads/replies/media.
**Suggest when:** "see <someone>'s profile", "follow this person".

### `/profile/:userId/followers?tab=followers|following|verified`
**Followers / following list** for a user. `?tab=` switches between `followers`,
`following`, and `verified`.
**Suggest when:** "who follows me", "see who I'm following", "<someone>'s followers".

### `/about-profile/:userId`
**Public profile card.** A read-only, shareable view of a user — display name, bio,
faculty, join date, and username history — separate from the interactive in-app
`/profile/:userId`. Use this when sharing a profile as a link rather than opening the
app view.
**Suggest when:** "share <someone>'s profile", "public link to a profile", "their bio / faculty".

---

## Threads

### `/thread/:id`
**Thread detail.** Full conversation chain — parent, root post, and all replies.
Reply, like, repost, bookmark, and open the **Fox summary** from here.
**Suggest when:** "open this post", "read the replies", "reply to a thread",
"summarize this thread".

> To create a post or reply: use the composer on the home feed (`/`) or the reply box
> inside a thread — there is no separate "new post" page.

---

## Communities

### `/communities`
**Communities home.** Tabs: **Home** (your joined communities feed) and **Explore**
(browse by category). *Creating a community requires premium.*
**Suggest when:** "my communities", "go to communities", "create a community".

### `/communities/discover`
**Discover communities.** Browse and search all communities, filter by category, join.
**Suggest when:** "find a community", "search communities", "join a group".

### `/communities/:id`
**Community detail.** Community info, its threads, and members. Join/leave here.
Public communities join instantly; private ones require owner/admin approval.
Tabs: **Top**, **Latest**, **Media**, **About**, and **Requests** (owner/admin).
**Suggest when:** "open this community", "post in a community", "join this group".

### `/communities/:id/members`
**Community members.** Member list with roles (member / moderator). Owners and admins
can change roles or remove members.
**Suggest when:** "see community members", "make someone a moderator", "remove a member".

---

## Events

### `/events`
**Events feed.** Browse approved events; RSVP. *Creating an event is free — no premium*
(official events need admin approval before badging; unofficial publish instantly).
**Suggest when:** "what events are happening", "RSVP to an event", "create an event".

### `/events/:id`
**Event detail.** Event info, RSVP, and the event's linked discussion threads.
**Suggest when:** "open this event", "see who's going", "event discussion".

---

## Messages

### `/messages?room=:roomId`
**Direct messages & group chats.** Real-time. Left panel lists conversations; right
panel is the open chat. `?room=:roomId` deep-links a specific room. Send text/files,
reply, react, edit, and unsend messages.
**Suggest when:** "message someone", "open my DMs", "group chat", "send a file".

---

## Bookmarks

### `/bookmarks`
**Saved threads.** Tabs: **All** (everything you bookmarked) and **Collections**
(custom folders). Create/edit/delete collections (max 10).
**Suggest when:** "my saved posts", "bookmarks", "make a collection".

### `/bookmarks/collections/:id`
**Collection detail.** All threads inside one bookmark collection.
**Suggest when:** "open my <name> collection".

---

## Notifications

### `/notifications`
**Notifications.** Likes, replies, reposts, follows. Marks read on view; delete
individual items.
**Suggest when:** "check my notifications", "who liked/replied", "clear a notification".

---

## Explore / Search

### `/explore?q=…&tab=foryou|trending`
**Search & discovery.** `?q=` is the search query. Without a query: `?tab=foryou` or
`?tab=trending` (trending hashtags). With a query: tabs for **Top**, **Latest**,
**People**, **Media**, **Communities**.
**Suggest when:** "search for X", "find people", "trending hashtags", "explore".

---

## Leaderboard

### `/leaderboard`
**Weekly leaderboard.** Top contributors ranked by engagement (replies weighted
heaviest; reposts and likes count; bots/self-engagement filtered). Resets weekly.
**Suggest when:** "who's top this week", "see the leaderboard", "my rank/score".

---

## Connect

### `/connect?tab=mutual|faculty`
**Connect / discover people.** Suggested students to follow. `?tab=mutual` (friends of
friends) or `?tab=faculty` (same faculty).
**Suggest when:** "who should I follow", "find people to connect with".

---

## Summaries (Fox)

### `/summaries?tab=summaries|agent`
**Fox summary history.** Archive of AI thread summaries you've generated. Tabs:
`summaries` (your saved summaries) and `agent` (Fox agent activity).
**Suggest when:** "my summaries", "where are my thread summaries".

---

## Feedback

### `/feedback`
**Send feedback.** Submit feedback or suggestions to the team/admins.
**Suggest when:** "give feedback", "report a problem with the app", "make a suggestion".

---

## Settings

Settings use a left-sidebar layout. Index: `/settings`.

| Route | Page | Suggest when |
|---|---|---|
| `/settings/account` | Display name, username, email, avatar, banner | "change my name / username / email / photo" |
| `/settings/privacy` | Private-account toggle (who can see your posts) | "make my account private" |
| `/settings/appearance` | Light / dark theme | "switch to dark mode", "change theme" |
| `/settings/password` | Change password (needs current password) | "change my password" |
| `/settings/blocked-users` | List & unblock blocked accounts | "see who I've blocked", "unblock someone" |
| `/settings/muted-users` | List & unmute muted accounts | "muted accounts", "unmute someone" |
| `/settings/subscription` | Premium — start the 1-month free trial (paid purchase coming soon) | "go premium", "start free trial", "manage premium" |

**Account deletion** lives in account settings (`/settings/account`) — see [[faq]] and
[[privacy-model]].
