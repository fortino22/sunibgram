---
summary: Real-time DMs and group rooms — no friend requests, multi-file attachments, edit/unsend, and Sending→Seen states over Socket.io.
last_synced: 2026-06-03
source:
  - frontend/app/about/messages/page.tsx
tags: [entity, messaging]
---

# Messages (DMs & groups)

Direct and group chat at `/messages?room=:roomId` (see [[user-routes]]).

## Who you can DM
Anyone with a verified Binus identity — no friend requests, no follower gates.

## Group rooms
The owner can rename, set a description, and remove members. Name / description caps are
in [[limits]].

## Attachments & delivery
- Multiple files per message; images preview inline; files are searchable in-chat.
  Attachment size cap in [[limits]].
- **Message states:** Sending · Sent · Delivered · Seen.
- **Edit / unsend:** any message can be edited; unsending removes it for everyone.
- **Real-time** delivery over Socket.io.

DMs and group rooms are visible only to the people you sent them to — see
[[privacy-model]].
