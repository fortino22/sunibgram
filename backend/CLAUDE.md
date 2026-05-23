# CLAUDE.md — Backend

Guidance for Claude Code when working in this directory.

## Architecture

Strict Clean Architecture, feature-modular. Every layer (`domain/`, `application/`, `infrastructure/`, `webapi/`) is split into per-module subdirectories so each feature is locatable as a folder. **Dependency rule: `CMD → WebAPI → Application → Domain ← Infrastructure`**. Domain has zero external imports.

### Layer / module map

| Layer | Path | Package convention | Responsibility |
|---|---|---|---|
| Domain (model) | `domain/model/<m>/` | `<m>model` (e.g. `threadmodel`, `usermodel`) | Structs, GORM tags, business logic, validation |
| Domain (repo iface) | `domain/repository/<m>/` | `<m>repoiface` | `IXxxRepository` interfaces |
| Domain (svc iface) | `domain/service/<m>/` | `<m>svciface` | `IXxxService` interfaces |
| Application (usecase) | `application/usecase/<m>/` | `<m>uc` | Implements `IXxxService`; orchestration only |
| Application (helper) | `application/helper/` | `helper` | Cross-cutting: pagination, hashtag extraction, media upload, thread visibility, chat crypto, notification helper |
| Infrastructure (repo impl) | `infrastructure/repository/<m>/` | `<m>repo` | Implements `IXxxRepository`; GORM + Redis |
| Infrastructure (shared) | `infrastructure/repository/_shared/` | `infrarepoShared` | `tag_keys.go`, `viewer_flag_sets.go`, `InteractionRepo`, `event_counter_helpers.go`, `explore_cache_invalidation.go` |
| Infrastructure (others) | `infrastructure/{cache,storage,socket,scheduler,ollama,email,ratelimit,config,migration,seeder}/` | one package per subdir | Cross-cutting infra |
| WebAPI (controller) | `webapi/controller/<m>/` | `<m>ctrl` | Fiber handlers → parse request → call service → DTO |
| WebAPI (controller shared) | `webapi/controller/_shared/` | `ctrlshared` | `RequireAuth`, `ParseIDParam`, `HandleServiceError`, `FeedBuilder` wiring helpers |
| WebAPI (router) | `webapi/router/<m>/` | `<m>route` | Per-module `Register<Xxx>Routes(...)` |
| WebAPI (router entry) | `webapi/router/router.go` | `router` | Top-level `Register(...)` that calls every per-module register |
| WebAPI (dto) | `webapi/dto/<m>/` | `<m>dto` | Request/response structs; WebAPI only |
| WebAPI (dto shared) | `webapi/dto/_shared/` | `utils` (dir/pkg name intentionally mismatched) | `FormattedTime`, `Wrap` — JSON time helpers |
| WebAPI (utils/middleware/socket/helper) | `webapi/{utils,middleware,socket,helper}/` | per dir | Claims, FeedBuilder, JWT MW, Socket.io handlers, timeago |
| CMD | `cmd/main.go` | `main` | DI wiring |

**Go module:** `github.com/sunibgram`

### Cycle merges (matching tightly-coupled features in one package)

Some domain entities share bidirectional GORM relations and must therefore live in the same package. These pairs are intentionally merged:

| Merged into | Why |
|---|---|
| `user_faculty` → `user/` (both `usermodel`) | `User.UserFaculty` ↔ `UserFaculty.User` mutual struct field |
| `event` → `thread/` (both `threadmodel`) | `Thread.ThreadEvents` ↔ `ThreadEvent.Thread` mutual struct field |
| `community_member` usecase → `community/` (`communityuc`) | Mutual helper calls between community and community-member services |
| `community_member` + `community_thread` controllers → `community/` (`communityctrl`) | Share `CommunityController` helpers |

---

## Adding a New Feature

One file per concern, one module-dir per feature, **at every layer**.

```
domain/model/xxx/xxx.go                          // package xxxmodel
domain/repository/xxx/i_repository.go            // package xxxrepoiface
domain/service/xxx/i_service.go                  // package xxxsvciface
application/usecase/xxx/xxx_service.go           // package xxxuc
infrastructure/repository/xxx/xxx_repository.go  // package xxxrepo
webapi/dto/xxx/xxx_dto.go                        // package xxxdto
webapi/controller/xxx/xxx_controller.go          // package xxxctrl
webapi/router/xxx/xxx_routes.go                  // package xxxroute
```

Wire in:
1. `infrastructure/migration/migrate.go` → `db.AutoMigrate(&xxxmodel.Xxx{})`
2. `cmd/main.go` → repo (`xxxrepo.New...`) → service (`xxxuc.New...`) → controller (`xxxctrl.New...`)
3. `webapi/router/router.go` → import `xxxroute "github.com/sunibgram/webapi/router/xxx"` + call `xxxroute.RegisterXxxRoutes(api, xxxCtrl, ...)`

If the new entity's GORM relations create an import cycle with another entity (mutual struct fields), merge their `<m>model` files into a single directory rather than fight the cycle.

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

Prefer tags over SCAN. Tags are O(set size), SCAN is O(keyspace). All tag builders live in `infrastructure/repository/_shared/tag_keys.go` (package `infrarepoShared`) — never inline `"tag:..."` strings. Tag set TTL = 24h (`cache.TagTTL`); any cached value with longer TTL must register under a different mechanism.

**Tag invariant** — every Set of a key that should be tag-invalidatable must register under the tag. Half-migration = permanent stale (until cache TTL). Migrate one entity surface at a time: change all Set sites first, then switch invalidate path. Existing tagged surfaces:
- `infrarepoShared.UserFeedsTag(uid)` — `thread:userfeed/usermedia/following:<uid>:*` (Set sites in `infrastructure/repository/thread/thread_repository_user_feed.go`)
- `infrarepoShared.UserForYouTag(uid)` — `feed:foryou:user:<uid>:src/newcount:*` (Set sites in `infrastructure/repository/thread/thread_repository_for_you.go`)
- `infrarepoShared.ThreadFeedsTag(tid)` — feed caches containing thread (Set sites in `infrastructure/repository/thread/thread_repository_*.go` via `storeFeedIDs`/`trackFeedCacheKey`)
- `infrarepoShared.InteractionThreadCacheKeysKey(table, tid)` — like/bookmark/repost listings containing thread (Set in `infrastructure/repository/_shared/interaction_query.go`)

**Profile-meta overlay** — `FeedBuilder.AuthorField` in `webapi/utils/feed_builder.go` overlays `is_premium` + `faculty` from `user:meta:<uid>` (`ProfileMetaCacheKey`) onto cached threads. Lets `thread:id:*`, `thread:chain:*`, `thread:parent:*`, and feed-list caches stay cached across `UpdatePremium` / `UserFaculty` mutations — only `user:meta:<uid>` needs invalidation.

Overlay correctness requires **both** at the controller call site:
1. `FeedBuilder.ProfileSvc` set (non-nil) — `webapi/controller/_shared/controller_builder_helpers.go:NewFeedBuilder` wires this; bespoke `FeedBuilder{...}` literals must include it.
2. `fb.LoadAvatarsForUserIDs(threadctrl.ThreadUserIDs(threads))` called before `ToFeedItem`/`AuthorField`.

Missing either → falls back to embedded `User.IsPremium` / `User.UserFaculty` from cached thread = stale.

**Mutation → invalidation cheatsheet** — surfaces that snapshot mutable user/entity fields:

| Mutation | DTOs that snapshot it | Required invalidation |
|---|---|---|
| `User.IsPremium` (`UpdatePremium`, `ExpireForDev`, trial activation) | `threaddto.FeedAuthor`, `reportdto.ReportUserResponse`, `notificationdto.NotificationSenderResponse`, `communitydto.CommunityAbout/Member/Request/UserResponse`, `exploredto.ExploreUserResponse`, `followdto.FollowUserResponse`, `userprofiledto.UserProfileResponse` | `user:id:<uid>`, `user:meta:<uid>`, `notification:sender:<uid>`, `reports:all`, `reports:id:*`, `follows:user:*`, `community:about/users/requests/id:v2:*`, `community:threads:top/latest:*`, `explore:*`, all thread feed families (via `invalidateUserThreadCaches`) |
| `UserFaculty.Save/Delete` (assignment) | `FeedAuthor.Faculty`, `ReportUserResponse.Faculty`, `NotificationSenderResponse.Faculty`, `UserProfileResponse.Faculty` | `userfaculty:user:<uid>`, `user:meta:<uid>`, `user:id:<uid>`, `profile:user:<uid>`, `notification:sender:<uid>`, `reports:all/id:*`, all thread feed families |
| `Faculty.Update/Delete` (entity rename) | every overlay reads `user:meta:<uid>.FacultyName` | `faculty:id:<id>`, `faculty:list:*`, `profile:user:*`, `userfaculty:user:*`, **`user:meta:*`** (pattern — affects all users in faculty), `notification:sender:*`, `explore:*`, all thread feed families incl. `thread:id:*`, `thread:chain:*`, `feed:trending:global`, `feed:foryou:*` |
| `Event.Update/Delete` | `threaddto.FeedItemResponse.EventID/EventName` (cached threads preload `ThreadEvents.Event`) | `invalidateEvent(id)` (lists + `event:threads:<id>:*`) **plus** `invalidateEventThreadCaches(id)` per affected thread |

Note: `feed:trending:global` (scheduler-written) and `feed:foryou:user:<uid>:src:*` (per-viewer foryou source) cache full `[]threadmodel.Thread` with full preload — they snapshot event_name/is_premium/faculty but do **not** register under `threadFeedsTag`. Mutations affecting fields embedded in cached threads must pattern-delete `feed:foryou:*` and direct-delete `feed:trending:global` explicitly. Overlay handles user fields, but event fields have no overlay → cache invalidation is the only fix.

**File validation** — lives in `domain/model/<m>/`. `mediamodel.ValidateImageExt` (jpg/jpeg/png/webp) in `media/media.go`. `threadmediamodel.ValidateThreadMediaExt` in `thread_media/thread_media.go`. Never in application layer.

**File upload** — endpoints accepting files use `ctx.FormValue` + `ctx.FormFile` (multipart). No-file endpoints use `ctx.BodyParser` (JSON).

**Storage** — `storageuc.StorageUsecase` resolves bucket via `bucketWhitelist` in `application/usecase/storage/storage_service.go` → env var → MinIO. Policy (`public`/`private`) via `MINIO_POLICY_<BUCKET_UPPER>`. Buckets auto-created on first upload.

**Auth** — JWT middleware at `webapi/middleware/auth.go`. Claims: `utils.Get(ctx)` from `webapi/utils/ctrl_helpers.go` → `Claims{UserID, Username, Email, RoleID, RoleName, SID}`.

**Pagination** — services return `([]T, int64, error)`. Controllers wrap in `<m>dto.PaginatedXxxResponse{Data, Meta: threaddto.FeedMeta{Total, Page, Limit}}`.

**For You feed ordering** — `FindForYouFeed` in `infrastructure/repository/thread/thread_repository_for_you.go` splits candidates into two bands by `created_at` age vs `forYouFreshWindow` (15m):
- `fresh` (age < 15m) → sorted `created_at DESC` → capped at `forYouFreshCap` (5); overflow spills into `ranked`
- `ranked` (age ≥ 15m or fresh overflow) → sorted `feed_score DESC` + `shuffleWithinBands`
- Final list = `[fresh..., ranked...]` → `forYouCandidateCap` (800) → `applyAuthorCap` (20/author)

Source SQL (`findForYou{Own,Following,Community,Global}`) uses `forYouFreshFirstOrderSQL()` — `ORDER BY (created_at > NOW() - INTERVAL '<window>') DESC, feed_score DESC` — so fresh posts with `feed_score = 0` are not cut by the per-source `LIMIT 100`. Non-own fresh candidates get a floor of `forYouFreshFloor` (0.1) in `addCandidates` to clear the `forYouMinScore` (0.01) filter; own fresh candidates get floor `5.0`. Visibility predicates (community visibility, `is_private` author, follower / member joins) are unchanged — fresh-first only reorders, never widens the WHERE clause.

`GET /threads/feed/new-count?since_id=<id>` (`CountNewSinceID`) uses strict `thread_id > since_id` — frontend must pass `Math.max(...visibleIDs)` not `data[0].id`, because fresh band sorts by `created_at` not by ID.

**Error handling:**
```go
"forbidden"     → fiber.StatusForbidden
"xxx not found" → fiber.StatusNotFound
else            → fiber.StatusBadRequest
```

**Mutation responses** — `fiber.Map{"message": "xxx successfully"}`. GET endpoints return full DTOs.

**Counters** — `gorm.Expr("GREATEST(count - 1, 0)")` for decrement. Patch Redis after DB update.

**WebSocket** — hub in `infrastructure/socket/hub.go` implements `socket_hubsvciface.ISocketHub`. Handlers in `webapi/socket/`. Push notifications + chat events via Socket.io.

**AI summaries** — `infrastructure/ollama/ollama_client.go` + `application/usecase/thread_summary/thread_summary_service.go`. Ollama generates thread summaries; seeded via `infrastructure/seeder/thread_summary_seeder.go`.

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
5. DTOs in `webapi/dto/<m>/` only — never in Domain or Application
6. New feature → its own per-module subdir in EVERY layer above. Never add files directly to a layer's root dir (except `cmd/main.go` and `webapi/router/router.go`)
7. Cross-cutting code goes in `_shared/` subdirs: `infrastructure/repository/_shared/` (`infrarepoShared`), `webapi/controller/_shared/` (`ctrlshared`), `webapi/dto/_shared/` (`utils`)
