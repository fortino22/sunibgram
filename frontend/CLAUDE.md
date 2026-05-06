# CLAUDE.md — Sunibgram (University Social Media App)

Read this fully before writing any code, creating any file, or making any architectural decision.

---

## Project Summary

University-exclusive social media platform (X/Threads clone) called **Sunibgram**. Students post threads with text/media, comment, reply, repost, like, bookmark, pin. Also supports communities, follows, premium orders, and OTP-based auth. Portfolio project — code quality and DX matter.

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
| Auth | Custom JWT (Bearer token) — **NOT NextAuth** |
| Real-time | Socket.io client |

Never use Redux, SWR, Axios, Context API for global state, or NextAuth.

---

## API Reference

**Base URL env var:** `NEXT_PUBLIC_API_URL` → e.g. `http://localhost:3000/`  
**Auth header:** `Authorization: Bearer {{BEARER_TOKEN}}`  
**Variables:** `API_SUNIBGRAM`, `SID`, `BEARER_TOKEN`

### Auth — `/auth`
| Method | Endpoint | Body | Notes |
|---|---|---|---|
| POST | `/auth/register` | formdata: `name, username, email, password, confirm_password, avatar(file), banner(file), bio, faculty_id` | |
| POST | `/auth/login/otp/send` | JSON: `{email}` | Send OTP first |
| POST | `/auth/login` | JSON: `{email, password, otp}` | OTP required |
| POST | `/auth/login/prod` | JSON: `{email, password}` | No OTP |
| POST | `/auth/refresh` | JSON: `{sid}` | Refresh token |
| POST | `/auth/logout` | — | Bearer required |
| PUT | `/auth/change-password` | JSON: `{old_password, new_password, confirm_password}` | Bearer |
| POST | `/auth/forgot-password` | JSON: `{email, new_password, confirm_password}` | |

### Verification — `/verification`
| Method | Endpoint | Body |
|---|---|---|
| POST | `/verification/send` | JSON: `{email, type}` — type: `REGISTER` or `FORGOT_PASSWORD` |
| POST | `/verification/verify` | JSON: `{email, code, type}` |

### Profile — `/profile`
| Method | Endpoint | Notes |
|---|---|---|
| GET | `/profile/me` | Bearer |
| GET | `/profile/:userId` | Bearer |
| PUT | `/profile/me` | formdata: `name, username, avatar(file), banner(file), bio, faculty_id, clear_avatar, clear_banner` |

### Users — `/users`
| Method | Endpoint |
|---|---|
| POST | `/users/:userId/follow` |
| DELETE | `/users/:userId/unfollow` |

### Threads — `/threads`
| Method | Endpoint | Body/Query | Notes |
|---|---|---|---|
| GET | `/threads?page=1&limit=10` | — | All threads |
| GET | `/threads/following` | — | Following feed |
| GET | `/threads/me?page=1&limit=10` | — | My posts |
| GET | `/threads/me/replies?page=1&limit=10` | — | My replies |
| GET | `/threads/:id` | — | Single thread |
| GET | `/threads/:id/chain?page=1&limit=10` | — | Thread with all replies |
| POST | `/threads` | formdata: `content, files(file), parent_id(optional for comment), clear_media` | Create or comment |
| PUT | `/threads/:id` | formdata: `content, files(file), clear_media` | |
| DELETE | `/threads/:id` | — | |
| POST | `/threads/:id/like` | — | |
| DELETE | `/threads/:id/unlike` | — | |
| POST | `/threads/:id/bookmark` | — | |
| DELETE | `/threads/:id/unbookmark` | — | |
| POST | `/threads/:id/repost` | — | |
| DELETE | `/threads/:id/unrepost` | — | |
| POST | `/threads/:id/pin` | — | |

### Pins / Likes / Bookmarks
| Method | Endpoint |
|---|---|
| GET | `/pins` |
| GET | `/likes` |
| GET | `/bookmarks` |

### Communities — `/communities`
| Method | Endpoint | Body |
|---|---|---|
| GET | `/communities?category=sport` | optional filter |
| GET | `/communities/me` | |
| GET | `/communities/joined` | |
| GET | `/communities/:id` | |
| POST | `/communities` | formdata: `name, description, rules, category, visibility(PUBLIC/PRIVATE), avatar(file), banner(file)` |
| PUT | `/communities/:id` | formdata: same as create + `clear_avatar, clear_banner` |
| DELETE | `/communities/:id` | |
| POST | `/communities/:id/join` | |
| DELETE | `/communities/:id/leave` | |
| GET | `/communities/:id/members` | |
| DELETE | `/communities/:id/members/:userId` | Kick member |
| PATCH | `/communities/:id/members/:userId/role` | JSON: `{role: "MEMBER"}` |
| POST | `/communities/:id/requests` | Request to join |
| DELETE | `/communities/:id/requests` | Cancel request |
| GET | `/communities/:id/requests` | Get pending requests |
| POST | `/communities/:id/requests/:userId/approve` | |
| POST | `/communities/:id/requests/:userId/decline` | |
| GET | `/communities/:id/threads` | |
| POST | `/communities/:id/threads` | formdata: `content, files(file), clear_media` |

### Orders/Premium — `/orders`
| Method | Endpoint |
|---|---|
| POST | `/orders/premium` |
| POST | `/orders/notification` |
| GET | `/orders/me` |
| GET | `/orders/:id` |
| GET | `/orders` |
| PATCH | `/orders/:id/cancel` |

### Admin — `/admin` (Bearer required)
| Method | Endpoint |
|---|---|
| GET | `/admin/users/?page=1&limit=10` |
| GET | `/admin/users/:id` |
| GET | `/admin/roles/` |
| POST | `/admin/roles/` body: `{name}` |
| GET | `/admin/roles/:id` |
| PUT | `/admin/roles/:id` body: `{name}` |
| DELETE | `/admin/roles/:id` |

### Faculties — `/faculties` (Bearer required)
| Method | Endpoint |
|---|---|
| GET | `/faculties` |
| GET | `/faculties/:id` |
| POST | `/faculties` body: `{faculty_name}` |
| PUT | `/faculties/:id` body: `{faculty_name}` |
| DELETE | `/faculties/:id` |

---

## Code Style & Auth Rules

- No `any` type. No raw `fetch` in components — use `api` from `lib/api.ts`.
- Named exports for components. Default export only in `app/` pages.
- Never `useState` for fetched data — TanStack Query only. Skeletons required.
- Optimistic updates on all mutations (like, follow, bookmark, repost).
- **Auth is JWT (no NextAuth).** Store `BEARER_TOKEN` + `SID` in Zustand `authStore`. `api.ts` auto-attaches `Authorization: Bearer <token>`. Refresh via `POST /auth/refresh` with `{sid}`.

---

## Project Structure

```
src/
├── app/
│   ├── (auth)/login/ signup/
│   ├── feed/
│   ├── threads/[id]/
│   ├── profile/[userId]/
│   ├── communities/[id]/
│   ├── bookmarks/
│   ├── notifications/
│   └── @modal/ (compose, reply)
├── components/
│   ├── thread/ feed/ community/ layout/ user/ ui/
├── hooks/
│   ├── useThreads useThread useCreateThread useLikeThread
│   ├── useBookmark useRepost useFollow useProfile
│   ├── useCommunity useNotifications useSocket
├── stores/
│   ├── authStore.ts   (BEARER_TOKEN, SID, user)
│   └── uiStore.ts
├── lib/
│   ├── api.ts         (fetch wrapper, auto-attach Bearer)
│   ├── queryClient.ts
│   └── socket.ts
└── types/
    ├── thread.ts user.ts community.ts notification.ts order.ts
```

---

## Environment Variables

```env
NEXT_PUBLIC_API_URL=http://localhost:3000/
NEXT_PUBLIC_SOCKET_URL=http://localhost:3000
```

## Feature Phases

**Phase 1:** Project setup, JWT auth (login with OTP flow, register, refresh), Zustand authStore, api wrapper  
**Phase 2:** Thread feed (infinite scroll), create/comment/reply threads, like/bookmark/repost  
**Phase 3:** Profile page, follow/unfollow, pins, my posts/replies tabs  
**Phase 4:** Communities (CRUD, join/leave, requests, member management, community threads)  
**Phase 5:** Notifications, real-time via Socket.io, premium/orders  
**Phase 6:** Admin panel, faculties management, dark mode, mobile responsive
