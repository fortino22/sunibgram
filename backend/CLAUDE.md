# CLAUDE.md — Backend

Guidance for Claude Code when working in this directory.

## Architecture

Strict Clean Architecture. **Dependency rule: `CMD → WebAPI → Application → Domain ← Infrastructure`**. Domain has zero external imports.

| Layer | Package | Responsibility |
|---|---|---|
| Domain | `domain/model/` | Structs, GORM tags, business logic, validation |
| Domain | `domain/repository/` | Repository interfaces (`IXxxRepository`) |
| Domain | `domain/service/` | Service interfaces (`IXxxService`) |
| Application | `application/usecase/` | Implements `IXxxService`; orchestration only |
| Application | `application/helper/` | Shared logic: hashtag extraction, media upload, pagination, thread visibility, notifications |
| Infrastructure | `infrastructure/repository/` | Implements `IXxxRepository`; GORM + Redis |
| Infrastructure | `infrastructure/storage/` | MinIO (`IStorageService`) |
| Infrastructure | `infrastructure/socket/` | WebSocket hub implementation |
| Infrastructure | `infrastructure/scheduler/` | Weekly background tasks (event/hashtag counter resets) |
| Infrastructure | `infrastructure/ollama/` | Ollama AI client (thread summaries) |
| Infrastructure | `infrastructure/email/` | SMTP client |
| Infrastructure | `infrastructure/ratelimit/` | Rate limiter middleware |
| WebAPI | `webapi/controller/` | Fiber handlers → parse request → call service → DTO |
| WebAPI | `webapi/router/` | Route registration |
| WebAPI | `webapi/socket/` | Socket.io event handlers (chat, notifications) |
| WebAPI | `webapi/utils/` | `Claims` struct, feed builder, timeago |
| DTO | `dto/` | Request/response structs; WebAPI only |
| CMD | `cmd/main.go` | DI wiring |

**Go module:** `github.com/sunibgram`

---

## Adding a New Feature

One file per concern. Multi-entity features get separate files at every layer.

```
domain/model/xxx.go
domain/repository/i_xxx_repository.go
domain/service/i_xxx_service.go
application/usecase/xxx_service.go
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

**Cache invalidation tiers** — pick the right one for the access pattern:

| Tier | When | API |
|---|---|---|
| Direct delete | Key is deterministic (one cached value per `entity:id`) | `cache.Delete(ctx, key1, key2...)` |
| Reverse-index tag | Many parametric keys share an entity (pagination, query variants) | `cache.RegisterCacheKey(ctx, tag, key)` on Set + `cache.InvalidateTag(ctx, tag)` on mutation |
| Pattern SCAN (batched) | Global low-cardinality families with no registry, or legacy paths | `cache.DeleteByPatterns(ctx, patterns...)` |

Prefer tags over SCAN. Tags are O(set size), SCAN is O(keyspace). Add tag builders to `infrastructure/repository/tag_keys.go` — never inline `"tag:..."` strings. Tag set TTL = 24h (`cache.TagTTL`); any cached value with longer TTL must register under a different mechanism.

**Tag invariant** — every Set of a key that should be tag-invalidatable must register under the tag. Half-migration = permanent stale (until cache TTL). Migrate one entity surface at a time: change all Set sites first, then switch invalidate path. Existing tagged surfaces:
- `userFeedsTag(uid)` — `thread:userfeed/usermedia/following:<uid>:*` (Set sites in `thread_repository_user_feed.go`)
- `userForYouTag(uid)` — `feed:foryou:user:<uid>:src/newcount:*` (Set sites in `thread_repository_for_you.go`)
- `threadFeedsTag(tid)` — feed caches containing thread (Set sites in `thread_repository_*.go` via `storeFeedIDs`/`trackFeedCacheKey`)
- `interactionThreadCacheKeysKey(table, tid)` — like/bookmark/repost listings containing thread (Set in `interaction_query.go`)

**Profile-meta overlay** — `FeedBuilder.AuthorField` in `webapi/utils/feed_builder.go` overlays `is_premium` + `faculty` from `user:meta:<uid>` (`ProfileMetaCacheKey`) onto cached threads. Lets `thread:id:*`, `thread:chain:*`, `thread:parent:*`, and feed-list caches stay cached across `UpdatePremium` / `UserFaculty` mutations — only `user:meta:<uid>` needs invalidation.

Overlay correctness requires **both** at the controller call site:
1. `FeedBuilder.ProfileSvc` set (non-nil) — `controller_builder_helpers.go:newFeedBuilder` wires this; bespoke `FeedBuilder{...}` literals must include it.
2. `fb.LoadAvatarsForUserIDs(threadUserIDs(threads))` called before `ToFeedItem`/`AuthorField`.

Missing either → falls back to embedded `User.IsPremium` / `User.UserFaculty` from cached thread = stale.

**Mutation → invalidation cheatsheet** — surfaces that snapshot mutable user/entity fields:

| Mutation | DTOs that snapshot it | Required invalidation |
|---|---|---|
| `User.IsPremium` (`UpdatePremium`, `ExpireForDev`, trial activation) | `FeedAuthor`, `ReportUserResponse`, `NotificationSenderResponse`, `CommunityAbout/Member/Request/UserResponse`, `ExploreUserResponse`, `FollowUserResponse`, `UserProfileResponse` | `user:id:<uid>`, `user:meta:<uid>`, `notification:sender:<uid>`, `reports:all`, `reports:id:*`, `follows:user:*`, `community:about/users/requests/id:v2:*`, `community:threads:top/latest:*`, `explore:*`, all thread feed families (via `invalidateUserThreadCaches`) |
| `UserFaculty.Save/Delete` (assignment) | `FeedAuthor.Faculty`, `ReportUserResponse.Faculty`, `NotificationSenderResponse.Faculty`, `UserProfileResponse.Faculty` | `userfaculty:user:<uid>`, `user:meta:<uid>`, `user:id:<uid>`, `profile:user:<uid>`, `notification:sender:<uid>`, `reports:all/id:*`, all thread feed families |
| `Faculty.Update/Delete` (entity rename) | every overlay reads `user:meta:<uid>.FacultyName` | `faculty:id:<id>`, `faculty:list:*`, `profile:user:*`, `userfaculty:user:*`, **`user:meta:*`** (pattern — affects all users in faculty), `notification:sender:*`, `explore:*`, all thread feed families incl. `thread:id:*`, `thread:chain:*`, `feed:trending:global`, `feed:foryou:*` |
| `Event.Update/Delete` | `FeedItemResponse.EventID/EventName` (cached threads preload `ThreadEvents.Event`) | `invalidateEvent(id)` (lists + `event:threads:<id>:*`) **plus** `invalidateEventThreadCaches(id)` per affected thread: `thread:id/parent`, `threadFeedsTag`, `interactionThreadCacheKeysKey`, user/community/chain patterns, `feed:trending:global`, `feed:foryou:*`, `thread:roots/all-communities/my-communities:*`, `community:threads:top/latest:*` |

Note: `feed:trending:global` (scheduler-written) and `feed:foryou:user:<uid>:src:*` (per-viewer foryou source) cache full `[]model.Thread` with full preload — they snapshot event_name/is_premium/faculty but do **not** register under `threadFeedsTag`. Mutations affecting fields embedded in cached threads must pattern-delete `feed:foryou:*` and direct-delete `feed:trending:global` explicitly. Overlay handles user fields, but event fields have no overlay → cache invalidation is the only fix.

**File validation** — lives in `domain/model/`. `model.ValidateImageExt` (jpg/jpeg/png/webp) in `media.go`. `model.ValidateThreadMediaExt` in `thread_media.go`. Never in application layer.

**File upload** — endpoints accepting files use `ctx.FormValue` + `ctx.FormFile` (multipart). No-file endpoints use `ctx.BodyParser` (JSON).

**Storage** — `StorageUsecase` resolves bucket via `bucketWhitelist` in `storage_service.go` → env var → MinIO. Policy (`public`/`private`) via `MINIO_POLICY_<BUCKET_UPPER>`. Buckets auto-created on first upload.

**Auth** — JWT middleware at `webapi/middleware/auth.go`. Claims: `utils.Get(ctx)` from `webapi/utils/ctrl_helpers.go` → `Claims{UserID, Username, Email, RoleID, RoleName, SID}`.

**Pagination** — services return `([]T, int64, error)`. Controllers wrap in `dto.PaginatedXxxResponse{Data, Meta: dto.FeedMeta{Total, Page, Limit}}`.

**Error handling:**
```go
"forbidden"     → fiber.StatusForbidden
"xxx not found" → fiber.StatusNotFound
else            → fiber.StatusBadRequest
```

**Mutation responses** — `fiber.Map{"message": "xxx successfully"}`. GET endpoints return full DTOs.

**Counters** — `gorm.Expr("GREATEST(count - 1, 0)")` for decrement. Patch Redis after DB update.

**WebSocket** — hub in `infrastructure/socket/hub.go` implements `domain/service/ISocketHub`. Handlers in `webapi/socket/`. Push notifications + chat events via Socket.io.

**AI summaries** — `infrastructure/ollama/ollama_client.go` + `application/usecase/thread_summary_service.go`. Ollama generates thread summaries; seeded via `infrastructure/seeder/thread_summary_seeder.go`.

**Schedulers** — `infrastructure/scheduler/` runs weekly cron jobs. Do not put business logic there; call service/repo methods directly.

**Application helpers** — reusable logic in `application/helper/`. Use `pagination.go` for limit/offset validation, `hashtag_extractor.go` for parsing, `thread_visibility.go` for privacy rules. Never put these in domain or webapi.

---

## Storage Buckets

| Upload Type | Env Var | Bucket | Policy |
|---|---|---|---|
| `avatar` | `MINIO_BUCKET_AVATAR` | `sunibgram-avatar` | public |
| `banner` | `MINIO_BUCKET_BANNER` | `sunibgram-banner` | public |
| `post` | `MINIO_BUCKET_POST` | `sunibgram-post` | public |
| `poll` | `MINIO_BUCKET_POLL` | `sunibgram-poll` | public |
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
