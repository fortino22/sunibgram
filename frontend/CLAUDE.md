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
All routes prefixed with `/api/`.

| Group | Key Endpoints |
|---|---|
| Auth | POST /auth/login · /auth/register · /auth/refresh · /auth/logout |
| Profile | GET /profile/me · GET /profile/:userId · PUT /profile/me (multipart) |
| Users | POST /users/:userId/follow · DELETE /users/:userId/unfollow |
| Followers | GET /users/:userId/followers · GET /users/:userId/following |
| Threads | GET /threads · /threads/user/:userId · /threads/user/:userId/replies · /threads/user/:userId/media · /threads/:id |
| Thread Actions | POST/DELETE /threads/:id/like · /threads/:id/bookmark · /threads/:id/repost |
| Thread CRUD | POST /threads (multipart) · PUT /threads/:id · DELETE /threads/:id |
| Communities | CRUD /communities · /communities/:id/join\|leave\|members\|threads |

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
| 1 | JWT auth (login, register, refresh, OTP) | Done |
| 2 | Thread feed, create/reply, like/bookmark/repost | Done |
| 3 | Profile (own + other users), follow/unfollow, followers/following modal, image crop | Done |
| 4 | Thread context menu (pin, delete, edit) | Pending |
| 5 | Communities (CRUD, join/leave, members) | Done |
| 6 | Notifications, Socket.io real-time | Pending |
| 7 | Search (backend has no search endpoints yet) | Pending |
