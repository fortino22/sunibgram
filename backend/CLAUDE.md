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
- `infrarepoShared.ThreadFeedsTag(tid)` — feed caches containing thread (Set sites in `infrastructure/repository/thread/thread_repository_*.go` via `storeFeedIDs`/`trackFeedCacheKey`)
- `infrarepoShared.InteractionThreadCacheKeysKey(table, tid)` — like/bookmark/repost listings containing thread (Set in `infrastructure/repository/_shared/interaction_query.go`)

**For You version counter** (replaces the legacy `UserForYouTag`) — `infrarepoShared.ForYouVersionKey(uid)` is a per-user atomic counter. Readers GET the version and bake it into every `feed:foryou:user:<uid>:src/newcount:v<N>` key; mutations call `infrarepoShared.IncrForYouVersion(ctx, cache, uid)` (single O(1) Redis INCR). Old `:v<N-1>` entries orphan and TTL-expire (30–60 s). No tag set, no SMEMBERS/DEL fanout.

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

Note: `feed:foryou:user:<uid>:src:v<N>:*` (per-viewer foryou source) caches full `[]threadmodel.Thread` with full preload — snapshots event_name/is_premium/faculty but does **not** register under `threadFeedsTag`. Invalidate by `IncrForYouVersion(viewerID)` (single Redis INCR; old `:v<N-1>` entries orphan and TTL-expire). Author-mutations that affect followers must INCR every follower's version (see `invalidateForYouFeedCachesForFollowers`).

`feed:trending:global` is a Redis **ZSET** of top-100 `thread_id → feed_score` pairs (scheduler-written; ~3 KB). Reader hydrates full threads via `thread:id:<id>` per-thread cache (`hydrateThreads`). To invalidate when fields embedded in cached threads change (event name, premium, faculty), either direct-delete `feed:trending:global` (scheduler refreshes within 1 min) OR invalidate the per-thread caches with `thread:id:*` (next hydration goes to DB). Overlay handles user fields on the hydrated thread; event fields have no overlay → per-thread cache invalidation is the only fix.

**File validation** — lives in `domain/model/<m>/`. `mediamodel.ValidateImageExt` (jpg/jpeg/png/webp) in `media/media.go`. `threadmediamodel.ValidateThreadMediaExt` in `thread_media/thread_media.go`. Never in application layer.

**File upload** — endpoints accepting files use `ctx.FormValue` + `ctx.FormFile` (multipart). No-file endpoints use `ctx.BodyParser` (JSON).

**Storage** — Two-bucket model: `media` (public, CloudFront-fronted) and `private` (presigned only). The bucket is derived from the S3 key prefix at the storage-layer (`infrastructure/storage/s3_client.go`); it is **not** stored anywhere outside that file. Upload type → prefix mapping lives in `bucketPrefixMap` in `application/usecase/storage/storage_service.go`. Storage interface (`domain/service/storage/i_service.go`) takes only `objectName` (e.g. `"avatar/<uuid>.jpg"`, `"chat/<uuid>.mp4"`) — no bucket param. The S3 client routes `chat/*` and `document-event/*` keys to the private bucket; everything else goes to the media bucket. Credentials resolve via AWS default chain: env vars → shared config/credentials file (incl. SSO via `AWS_PROFILE`) → EC2/ECS task role. **No static keys in env.** Local dev: `aws sso login --profile sunibgram-dev` + `AWS_PROFILE=sunibgram-dev`. Prod: attach IAM role to task; unset `AWS_PROFILE`. Buckets pre-provisioned via the separate `sunibgram-infra` Terraform repo (private, OAC-fronted). `S3_CDN_BASE_URL` (CloudFront domain) is the host used by `GetURL` for media-bucket keys; URL format = `<cdn>/<object>`. Presigned URLs (`GeneratePresignedURL`) sign against the S3 regional endpoint via the SDK presigner — used for sensitive content (chat, document_event).

**Auth** — JWT middleware at `webapi/middleware/auth.go`. Claims: `utils.Get(ctx)` from `webapi/utils/ctrl_helpers.go` → `Claims{UserID, Username, Email, RoleID, RoleName, SID}`.

**Pagination** — services return `([]T, int64, error)`. Controllers wrap in `<m>dto.PaginatedXxxResponse{Data, Meta: threaddto.FeedMeta{Total, Page, Limit}}`.

**For You feed ordering** — `FindForYouFeed` in `infrastructure/repository/thread/thread_repository_for_you.go` splits candidates into two bands by `created_at` age vs `forYouFreshWindow` (15m):
- `fresh` (age < 15m) → sorted `created_at DESC` → capped at `forYouEffectiveFreshCap(limit) = min(forYouFreshCap=5, ceil(0.2 * limit))`; overflow spills into `ranked`
- `ranked` (age ≥ 15m or fresh overflow) → sorted `feed_score DESC` + `shuffleWithinBands`
- Final list = `[fresh..., ranked...]` → `forYouCandidateCap` (800) → `applyAuthorCap` (position-aware: 3 in top-10, 5 in top-25, 20 across whole feed)

Source SQL (`findForYou{Own,Following,Community,Global}`) uses `forYouFreshFirstOrder(freshCutoff)` — `ORDER BY (created_at > ?) DESC, feed_score DESC` (parameter-bound) — so fresh posts with `feed_score = 0` are not cut by the per-source `LIMIT 100`. Non-own fresh candidates get a floor of `forYouFreshFloor` (0.1) in `addCandidates` to clear the `forYouMinScore` (0.01) filter; own fresh candidates get floor `5.0`. Replies are demoted via `forYouReplyWeight` (0.5) applied to the effective source weight. Visibility predicates (community visibility, `is_private` author, follower / member joins) are unchanged — fresh-first only reorders, never widens the WHERE clause.

`feed_score` stores **raw weighted engagement** (no decay applied); read-time decay `1/(1+age_h/forYouHalflifeHours)` is the only decay. Scheduler (`infrastructure/scheduler/feed_score_scheduler.go`) recomputes every 1 min and writes the global trending ZSET (`feed:trending:global`) — top-100 (thread_id, feed_score) pairs; readers hydrate via `thread:id:<id>` per-thread cache.

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

Two buckets only — pre-provisioned via Terraform in the separate `sunibgram-infra` Terraform repo (private, OAC-fronted CloudFront for media; private S3-only for sensitive). Source of truth: `bucketPrefixMap` in `application/usecase/storage/storage_service.go`. The S3 client (`infrastructure/storage/s3_client.go`) routes keys with `chat/` or `document-event/` prefix to the private bucket; everything else goes to the media bucket. **Bucket is never stored in DB** — it's derived from the object key prefix.

| Bucket | Env Var | Bucket (dev) | Access |
|---|---|---|---|
| media | `S3_BUCKET_MEDIA` | `sunibgram-media-dev` | CloudFront (`S3_CDN_BASE_URL`) |
| private | `S3_BUCKET_PRIVATE` | `sunibgram-private-dev` | S3 presigned only |

| Upload Type | Key Prefix | Target Bucket |
|---|---|---|
| `avatar` | `avatar/` | media |
| `banner` | `banner/` | media |
| `post` | `post/` | media |
| `poll` | `poll/` | media |
| `community_banner` | `community-banner/` | media |
| `event_banner` | `event-banner/` | media |
| `group_avatar` | `group-avatar/` | media |
| `chat` | `chat/` | private (presigned) |
| `document_event` | `document-event/` | private (presigned) |

CloudFront single distribution fronts the media bucket only (no viewer-request function needed — keys map 1:1). `GetURL` builds `<cdn>/<object>`. `GeneratePresignedURL` signs against the S3 regional endpoint for keys in the private bucket.

---

## Architecture Rules

1. Business logic → Domain model methods, never Application
2. WebAPI imports Domain interfaces only, never Infrastructure
3. Domain imports standard lib only — zero external packages
4. Application knows Domain interfaces only, never concrete Infrastructure structs
5. DTOs in `webapi/dto/<m>/` only — never in Domain or Application
6. New feature → its own per-module subdir in EVERY layer above. Never add files directly to a layer's root dir (except `cmd/main.go` and `webapi/router/router.go`)
7. Cross-cutting code goes in `_shared/` subdirs: `infrastructure/repository/_shared/` (`infrarepoShared`), `webapi/controller/_shared/` (`ctrlshared`), `webapi/dto/_shared/` (`utils`)
