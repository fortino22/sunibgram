# API Coverage Report
_Generated 2026-05-21_

Comparison of every registered backend route against every frontend API call.

---

## 1. Entire Feature Areas — Zero Frontend Implementation

### User Settings
| Method | Route | Status |
|---|---|---|
| GET | `/api/user-settings/me` | ❌ Not called |
| PUT | `/api/user-settings/me` | ❌ Not called |

### Reports & Violations
| Method | Route | Status |
|---|---|---|
| GET | `/api/violations` | ❌ Not called |
| GET | `/api/violations/:id` | ❌ Not called |
| POST | `/api/violations` (admin) | ❌ Not called |
| PUT | `/api/violations/:id` (admin) | ❌ Not called |
| DELETE | `/api/violations/:id` (admin) | ❌ Not called |
| POST | `/api/reports` | ❌ Not called |
| GET | `/api/reports` (admin) | ❌ Not called |
| GET | `/api/reports/:id` (admin) | ❌ Not called |
| PUT | `/api/reports/:id/status` (admin) | ❌ Not called |
| DELETE | `/api/reports/:id` (admin) | ❌ Not called |

### Premium / Subscriptions
| Method | Route | Status |
|---|---|---|
| POST | `/api/subscriptions/trial/activate` | ❌ Not called |

---

## 2. Individual Endpoints Not Wired in Frontend

### Threads
| Method | Route | Notes |
|---|---|---|
| GET | `/api/threads/faculty` | Faculty feed — no hook or page |
| GET | `/api/threads/mention` | Mention feed — no hook or page |
| GET | `/api/threads/:id/summary` | AI summary — endpoint ready, not surfaced |
| POST | `/api/threads/:id/poll/vote` | Poll voting — backend ready, no UI |
| DELETE | `/api/threads/:id/unpin` | Frontend only calls `POST /pin`; unpin is never sent |

### Events
Only `GET /api/events` and `GET /api/events/:id` are called. All management endpoints are missing:

| Method | Route | Status |
|---|---|---|
| GET | `/api/events/me` | ❌ Not called |
| GET | `/api/events/admin` | ❌ Not called |
| PUT | `/api/events/:id` | ❌ Not called |
| DELETE | `/api/events/:id` | ❌ Not called |
| PATCH | `/api/events/:id/approve` | ❌ Not called |
| PATCH | `/api/events/:id/decline` | ❌ Not called |
| PATCH | `/api/events/:id/revision` | ❌ Not called |
| GET | `/api/events/:id/threads` | ❌ Not called |

### Follows
| Method | Route | Notes |
|---|---|---|
| GET | `/api/follows/me/followers` | Frontend uses `/:userId/followers` with own ID instead — functional workaround |
| GET | `/api/follows/me/following` | Same workaround |

### Communities
| Method | Route | Notes |
|---|---|---|
| GET | `/api/communities/:id/rules` | Rules are bundled in the community detail response — likely redundant by design |

### Chat
| Method | Route | Notes |
|---|---|---|
| GET | `/api/chats/rooms/:id` | Frontend derives active room by filtering the rooms-list cache instead |

### Misc
| Method | Route | Notes |
|---|---|---|
| GET | `/api/faculties/me` | Faculty data arrives embedded in the profile response — not called separately |
| POST | `/api/email/send` | Likely internal/server-only |
| POST | `/api/admin/cache/clear` | Intentionally absent from frontend |

---

## 3. Frontend Calls Routes That Don't Exist in the Backend

### Events service (`frontend/lib/services/events.ts`)
The events service was written against a different API spec than what the backend router actually registers:

| Method | Route frontend calls | Backend reality |
|---|---|---|
| POST | `/api/events/:id/bookmark` | ❌ Not registered |
| DELETE | `/api/events/:id/bookmark` | ❌ Not registered |
| POST | `/api/events/:id/share` | ❌ Not registered |
| GET | `/api/events/:id/official-posts` | ❌ Not registered |
| GET | `/api/events/:id/discussion` | ❌ Not registered |

### Orders service (`frontend/lib/services/orders.ts` or similar)
| Method | Route frontend calls | Backend reality |
|---|---|---|
| POST | `/api/orders/premium` | ❌ Not registered in router |
| GET | `/api/orders/me` | ❌ Not registered in router |
| PATCH | `/api/orders/:id/cancel` | ❌ Not registered in router |

### Threads service
| Method | Route frontend calls | Backend reality |
|---|---|---|
| GET | `/api/pins` | ❌ Not registered |

---

## 4. Summary by Priority

| Priority | Area | Gap |
|---|---|---|
| 🔴 High | **Events** | 5 frontend calls hit non-existent routes; 7 backend management routes have no frontend |
| 🔴 High | **Orders / Premium** | 3 frontend calls (`/orders/premium`, `/orders/me`, `/orders/:id/cancel`) hit non-existent backend routes |
| 🟡 Medium | **User Settings** | 2 endpoints, zero frontend — private account toggle is blocked |
| 🟡 Medium | **Reports & Violations** | 10 endpoints, zero frontend — reporting flow is completely absent |
| 🟡 Medium | **Poll voting** | Backend fully implemented, no UI to cast a vote |
| 🟡 Medium | **Thread faculty / mention feeds** | Backend ready, no page or hook wired |
| 🟡 Medium | **Thread AI summary** | Endpoint ready, never surfaced in UI |
| 🟢 Low | **Thread unpin** | Likely broken — DELETE never sent; frontend uses pin as a toggle |
| 🟢 Low | **Subscriptions trial** | `POST /api/subscriptions/trial/activate` exists, not callable from UI |
| 🟢 Low | **Chat room GET by ID** | Functional workaround via list cache |
| 🟢 Low | **Community rules GET** | Redundant — rules come bundled in community response |
