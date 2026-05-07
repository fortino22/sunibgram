# CLAUDE.md — Sunibgram

Read fully before writing any code or making any architectural decision.

---

## Project Summary

University-exclusive social media platform (X/Threads clone). Students post threads with text/media, comment, reply, repost, like, bookmark, pin. Supports communities, follows, premium orders, OTP-based auth.

---

## Tech Stack (Non-Negotiable)

| Layer | Technology |
|---|---|
| Framework | Next.js 15 (App Router) |
| Language | TypeScript strict |
| Styling | Tailwind CSS v3 |
| UI Components | shadcn/ui |
| Server State | TanStack Query v5 |
| Client State | Zustand v4 |
| Forms | React Hook Form + Zod |
| Auth | Custom JWT — **NOT NextAuth** |
| Real-time | Socket.io client |

Never use: Redux, SWR, Axios, Context API for global state, NextAuth.

---

## Architecture Rules

### Two-Layer Component Model

1. **Page** (`app/**/page.tsx`) — Max 150–200 lines. Owns one route's concern: calls hooks, handles loading/error/skeleton states, composes atoms. No raw `fetch()`.
2. **Atom** (`components/[feature]/[Name].tsx`) — Pure presentational. Props only. No fetching, no hooks (except purely UI hooks like `useMediaQuery`).

**Rule:** If a page grows past 200 lines, extract atoms — not sections. Extract into `components/[feature]/[Name].tsx` only when it's reused elsewhere OR genuinely too complex to inline.

### Hard Rules

- Component files: max ~100–150 lines. Exceed it → split.
- Hook files: max ~80 lines.
- No `any`. No raw `fetch()` in components. No `useState` for server data.
- Named exports for components. Default export only in `app/` pages.
- Every data-fetching page needs a skeleton. Never show blank screen.
- Optimistic updates (`onMutate`/`onError`) on: like, bookmark, repost, follow.
- Check `components/shared/` before creating any new component.

### Shared Components (never inline these)

`<Avatar />` · `<UserHandle />` · `<MediaGrid />` · `<ActionBar />` · `<InfiniteList />` · `<EmptyState />` · `<RelativeTime />` · `<FacultyBadge />` · `<OptimisticToggle />`

### API & State

- `lib/api.ts` — base fetch wrapper only (Bearer attach, token refresh)
- `lib/api/[feature].ts` — typed endpoint functions per feature
- Hooks import from `lib/api/[feature].ts`, never from `lib/api.ts` directly
- Server data → TanStack Query. Auth state → `authStore`. UI state → `uiStore`. Forms → RHF local.

---

## Project Structure
src/
├── app/
│   ├── (auth)/login/   signup/
│   ├── feed/
│   ├── threads/[id]/
│   ├── profile/[userId]/
│   ├── communities/    [id]/
│   ├── bookmarks/      notifications/
│   ├── admin/          users/  faculties/
│   └── @modal/compose/ reply/[id]/
│
├── components/
│   ├── auth/
│   │   ├── AuthGuard.tsx          ← redirect-if-unauthenticated wrapper
│   │   ├── schemas.ts             ← zod schemas for login, register, forgot-password
│   │   ├── LoginForm.tsx
│   │   ├── RegisterForm.tsx
│   │   ├── OtpInput.tsx
│   │   └── ForgotPasswordForm.tsx
│   ├── thread/
│   │   ├── ThreadCard.tsx
│   │   ├── ThreadHeader.tsx
│   │   ├── ThreadBody.tsx
│   │   ├── ThreadActions.tsx
│   │   ├── ThreadComposer.tsx
│   │   ├── ThreadReplyItem.tsx
│   │   └── skeletons/             ← ThreadCardSkeleton, ThreadFeedSkeleton
│   ├── profile/
│   │   ├── ProfileAvatar.tsx
│   │   ├── ProfileBanner.tsx
│   │   ├── ProfileBio.tsx
│   │   ├── ProfileStats.tsx
│   │   ├── ProfileTabs.tsx
│   │   ├── ProfileEditModal.tsx
│   │   ├── schemas.ts             ← zod schema for profile edit
│   │   └── skeletons/             ← ProfileHeaderSkeleton
│   ├── community/
│   │   ├── CommunityCard.tsx
│   │   ├── CommunityHeader.tsx
│   │   ├── CommunityMemberItem.tsx
│   │   ├── CommunityRequestItem.tsx
│   │   ├── CommunityComposer.tsx
│   │   ├── schemas.ts             ← zod schema for create/edit community
│   │   └── skeletons/             ← CommunityCardSkeleton
│   ├── layout/
│   │   ├── AppShell.tsx
│   │   ├── Sidebar.tsx
│   │   ├── BottomNav.tsx
│   │   ├── TopBar.tsx
│   │   └── RightPanel.tsx
│   ├── user/
│   │   ├── UserCard.tsx
│   │   ├── UserHandle.tsx
│   │   └── FollowButton.tsx
│   ├── shared/
│   │   ├── Avatar.tsx
│   │   ├── MediaGrid.tsx
│   │   ├── ActionBar.tsx
│   │   ├── InfiniteList.tsx
│   │   ├── EmptyState.tsx
│   │   ├── RelativeTime.tsx
│   │   ├── FacultyBadge.tsx
│   │   └── OptimisticToggle.tsx
│   └── ui/                        ← shadcn/ui + custom primitives
│
├── hooks/
│   ├── thread/    ← useThreadFeed, useThread, useCreateThread, useThreadLike, useThreadBookmark, useThreadRepost
│   ├── profile/   ← useProfile, useMyProfile, useUpdateProfile
│   ├── community/ ← useCommunities, useCommunity, useCommunityThreads, useJoinCommunity
│   ├── auth/      ← useLogin, useRegister, useLogout, useRefreshToken
│   └── notifications/ ← useNotifications, useSocket
│
├── stores/
│   ├── authStore.ts    ← BEARER_TOKEN, SID, user
│   └── uiStore.ts      ← sidebar, modal, composer draft
│
├── lib/
│   ├── api.ts          ← base fetch wrapper
│   ├── api/            ← threads, auth, profile, communities, orders, admin, faculties
│   ├── queryClient.ts
│   └── socket.ts
│
└── types/              ← thread, user, community, notification, order

**Component folder rules:**
- `schemas.ts` lives inside the feature folder that owns the form, not in `lib/`
- Guards and route-protection wrappers live in `components/auth/`
- Skeletons always in `[feature]/skeletons/`, never mixed with atom files

---

## API Reference

**Base URL:** `NEXT_PUBLIC_API_URL` · **Auth:** `Authorization: Bearer <token>`

| Group | Key Endpoints |
|---|---|
| Auth | POST /auth/login · /auth/login/otp/send · /auth/register · /auth/refresh · /auth/logout · /auth/forgot-password |
| Verification | POST /verification/send · /verification/verify |
| Profile | GET/PUT /profile/me · GET /profile/:userId |
| Users | POST/DELETE /users/:userId/follow\|unfollow |
| Threads | GET /threads · /threads/following · /threads/me · /threads/:id · /threads/:id/chain |
| Thread Actions | POST/DELETE /threads/:id/like\|bookmark\|repost · POST /threads/:id/pin |
| Thread CRUD | POST /threads · PUT /threads/:id · DELETE /threads/:id |
| Feeds | GET /pins · /likes · /bookmarks |
| Communities | CRUD /communities · /communities/:id/join\|leave\|members\|requests\|threads |
| Orders | POST /orders/premium · GET /orders/me · PATCH /orders/:id/cancel |
| Admin | GET/POST/PUT/DELETE /admin/users · /admin/roles |
| Faculties | GET/POST/PUT/DELETE /faculties |

Thread & community POST/PUT use `multipart/form-data`. All others JSON. Bearer required on all except login/register/forgot-password.

---

## Environment Variables

```env
NEXT_PUBLIC_API_URL=http://localhost:3000/
NEXT_PUBLIC_SOCKET_URL=http://localhost:3000
```

---

## Feature Phases

| Phase | Scope |
|---|---|
| 1 | Setup, JWT auth (OTP login, register, refresh), authStore, api wrapper |
| 2 | Thread feed (infinite scroll), create/comment/reply, like/bookmark/repost |
| 3 | Profile, follow/unfollow, pins, my posts/replies tabs |
| 4 | Communities (CRUD, join/leave, requests, members, community threads) |
| 5 | Notifications, Socket.io real-time, premium/orders |
| 6 | Admin panel, faculties, dark mode, mobile responsive |