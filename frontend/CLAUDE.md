# CLAUDE.md — Sunibgram Frontend

Read fully before writing any code or making any architectural decision.

---

## Tech Stack (Non-Negotiable)

| Layer | Technology |
|---|---|
| Framework | Next.js 15 (App Router) |
| Language | TypeScript strict |
| Styling | Tailwind CSS v4 + CSS custom properties (`var(--*)`) for colors |
| Server State | TanStack Query v5 |
| Client State | Zustand v4 |
| Forms | React Hook Form + Zod |
| Auth | Custom JWT via httpOnly `sid` cookie — **NOT NextAuth** |
| Real-time | Socket.io client |

Never use: Redux, SWR, Axios, Context API for global state, NextAuth.

**Styling rules:** Layout/spacing → Tailwind. Colors → CSS vars inline style. Never hardcode colors.

---

## Architecture Rules

### Two-Layer Component Model
1. **Page** (`app/**/page.tsx`) — Owns one route. Calls hooks, handles loading/error/skeleton, composes atoms. Max ~200 lines.
2. **Atom** (`components/[feature]/[Name].tsx`) — Pure presentational, props only. No fetching, no hooks.

**Hard rules:** No `any`. No raw `fetch()` in components. No `useState` for server data. Named exports everywhere. Every data page needs a skeleton.

### Optimistic Updates (Required)
Apply `onMutate`/`onError` pattern on: like, bookmark, repost, follow/unfollow.

---

## Auth & Token Pattern

- `authStore` holds decoded JWT payload: `user_id`, `username`, `email`, `role_id`, `role_name` — **no avatar_url**
- Token lives in memory only; refreshed via httpOnly `sid` cookie through `/api/auth/refresh`
- **Never read avatar from authStore** — always use `useMyProfile()` instead

```ts
// CORRECT — avatar from profile cache, updates when profile is edited
const { data: profile } = useMyProfile();
const avatarSrc = profile?.avatar ?? null;

// WRONG — authStore.user has no avatar field
const user = useAuthStore(s => s.user);
const avatarSrc = user?.avatar_url; // always undefined
```

---

## Services & Data Fetching

Services live in `lib/services/[feature].ts`. Hooks import from services, never call `fetch()` directly.

### Backend Response Mapping
Backend field names differ from frontend model — always map at the service layer:

```ts
// Backend returns: user_id, avatar_url, banner_url, follower_count, is_following
// Frontend UserProfile uses: id, avatar, banner, followers_count, is_following
function mapProfile(raw: BackendProfileResponse): UserProfile { ... }
```

### Key Backend Gotchas
- **No `/api/threads/me`** — use `GET /api/threads/user/:userId` with the user's numeric ID
- **No `/api/profile/me` wrapping** — backend returns flat object, no `{ data: ... }` wrapper
- **Thread user endpoints:** `GET /api/threads/user/:userId`, `GET /api/threads/user/:userId/replies`, `GET /api/threads/user/:userId/media`

### Query Key Conventions
```ts
['profile', 'me']          // own profile (useMyProfile)
['profile', userId]        // other user's profile
['threads', userId]        // user's threads
['threads', userId, 'replies'] // user's replies
['followers', userId]
['following', userId]
```

---

## Shared Components (Never Inline)

| Component | Location |
|---|---|
| `<Avatar />` | `components/shared/Avatar.tsx` |
| `<ActionBar />` | `components/shared/ActionBar.tsx` |
| `<EmptyState />` | `components/shared/EmptyState.tsx` |
| `<ImageCropModal />` | `components/profile/ImageCropModal.tsx` |
| `<FollowListModal />` | `components/profile/FollowListModal.tsx` |

---

## Actual File Structure

```
app/
  page.tsx                    ← home feed
  profile/
    page.tsx                  ← own profile (useMyProfile, useMyThreads)
    [userId]/page.tsx         ← other user profile (useProfile, useUserThreads)
  threads/[id]/page.tsx
  communities/  bookmarks/  notifications/  explore/

components/
  layout/   Sidebar.tsx  (uses useMyProfile for avatar)
  profile/  ProfileInfo.tsx  ProfileThreadCard.tsx  ProfileEditModal.tsx
            ImageCropModal.tsx  FollowListModal.tsx
  shared/   Avatar.tsx  ActionBar.tsx  EmptyState.tsx  CreatePostBox.tsx
  thread/   ThreadCard.tsx  ThreadComposer.tsx
  ui/       Icon.tsx  avatar.tsx

hooks/
  useMyProfile.ts       useProfile.ts        useUpdateProfile.ts
  useMyThreads.ts       useUserThreads.ts    useMyReplies.ts  useUserReplies.ts
  useFollowUser.ts      useFollowers.ts      useFollowing.ts
  thread/  useCreateThread.ts  useThreadFeed.ts  useThreadActions.ts

lib/
  services/  profile.ts  threads.ts  follows.ts  auth.ts
  cropImage.ts           ← canvas-based crop utility for react-easy-crop
  api.ts                 ← base fetch wrapper (Bearer attach, token refresh)
  queryClient.ts

stores/  authStore.ts  uiStore.ts
```

---

## API Reference

**Base URL:** `NEXT_PUBLIC_API_URL` (must be set at build time for browser)  
**Auth:** `Authorization: Bearer <token>`  
All routes prefixed with the group name (no `/api/` prefix).

### Auth
| Method | Endpoint | Notes |
|---|---|---|
| POST | /auth/register | multipart/form-data |
| POST | /auth/login | JSON: `{email, password}` — issues tokens, fires "was this you?" alert email |
| GET | /auth/login/verify | public; `?token=&action=yes\|no` from alert email — `no` revokes that session |
| POST | /auth/refresh | JSON: `{sid}` — uses httpOnly sid cookie |
| POST | /auth/logout | |
| PUT | /auth/change-password | JSON: `{old_password, new_password, confirm_password}` |
| POST | /auth/forgot-password | JSON: `{email, new_password, confirm_password}` |

### Verification (OTP)
| Method | Endpoint | Notes |
|---|---|---|
| POST | /verification/send | JSON: `{email, type}` — type: `REGISTER` or `FORGOT_PASSWORD` |
| POST | /verification/verify | JSON: `{email, code, type}` |

### Profile
| Method | Endpoint | Notes |
|---|---|---|
| GET | /profile/me | flat object, no wrapper |
| GET | /profile/:userId | |
| PUT | /profile/me | multipart — supports `clear_avatar=true`, `clear_banner=true` |

### Users
| Method | Endpoint | Notes |
|---|---|---|
| POST | /users/:userId/follow | |
| DELETE | /users/:userId/unfollow | |
| GET | /users/suggested | suggested people to follow |

### Follows (separate from /users)
| Method | Endpoint | Notes |
|---|---|---|
| GET | /follows/me/followers | own followers |
| GET | /follows/me/following | own following |
| GET | /follows/:userId/followers | other user's followers |
| GET | /follows/:userId/following | other user's following |

### Threads
| Method | Endpoint | Notes |
|---|---|---|
| GET | /threads?page&limit | all threads (home feed) |
| GET | /threads/following?page&limit | following feed |
| GET | /threads/faculty?faculty_id&page&limit | faculty feed |
| GET | /threads/user/:userId?page&limit | user's posts |
| GET | /threads/user/:userId/replies?page&limit | user's replies |
| GET | /threads/user/:userId/media?page&limit | user's media threads |
| GET | /threads/:id | single thread |
| GET | /threads/:id/chain?page&limit | thread + all replies (full chain) |
| GET | /threads/:id/summary | LLM-powered thread summary |
| GET | /threads/mention?q= | @mention autocomplete search |
| POST | /threads | multipart — create thread or reply (`parent_id` field) |
| POST | /threads | multipart — poll: add `poll_question`, `poll_options[]`, `poll_duration_seconds`, `poll_option_images[]` |
| PUT | /threads/:id | multipart — `keep_media_ids[]` to preserve existing media |
| DELETE | /threads/:id | |
| POST | /threads/:id/like | |
| DELETE | /threads/:id/unlike | distinct path, not DELETE /like |
| POST | /threads/:id/bookmark | |
| DELETE | /threads/:id/unbookmark | distinct path, not DELETE /bookmark |
| POST | /threads/:id/repost | |
| DELETE | /threads/:id/unrepost | distinct path, not DELETE /repost |
| POST | /threads/:id/pin | both pin and unpin are POST |
| POST | /threads/:id/unpin | |

### Likes & Bookmarks
| Method | Endpoint | Notes |
|---|---|---|
| GET | /likes | own liked threads |
| GET | /bookmarks | own bookmarked threads |

### Communities
| Method | Endpoint | Notes |
|---|---|---|
| POST | /communities | multipart |
| GET | /communities/me | communities I own |
| GET | /communities/joined | communities I joined |
| GET | /communities/discover?category= | discover by category UUID |
| GET | /communities/search?community_name= | search communities |
| GET | /communities/threads?category= | all community threads |
| GET | /communities/threads/top?query= | top community threads |
| GET | /communities/threads/latest?query= | latest community threads |
| GET | /communities/me/threads | threads from my communities |
| GET | /communities/:id | |
| PUT | /communities/:id | multipart — `new_rules[]`, `keep_rule_ids[]`, `clear_banner` |
| DELETE | /communities/:id | |
| POST | /communities/:id/join | public community |
| DELETE | /communities/:id/leave | |
| GET | /communities/:id/members | |
| DELETE | /communities/:id/members/:userId | kick member |
| PATCH | /communities/:id/members/:userId/role | JSON: `{role}` — `MEMBER` or `MODERATOR` |
| POST | /communities/:id/requests | request to join private community |
| DELETE | /communities/:id/requests | cancel own join request |
| GET | /communities/:id/requests | list pending requests (admin/owner) |
| POST | /communities/:id/requests/:userId/approve | |
| POST | /communities/:id/requests/:userId/decline | |
| POST | /communities/:id/threads | create thread in community (multipart) |
| GET | /communities/:id/threads/top | |
| GET | /communities/:id/threads/latest | |
| GET | /communities/:id/threads/media | |

### Explore / Search
| Method | Endpoint | Notes |
|---|---|---|
| GET | /explore/search?q= | unified search results |
| GET | /explore/top?q= | top threads matching query |
| GET | /explore/latest?q= | latest threads matching query |
| GET | /explore/people?q= | people search |
| GET | /explore/media?q= | media threads matching query |
| GET | /explore/communities?q= | communities matching query |

### Notifications
| Method | Endpoint | Notes |
|---|---|---|
| GET | /notifications | own notifications |
| POST | /notifications | send notification (system use) |
| DELETE | /notifications/:id | |

### Hashtags
| Method | Endpoint | Notes |
|---|---|---|
| GET | /hashtags/trending | trending hashtags |

### Events
| Method | Endpoint | Notes |
|---|---|---|
| POST | /events/ | multipart — `type: official\|unofficial`, `attachment` required for official |
| GET | /events/ | all approved events |
| GET | /events/me?status= | own events — status: `approved\|declined\|revision` |
| GET | /events/:id | |
| PUT | /events/:id | multipart |
| DELETE | /events/:id | |
| GET | /events/admin?status&name | admin: list event requests |
| PATCH | /events/:id/approve | admin |
| PATCH | /events/:id/decline | admin — JSON: `{reason}` |
| PATCH | /events/:id/revision | admin — JSON: `{note}` |
| GET | /events/:id/threads?filter= | threads linked to event — filter: `official` |

### Orders / Premium
| Method | Endpoint | Notes |
|---|---|---|
| POST | /orders/premium | buy premium subscription |
| POST | /orders/notification | Midtrans payment webhook |
| GET | /orders/me | own orders |
| GET | /orders/:id | |
| PATCH | /orders/:id/cancel | |
| GET | /orders | admin — all transactions |

### Subscriptions
| Method | Endpoint | Notes |
|---|---|---|
| POST | /subscriptions/trial/activate | activate free trial |

### Chats
| Method | Endpoint | Notes |
|---|---|---|
| POST | /chats/rooms/private | JSON: `{target_user_id}` |
| POST | /chats/rooms/group | JSON: `{name, description, member_ids[]}` |
| GET | /chats/rooms | list all chat rooms |
| GET | /chats/rooms/:id | |
| PATCH | /chats/rooms/:id | edit group info: `{name, description}` |
| DELETE | /chats/rooms/:id | delete group |
| POST | /chats/rooms/:id/members | add members: `{user_ids[]}` |
| GET | /chats/rooms/:id/messages | get messages in room |
| POST | /chats/rooms/:id/messages | send message (multipart — `message` + `files[]`) |
| PATCH | /chats/messages/:id | edit message |
| DELETE | /chats/messages/:id | unsend message |
| POST | /chats/messages/read | mark read: `{room_id, chat_ids[]}` |

### Faculties
| Method | Endpoint | Notes |
|---|---|---|
| GET | /faculties | all faculties |
| GET | /faculties/me | current user's faculty |
| GET | /faculties/:id | |
| POST | /faculties | admin — JSON: `{faculty_name}` |
| PUT | /faculties/:id | admin |
| DELETE | /faculties/:id | admin |

### Categories (Community Categories)
| Method | Endpoint | Notes |
|---|---|---|
| GET | /categories | all categories |
| POST | /categories | admin — JSON: `{name}` |
| PUT | /categories/:id | admin |
| DELETE | /categories/:id | admin |

### User Settings
| Method | Endpoint | Notes |
|---|---|---|
| GET | /user-settings/me | |
| PUT | /user-settings/me | JSON: `{is_private}` |

### Reports
| Method | Endpoint | Notes |
|---|---|---|
| POST | /reports | user report: `{violation_id, explanation, reported_user_id}` or post report: `{violation_id, explanation, reported_post_id}` |
| GET | /reports | admin — all reports |
| GET | /reports/:id | |
| DELETE | /reports/:id | |

### Violations
| Method | Endpoint | Notes |
|---|---|---|
| GET | /violations | all violation types |
| GET | /violations/:id | |
| POST | /violations | admin — JSON: `{name, description}` |
| PUT | /violations/:id | admin |
| DELETE | /violations/:id | admin |

### Admin
| Method | Endpoint | Notes |
|---|---|---|
| GET | /admin/users/?page&limit | |
| GET | /admin/users/:id | |
| GET | /admin/roles/ | |
| POST | /admin/roles/ | JSON: `{name}` |
| GET | /admin/roles/:id | |
| PUT | /admin/roles/:id | |
| DELETE | /admin/roles/:id | |
| POST | /admin/cache/clear | flush Redis cache |

---

## Environment Variables

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_SOCKET_URL=http://localhost:8080
```

Must be provided as build-time ARGs in Docker — browser-side env vars are baked at build, not runtime.

---

## Feature Status

| Phase | Scope | Status |
|---|---|---|
| 1 | JWT auth (login, register, OTP send/verify, refresh, logout, forgot password, change password) | Done |
| 2 | Thread feed (home + following + faculty), create/reply/poll, like/bookmark/repost, pin/unpin | Done |
| 3 | Profile (own + other users), follow/unfollow, followers/following modal, image crop | Done |
| 4 | Thread context menu (pin, delete, edit), thread chain view, thread LLM summary | Pending |
| 5 | Communities (CRUD, join/leave/request, members, roles, threads by tab) | Done |
| 6 | Notifications, Socket.io real-time | Pending |
| 7 | Explore/Search (top, latest, people, media, communities) | Pending |
| 8 | Chats (private + group rooms, messages, read receipts) | Pending |
| 9 | Events (create, admin approve/decline/revision, event threads) | Pending |
| 10 | Premium subscription (orders, Midtrans, free trial) | Pending |
| 11 | Reports & Violations | Pending |
| 12 | Trending hashtags, @mention autocomplete | Pending |
| 13 | User settings (private account toggle) | Pending |
