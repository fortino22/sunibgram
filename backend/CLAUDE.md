# CLAUDE.md — Backend

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture

Strict Clean Architecture. **Dependency rule: `CMD → WebAPI → Application → Domain ← Infrastructure`**. Domain has zero external imports.

| Layer | Package | Responsibility |
|---|---|---|
| Domain | `domain/model/` | Structs, GORM tags, business logic, validation |
| Domain | `domain/repository/` | Repository interfaces (`IXxxRepository`) |
| Domain | `domain/service/` | Service interfaces (`IXxxService`) |
| Application | `application/usecase/` | Implements `IXxxService`; orchestration only |
| Infrastructure | `infrastructure/repository/` | Implements `IXxxRepository`; GORM + Redis |
| Infrastructure | `infrastructure/storage/` | MinIO (`IStorageService`) |
| WebAPI | `webapi/controller/` | Fiber handlers → parse request → call service → DTO |
| WebAPI | `webapi/router/` | Route registration |
| DTO | `dto/` | Request/response structs; WebAPI only |
| CMD | `cmd/main.go` | DI wiring |

**Go module:** `github.com/sunibgram`

---

## Adding a New Feature

One file per concern. Multi-entity features (e.g. community + member + request) get separate files at every layer.

```
domain/model/xxx.go                      ← struct + validation methods
domain/repository/i_xxx_repository.go    ← interface
domain/service/i_xxx_service.go          ← interface
application/usecase/xxx_service.go       ← implements IXxxService
infrastructure/repository/xxx_repository.go
dto/xxx_dto.go
webapi/controller/xxx_controller.go
webapi/router/xxx_routes.go
```

Wire in:
1. `infrastructure/migration/migrate.go` → `db.AutoMigrate(&model.Xxx{})`
2. `cmd/main.go` → repo → service → controller → `router.Register`
3. `webapi/router/router.go` → add param + call `registerXxxRoutes`

---

## Key Patterns

**R/W DB split** — repos take `writeDB, readDB *gorm.DB`. Mutations → `writeDB`, queries → `readDB`.

**Redis cache** — `cache.Get → readDB → cache.Set`. Invalidate on mutations. Keys: `entity:id`, `entity:list:page:limit`.

**File validation** — lives in `domain/model/`. `model.ValidateImageExt` (jpg/jpeg/png/webp) in `media.go`. `model.ValidateThreadMediaExt` in `thread_media.go`. Never in application layer.

**File upload** — endpoints accepting files use `ctx.FormValue` + `ctx.FormFile` (multipart). No-file endpoints use `ctx.BodyParser` (JSON).

**Storage** — `StorageUsecase` resolves bucket via `bucketWhitelist` in `storage_service.go` → env var → MinIO. Policy (`public`/`private`) via `MINIO_POLICY_<BUCKET_UPPER>`. Buckets auto-created on first upload.

**Auth** — JWT middleware at `webapi/middleware/auth.go`. Claims: `ctxutil.Get(ctx)` → `Claims{UserID, Username, Email, RoleID, RoleName, SID}`.

**Pagination** — services return `([]T, int64, error)`. Controllers wrap in `dto.PaginatedXxxResponse{Data, Meta: dto.FeedMeta{Total, Page, Limit}}`.

**Error handling:**
```go
"forbidden"     → fiber.StatusForbidden
"xxx not found" → fiber.StatusNotFound
else            → fiber.StatusBadRequest
```

**Mutation responses** — `fiber.Map{"message": "xxx successfully"}`. GET endpoints return full DTOs.

**Counters** — `gorm.Expr("GREATEST(count - 1, 0)")` for decrement. Patch Redis after DB update.

---

## Storage Buckets

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

## Architecture Rules

1. Business logic → Domain model methods, never Application
2. WebAPI imports Domain interfaces only, never Infrastructure
3. Domain imports standard lib only — zero external packages
4. Application knows Domain interfaces only, never concrete Infrastructure structs
5. DTOs in WebAPI only — never in Domain or Application
