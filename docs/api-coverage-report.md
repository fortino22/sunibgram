# API Coverage Report
_Updated 2026-05-23_

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
| GET | `/api/threads/mention` | Mention feed — no hook or page |
| GET | `/api/threads/:id/summary` | AI summary — endpoint ready, not surfaced |
| DELETE | `/api/threads/:id/unpin` | Frontend only calls `POST /pin`; unpin is never sent |

### Events
Only `GET /api/events`, `GET /api/events/:id`, and `POST /api/events` are called. All management and participation endpoints are missing:

| Method | Route | Status |
|---|---|---|
| GET | `/api/events/me` | ❌ Not called |
| GET | `/api/events/admin` | ❌ Not called |
| GET | `/api/events/pins/me` | ❌ Not called |
| GET | `/api/events/joined/me` | ❌ Not called |
| PUT | `/api/events/:id` | ❌ Not called |
| DELETE | `/api/events/:id` | ❌ Not called |
| PATCH | `/api/events/:id/approve` | ❌ Not called |
| PATCH | `/api/events/:id/decline` | ❌ Not called |
| PATCH | `/api/events/:id/revision` | ❌ Not called |
| GET | `/api/events/:id/threads` | ❌ Not called |
| POST | `/api/events/:id/threads/:threadId/pin` | ❌ Not called |
| DELETE | `/api/events/:id/threads/:threadId/unpin` | ❌ Not called |
| POST | `/api/events/:id/pin` | ❌ Not called |
| DELETE | `/api/events/:id/unpin` | ❌ Not called |
| POST | `/api/events/:id/join` | ❌ Not called |
| DELETE | `/api/events/:id/leave` | ❌ Not called |

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
| POST | `/api/dev/subscriptions/expire` | Dev-only route |

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

### Orders service (`frontend/lib/services/orders.ts`)
Backend registers `POST /api/orders/notification`, `GET /api/orders`, and `GET /api/orders/:id` — the following frontend calls hit non-existent routes:

| Method | Route frontend calls | Backend reality |
|---|---|---|
| POST | `/api/orders/premium` | ❌ Not registered |
| GET | `/api/orders/me` | ❌ Not registered (backend has `/api/orders`) |
| PATCH | `/api/orders/:id/cancel` | ❌ Not registered |

### Threads service
| Method | Route frontend calls | Backend reality |
|---|---|---|
| GET | `/api/pins` | ❌ Not registered (backend only has POST/DELETE `/threads/:id/pin`) |

---

## 4. Resolved Since Last Report

These gaps from the previous report are now closed:

| Area | What changed |
|---|---|
| **Poll voting** | `POST /api/threads/:id/poll/vote` — frontend now calls this ✅ |
| **Thread faculty feed** | `GET /api/threads/faculty` — frontend now wires `faculty_id` param and calls this ✅ |

---

## 5. Summary by Priority

| Priority | Area | Gap |
|---|---|---|
| 🔴 High | **Events** | 5 frontend calls hit non-existent routes; 16 backend management/participation routes have no frontend |
| 🔴 High | **Orders / Premium** | 3 frontend calls (`/orders/premium`, `/orders/me`, `/orders/:id/cancel`) hit non-existent backend routes |
| 🟡 Medium | **User Settings** | 2 endpoints, zero frontend — private account toggle is blocked |
| 🟡 Medium | **Reports & Violations** | 10 endpoints, zero frontend — reporting flow is completely absent |
| 🟡 Medium | **Thread mention feed** | `GET /api/threads/mention` — backend ready, no page or hook wired |
| 🟡 Medium | **Thread AI summary** | `GET /api/threads/:id/summary` — endpoint ready, never surfaced in UI |
| 🟢 Low | **Thread unpin** | `DELETE /api/threads/:id/unpin` — likely broken; frontend uses pin as a toggle only |
| 🟢 Low | **Subscriptions trial** | `POST /api/subscriptions/trial/activate` exists, not callable from UI |
| 🟢 Low | **Chat room GET by ID** | Functional workaround via list cache |
| 🟢 Low | **Community rules GET** | Redundant — rules come bundled in community response |
