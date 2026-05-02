# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Sunibgram — university-exclusive social media platform (X/Threads clone). Students at Binus University only (`@binus.ac.id` email). Features: threads, replies, reposts, likes, bookmarks, follows, communities.

**Monorepo layout:**
- `backend/` — Go/Fiber REST API (port 8080)
- `frontend/` — Next.js 16 app (port 3000)
- `docker-compose.yml` — Postgres 16, Redis 7, MinIO, app container

---

## Dev Environment

**Start infrastructure (Postgres, Redis, MinIO):**
```bash
docker compose up postgres redis minio -d
```

**Run backend locally:**
```bash
cd backend
go run ./cmd/main.go
# uses backend/.env
```

**Run full stack via Docker:**
```bash
docker compose up --build
# uses backend/.env.docker (DB_HOST=postgres, REDIS_ADDR=redis:6379, MINIO_ENDPOINT=minio:9000)
```

**Run frontend:**
```bash
cd frontend
npm run dev     # dev server
npm run build   # production build
npm run lint    # eslint
```

**Build backend binary:**
```bash
cd backend
go build -o server ./cmd/main.go
```

---

## Backend Architecture

Strict Clean Architecture. **Dependency rule: `CMD → WebAPI → Application → Domain ← Infrastructure`**. Domain has zero external imports.

| Layer | Package | Responsibility |
|---|---|---|
| Domain | `domain/model/` | Structs, GORM tags, business logic methods, validation |
| Domain | `domain/repository/` | Repository interfaces (`IXxxRepository`) |
| Domain | `domain/service/` | Service interfaces (`IXxxService`) |
| Application | `application/usecase/` | Implements `IXxxService`; orchestration only, no business logic |
| Infrastructure | `infrastructure/repository/` | Implements `IXxxRepository`; GORM, Redis, R/W DB split |
| Infrastructure | `infrastructure/storage/` | MinIO client (`IStorageService`) |
| Infrastructure | `infrastructure/cache/` | Redis wrapper |
| WebAPI | `webapi/controller/` | Fiber handlers; parse request → call service → return DTO |
| WebAPI | `webapi/router/` | Route registration functions |
| DTO | `dto/` | Request/response structs; WebAPI only, never Domain/Application |
| CMD | `cmd/main.go` | DI wiring: infra → application → webapi → start Fiber |

**Go module:** `github.com/sunibgram`

### Adding a New Feature

Always create one file per concern:

```
domain/model/xxx.go                    ← struct + business validation methods
domain/model/xxx_member.go             ← related sub-entity (if needed)
domain/repository/i_xxx_repository.go  ← interface
domain/service/i_xxx_service.go        ← interface
application/usecase/xxx_service.go     ← implements IXxxService
infrastructure/repository/xxx_repository.go ← implements IXxxRepository
dto/xxx_dto.go                         ← response structs
webapi/controller/xxx_controller.go    ← Fiber handlers
webapi/router/xxx_routes.go            ← route registration
```

Then wire in:
1. `infrastructure/migration/migrate.go` — add `db.AutoMigrate(&model.Xxx{})`
2. `cmd/main.go` — instantiate repo → service → controller → pass to `router.Register`
3. `webapi/router/router.go` — add param + call `registerXxxRoutes`

### Key Infrastructure Patterns

**Read/Write DB split** — all repos take `writeDB, readDB *gorm.DB`. Mutations use `writeDB`, queries use `readDB`.

**Redis cache** — cache-first pattern in repos: `cache.Get → readDB → cache.Set`. Invalidate on mutations. TTL via `cache.DefaultTTL`. Keys follow pattern: `entity:id`, `entity:list:page:limit`.

**File storage** — MinIO via `StorageUsecase`. Bucket type resolves via `bucketWhitelist` in `storage_service.go` → env var → MinIO bucket name. Policy (`public`/`private`) set via `MINIO_POLICY_<BUCKET_NAME_UPPER>` env var. Buckets auto-created on first upload.

**File validation** — extension validation belongs in `domain/model/`. See `model.ValidateImageExt` (images) and `model.ValidateThreadMediaExt` (thread media). Never put extension maps in application layer.

**Auth** — JWT via `webapi/middleware/auth.go`. Claims extracted with `ctxutil.Get(ctx)` → returns `Claims{UserID, Username, Email, RoleID, RoleName, SID}`.

**Pagination** — list service methods return `([]T, int64, error)`. Controllers wrap in `dto.PaginatedXxxResponse{Data, Meta: dto.FeedMeta{Total, Page, Limit}}`.

**Error handling in controllers:**
```go
if err.Error() == "forbidden" → fiber.StatusForbidden
if err.Error() == "xxx not found" → fiber.StatusNotFound
else → fiber.StatusBadRequest
```

**Mutation responses** — return `fiber.Map{"message": "xxx successfully"}`, not full entity. GET endpoints return full DTOs.

**Counters** — use `gorm.Expr("GREATEST(count - 1, 0)")` for decrement to prevent negatives. Patch Redis cache after DB counter update.

### Storage Buckets

| Upload Type | Env Var | Bucket | Policy |
|---|---|---|---|
| `avatar` | `MINIO_BUCKET_AVATAR` | `sunibgram-avatar` | public |
| `banner` | `MINIO_BUCKET_BANNER` | `sunibgram-banner` | public |
| `post` | `MINIO_BUCKET_POST` | `sunibgram-post` | public |
| `document` | `MINIO_BUCKET_DOCUMENT` | `sunibgram-document` | private |
| `assignment` | `MINIO_BUCKET_ASSIGNMENT` | `sunibgram-assignment` | private |
| `community_avatar` | `MINIO_BUCKET_COMMUNITY_AVATAR` | `sunibgram-community-avatar` | public |
| `community_banner` | `MINIO_BUCKET_COMMUNITY_BANNER` | `sunibgram-community-banner` | public |

---

## Frontend Architecture

Next.js 16 (App Router), TypeScript strict, Tailwind CSS v4. See `frontend/CLAUDE.md` for full frontend conventions (component patterns, state management, file structure, type definitions).

**Key constraints from frontend/CLAUDE.md:**
- No `any` type; explicit return types always
- Named exports for all components (default exports only for `app/` pages)
- TanStack Query for server state; Zustand for UI/auth state
- All HTTP via `src/lib/api.ts` — never raw fetch in components
- Optimistic updates on all mutations

---

## Architecture Rules (from Rules.md)

1. Never put business logic in Application — belongs in Domain model methods
2. Never import Infrastructure in WebAPI — use Domain interfaces only
3. Never import external libraries in Domain
4. Application only knows Domain interfaces, never concrete Infrastructure structs
5. DTO used in WebAPI only — never in Domain or Application
