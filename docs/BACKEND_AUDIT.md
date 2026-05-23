# Sunibgram Backend Audit — Security + Performance

Date: 2026-05-23
Scope: `/backend` — Go/Fiber, Postgres, Redis, MinIO
Rule for fixes: **no business-logic changes**. Every fix is a drop-in code/config/index change.

---

# Part 1 — Security Vulnerabilities

Severity rubric — **CRITICAL**: auth/RCE/full data exposure · **HIGH**: privilege escalation, account takeover · **MEDIUM**: info disclosure, brute force vectors · **LOW**: defense-in-depth gaps.

## CRITICAL

### C1. Hardcoded fallback JWT secret allows token forgery
- `backend/webapi/middleware/auth_middleware.go:119-124`
- `backend/application/usecase/auth/auth_service.go:388-393`
- `backend/webapi/socket/socket.go:187-192`

All three `jwtSecret()` helpers silently fall back to the literal `"sunibgram-secret"` when the env var is empty. If `JWT_SECRET` is ever unset, anyone can forge admin JWTs with a publicly-known key.

**Fix:** remove the fallback. Mirror the `corsAllowOrigins` pattern in `server.go:51`:
```go
s := os.Getenv("JWT_SECRET")
if s == "" { log.Fatal("JWT_SECRET is required") }
```

### C2. OTP HMAC secret falls back to a hardcoded literal
- `backend/application/usecase/verification/verification_service.go:32-40`

`otpHMACSecret()` last-resort returns `[]byte("sunibgram-otp")`. If both `OTP_HMAC_SECRET` and `JWT_SECRET` are unset, every login/register/password-reset OTP hash is computable offline.

**Fix:** `log.Fatal("OTP_HMAC_SECRET or JWT_SECRET is required")` instead of returning the literal.

### C3. Login/Register/ForgotPassword have no IP rate limit
- `backend/webapi/router/auth/auth_routes.go:24-32`

Only `/auth/login/otp/send` has the `otpSendRateLimit` middleware. `/auth/login`, `/auth/login/prod`, `/auth/register`, `/auth/refresh`, `/auth/forgot-password` are unrestricted. `LoginProd` (`auth_service.go:269-297`) is the admin-credential path — brute-forceable. The `otpFailLimiter` (`auth_service.go:73`) is keyed by **email**, so credential-stuffing across emails is free.

**Fix:** add an IP-keyed limiter on every `/auth/*` POST:
```go
ipLimiter := ratelimit.New(c, "auth_ip", 10, 5*time.Minute) // 10 req / 5 min / IP
auth.Post("/login", ipLimiter, controller.Login)
auth.Post("/login/prod", ipLimiter, controller.LoginProd)
auth.Post("/register", ipLimiter, controller.Register)
auth.Post("/forgot-password", ipLimiter, controller.ForgotPassword)
```

## HIGH

### H1. JWT accepted from query string — leaks via logs / Referer
- `backend/webapi/socket/socket.go:73-77, 126-129`

WS upgrade + `authenticateSocket` accept the access token from `?token=` / `?access_token=`. Query strings are recorded by reverse proxies, CDNs, browser history, and `Referer` headers — long-lived access tokens get leaked.

**Fix:** drop the query-string branches. Keep only `Authorization: Bearer` and `Sec-WebSocket-Protocol` (already implemented at lines 69-71/121-123).

### H2. SMTP outbound TLS verification disabled (MitM on email)
- `backend/infrastructure/email/smtp_client.go:43-49`

`tls.Config{InsecureSkipVerify: true}` on STARTTLS. SMTP creds (`PlainAuth` line 66) plus all OTP / password-reset content are MitM-able.

**Fix:** set `InsecureSkipVerify: false`. The `ServerName: c.host` already pins to the configured host, so verification works.

### H3. Midtrans webhook signature compare is not constant-time
- `backend/webapi/controller/order/order_controller.go:144-158`

`if incomingSig != expected` is byte-by-byte. Endpoint is unauthenticated (`order_routes.go:14`). Timing-oracle for forging payment-success callbacks that mutate premium state.

**Fix:**
```go
import "crypto/subtle"
if subtle.ConstantTimeCompare([]byte(incomingSig), []byte(expected)) != 1 { ... }
```

### H4. Open authenticated SMTP relay via `/api/email/send`
- `backend/webapi/controller/email/email_controller.go:21-45`
- `backend/webapi/router/email/email_routes.go:11`

Any logged-in `@binus.ac.id` user can send arbitrary `to/subject/body` through the platform SMTP. Spam/phishing vector that abuses the platform's reputation.

**Fix:** restrict the route — either `middleware.AdminOnly()`, or hard-cap `to` to `claims.Email` and add `ratelimit.New(c, "email_send", 5, time.Hour)`.

### H5. Token algorithm check too loose — accepts any HMAC variant
- `backend/webapi/middleware/auth_middleware.go:95-117`
- `backend/application/usecase/auth/auth_service.go:360-385`

`parseAccessToken`/`parseToken` allow any `*jwt.SigningMethodHMAC` (HS256/HS384/HS512). `signToken` only issues HS256.

**Fix:**
```go
if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
    return nil, fmt.Errorf("unexpected signing method")
}
```

## MEDIUM

### M1. `utils.Get` panics on malformed JWT claim types
- `backend/webapi/utils/ctrl_helpers.go:29-36`

`raw["user_id"].(float64)` and `raw["role_id"].(float64)` are unchecked type assertions. Malformed token → panic. No `recover` middleware in `server.go:34-45`, so panic crashes the worker.

**Fix:** guarded assertion + mount `recover.New()`:
```go
uid, ok := raw["user_id"].(float64)
if !ok { return Claims{}, false }
```
And in `cmd/bootstrap/server.go`:
```go
app.Use(recover.New())
```

### M2. `/verification/send` and `/verification/verify` lack IP rate limit
- `backend/webapi/router/verification/verification_routes.go:9-13`

Inner `sendLimiter`/`verifyLimiter` (`verification_service.go:86-87`) keyed by `type:email`. Attacker rotating the email field has unlimited send/verify from one IP. Also enables email enumeration via differential error messages (`verification_service.go:106-111`).

**Fix:** wrap both routes in an IP-keyed limiter like the one in `auth_routes.go:16-22`.

### M3. Email field disclosed to anyone via `GET /api/profile/:id`
- `backend/webapi/controller/user_profile/user_profile_controller.go:213`
- `backend/webapi/router/user_profile/user_profile_routes.go:14` (`OptionalAuth`)

`UserProfileResponse.Email` is populated for every profile lookup including public path. Every `@binus.ac.id` becomes enumerable.

**Fix:** remove `Email` from `userprofiledto.UserProfileResponse`. Expose email only on `/profile/me` and only when `viewerID == userID`.

### M4. `Email` shipped in admin list — verify response shape isolation
- `backend/webapi/controller/user/user_controller.go:36-44`

Fine for admin route (`admin_routes.go:25-26`), but ensure `userdto.UserResponse` is not reused on public routes.

**Fix:** confirm reuse; if shared, add `json:"email,omitempty"` and a viewer-aware redactor.

### M5. Pagination `page` is unbounded → huge OFFSET enables resource exhaustion
- `backend/webapi/utils/ctrl_helpers.go:67-69`
- `backend/webapi/controller/_shared/ctrl_helpers.go:38-40`

`offset := (page - 1) * limit` is used in repos (e.g. `thread_repository_query.go:308`). `page=999999999` triggers DB scan / pointless work.

**Fix:** cap `page <= 10000` in both pagination parsers.

### M6. Verbose service errors bypass `safeErrorMessage`
- `backend/webapi/controller/_shared/ctrl_helpers.go:55-76`

`safeErrorMessage` filters `sql:`/`gorm`/`pq:`, but bare `err.Error()` is returned at `auth_controller.go:69,113`, `user_profile_controller.go:140`, `order_controller.go:54` etc.

**Fix:** replace bare `JSON(...ErrorResponse{Message: err.Error()})` with `HandleServiceError(ctx, err)`.

### M7. `/api/orders/notification` allows webhook flood + lacks idempotency
- `backend/webapi/router/order/order_routes.go:13`
- `backend/webapi/controller/order/order_controller.go:127-142`

Unauthenticated, no rate limit, no idempotency key tracking.

**Fix:**
```go
orders.Post("/notification", ratelimit.New(c, "midtrans_webhook", 60, time.Minute), controller.HandleNotification)
```
And in `HandleNotification`, `SETNX order:notif:<order_id>` in Redis with 24h TTL to skip duplicates.

## LOW

### L1. CORS `AllowCredentials: true` — reject literal `*`
- `backend/cmd/bootstrap/server.go:36-52`

`log.Fatal` on empty is good, but does not reject `*`. Combined with credentials it is a real misconfig hazard.

**Fix:** in `corsAllowOrigins()`, `if raw == "*" { log.Fatal("CORS wildcard incompatible with credentials") }`.

### L2. JWT in localStorage (XSS-exposed by design)
JWTs returned in JSON (`auth_controller.go:147-151`); frontend stores in localStorage. Documented choice — flag only.
**Fix (optional):** migrate to HttpOnly + Secure + SameSite=Strict cookies if threat model warrants.

### L3. No `recover` middleware
- `backend/cmd/bootstrap/server.go:24-45`

Tied to M1 — fix together: `app.Use(recover.New())`.

### L4. Global `BodyLimit: 50 MiB`; per-upload caps checked after buffering
- `backend/cmd/bootstrap/server.go:26`

Attacker can flood any endpoint with 50 MiB before the controller rejects.

**Fix:** lower global to 25 MiB; add per-route `bodylimit.New(...)` matching the smallest applicable cap (`MaxAvatarUploadSize = 5MB`, etc.).

### L5. `Subject`/`Body` accept CRLF — header injection vector
- `backend/webapi/controller/email/email_controller.go:25-37`

`\r\n` in `Subject` enables header smuggling (BCC, content-type swap).

**Fix:** `if strings.ContainsAny(req.Subject, "\r\n") { return 400 }` (also for sender-controlled body if used as headers downstream).

## Audited Clean
- SQL injection: all queries parameterized; sort whitelist at `thread_repository_query.go:249-258`
- Password hashing: bcrypt (`auth_service.go:125,155,181`)
- OTP compare: `hmac.Equal` constant-time (`verification_service.go:63`)
- File uploads: UUID rename (`storage_service.go:53`), content sniffing (`media.go:48-61`), extension + bucket whitelist
- Password JSON tag: `json:"-"` (`user.go:24`)
- AES-GCM chat: 12-byte nonces, `crypto/rand` (`chat_crypto.go:45-87`)
- `exec.Command`: not used

---

# Part 2 — High-Cost APIs (Perf Fixes Without Logic Change)

Impact rubric — **HIGH**: request-path, frequent, scales with data · **MEDIUM**: request-path but bounded · **LOW**: cleanup / micro-opts.

## HIGH

### P-H1. Synchronous notification fan-out in chat `Send` blocks request
- `backend/application/usecase/chat/chat_service.go:87, 91-135`

`sendMessageNotifications` runs inline. For a group of N members it does N `FindByID` + N `SendNotification` (each: profile cache lookup, DB INSERT, hub emit). Sender's HTTP response waits for all of it.

**Fix:** wrap the call in a goroutine with recover:
```go
go func() {
    defer func() { if r := recover(); r != nil { log.Printf("notif fanout panic: %v", r) } }()
    s.sendMessageNotifications(room, chat)
}()
```

### P-H2. Synchronous SMTP send on OTP endpoints
- `backend/application/usecase/verification/verification_service.go:138-142`
- `backend/application/usecase/email/email_service.go:32-33`

`SendCode` calls `SendWelcomeOTP/SendLoginOTP/SendPasswordResetOTP` — blocking `DialAndSend`. On the login/register/reset critical path. A `QueueMail`/`BLPop` worker already exists at `email_service.go:36-61` but is unused for OTP.

**Fix:** in `verification_service.go:138-142`, swap to `s.emailService.QueueMail(...)`. Render body inline. Ensure `StartWorker` is wired in `cmd/main.go`.

### P-H3. Synchronous interaction notifications on every like/follow/repost/reply
- `backend/application/helper/notification_helper.go:36-42`
- callers: `like_service.go:83`, `follow_service.go:46`, `thread_service_command.go:138,303`, `chat_service.go:126`

`SendInteractionNotification` runs `userRepo.FindByID` + senderProfile (≤3 repo calls on cache miss) + INSERT + hub emit on every interaction.

**Fix:** one-line change in the helper body:
```go
go func() {
    defer func() { _ = recover() }()
    _ = notificationService.SendNotification(...)
}()
```
No caller changes needed.

### P-H4. Follower/community fan-out cache invalidation blocks writes
- `backend/infrastructure/repository/thread/thread_cache_invalidation.go:144-168`
- callers: `thread_repository_write.go:50, 56, 198, 200`

`invalidateForYouFeedCachesForFollowers` plucks every `follower_id` then calls `InvalidateTag` per follower (SMEMBERS + DEL per follower). For 50k followers → 50k Redis round-trips per like/repost/reply. Same for community members.

**Fix (quick win):** goroutine + recover at the four call sites in `thread_repository_write.go`. Better long term: store a single per-author "feeds dirty since" timestamp; readers check it.

### P-H5. `Exists` checks use `COUNT(*)` instead of `EXISTS`
- `backend/infrastructure/repository/_shared/interaction_query.go:81-85` (likes/bookmarks/reposts)
- `backend/infrastructure/repository/follow/follow_repository.go:136-143`
- `backend/infrastructure/repository/user/user_repository.go:101-111`
- `backend/infrastructure/repository/thread/thread_repository_write.go:128-140`

Existence checks (`Exists`, `ExistsByEmail`, `ExistsByUserName`, `validateThreadEventExists`) do `COUNT(*)`. Postgres keeps scanning matched rows even when one is enough.

**Fix:**
```go
var exists bool
db.Raw("SELECT EXISTS(SELECT 1 FROM likes WHERE user_id=? AND thread_id=? LIMIT 1)", uid, tid).Scan(&exists)
```
Or GORM `.Select("1").Limit(1).Scan(&dest)`.

### P-H6. R/W split violated — writes routed to readDB
- `backend/infrastructure/repository/follow/follow_repository.go:269, 364`

`GetStat` and `patchStatCache` call `r.readDB.FirstOrCreate(...)`. `FirstOrCreate` issues an `INSERT` if missing — against the read replica.

**Fix:** swap `r.readDB.FirstOrCreate` → `r.writeDB.FirstOrCreate`. Or split: `Find` on readDB, `Create` on writeDB.

### P-H7. Missing composite index on `notifications`
- `backend/infrastructure/repository/notification/notification_repository.go:45-55`
- model has only single-column indexes (`domain/model/notification/notification.go:14-16`)

`WHERE recipient_user_id = ? ORDER BY created_at DESC` cannot use both indexes together → index scan + sort.

**Fix:** add to `infrastructure/migration/migrate.go`:
```sql
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_created
    ON notifications(recipient_user_id, created_at DESC);
```

### P-H8. Missing indexes on `threads.parent_id` and `threads.original_thread_id`
- `backend/infrastructure/repository/thread/thread_repository_query.go:108, 131`
- `thread_repository_write.go:26-34, 213-220`

Reply trees and repost lookups degenerate to seq scan as the table grows.

**Fix:** add to `migrate.go`:
```sql
CREATE INDEX IF NOT EXISTS idx_threads_parent_id
    ON threads(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_threads_original_thread_id
    ON threads(original_thread_id) WHERE original_thread_id IS NOT NULL;
```

### P-H9. Chat `MarkRead` issues N serial UPDATEs
- `backend/application/usecase/chat/chat_service.go:267-276`

After `BatchMarkRead`, loops calling `UpdateIsRead(chatID)` per chat — hundreds of UPDATEs per call in a busy group.

**Fix:** collect IDs that crossed the threshold, then one statement:
```go
db.Exec(`UPDATE chats SET is_read = true WHERE chat_id IN (?)`, ids)
```

### P-H10. ThreadSummary worker over-hydrates the entire reply tree
- `backend/application/usecase/thread_summary/thread_summary_service.go:129-141`

`topReplies` calls `FindByRootID(threadID, 1, 50)` which preloads User, Community, OriginalThread.*, ThreadEvents.Event — then iterates and filters direct children in Go. The OpenAI prompt only needs `Content`.

**Fix:** add a lean repo method:
```go
func (r *repo) FindRepliesContent(ctx context.Context, rootID uint, limit int) ([]string, error) {
    var out []string
    return out, r.readDB.Raw(`SELECT content FROM threads WHERE parent_id = ? ORDER BY like_count DESC LIMIT ?`, rootID, limit).Scan(&out).Error
}
```

### P-H11. `feed_score_scheduler` runs heavy work every minute with no idle guard
- `backend/infrastructure/scheduler/feed_score_scheduler.go:25-83`

Every 60s: rewrites `feed_score` for the entire 7-day window, zeroes older rows (every minute, every row), re-runs full preloaded `findForYouGlobal`, then `DeleteByPattern("feed:foryou:*")` (full Redis SCAN).

**Fix:**
1. Skip recompute if no thread mutation since last tick (Redis counter incremented by `Save`/`Delete`/like).
2. Drop `feed:foryou:*` pattern delete — per-user keys already TTL within 30-60s.
3. Add `AND feed_score <> 0` to the zero-out UPDATE so only newly-aged rows are touched.

## MEDIUM

### P-M1. `FindCommunityMedia` returns full unpaginated dataset
- `backend/infrastructure/repository/thread/thread_repository_query.go:334-361`

No LIMIT/OFFSET on `SELECT ... FROM thread_media JOIN threads ... WHERE community_id = ?`.

**Fix:** add `page, limit int` params, `.Offset((page-1)*limit).Limit(limit)`; thread through service + controller.

### P-M2. Explore endpoints hard-capped to 20, no pagination support
- `backend/application/usecase/explore/explore_service.go:18, 80, 96, 112, 120, 132`

`defaultExploreLimit = 20`; controllers cannot ask for page 2. Search ILIKE `%query%` is full-table scan re-executed every 30s.

**Fix:** wire `page` through repo (cache key already supports it via `exploreThreadSectionKey`). Add cursor `WHERE created_at < ?` + `LIMIT 20 OFFSET (page-1)*20`.

### P-M3. `FindByRootID` sorts without supporting index
- `backend/infrastructure/repository/thread/thread_repository_query.go:222`

Used on every thread detail load + ThreadSummary worker. No covering index for `WHERE thread_id = ? OR root_id = ? ORDER BY like_count DESC, created_at DESC`.

**Fix:**
```sql
CREATE INDEX IF NOT EXISTS idx_threads_root_like
    ON threads(root_id, like_count DESC, created_at DESC) WHERE root_id IS NOT NULL;
```

### P-M4. Notification list never cached
- `backend/infrastructure/repository/notification/notification_repository.go:45-55`

Bell badge polls this — every poll hits DB.

**Fix:** cache page 1 with key `notification:list:user:<uid>:page:1:limit:<n>`, TTL 30s, registered under tag `notification:user:<uid>`. Invalidate the tag in `Save` and `Delete` (`:25-43`).

### P-M5. `FindForYouFeed` runs 4 serial source queries on cache miss
- `backend/infrastructure/repository/thread/thread_repository_for_you.go:165-188`

Fires `findForYouOwn`, `findForYouFollowing`, `findForYouCommunity`, `findForYouGlobal` sequentially. Each preloads 5 joins + 2 preloads at LIMIT 100. Cache TTL 30-60s → fires often during scroll.

**Fix:** run the four calls in parallel via goroutines + `sync.WaitGroup`. They share no state until merge.

### P-M6. `user_relationships` lacks `(uid, created_at DESC)` composite
- `backend/infrastructure/repository/follow/follow_repository.go:285-303, 319-340`

`GetFollowing`/`GetFollowers` order by `created_at DESC`. Existing single-column index does not satisfy sort.

**Fix:**
```sql
CREATE INDEX IF NOT EXISTS idx_user_rel_follower_created
    ON user_relationships(follower_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_rel_following_created
    ON user_relationships(following_id, created_at DESC);
```

### P-M7. Eager invalidation of trending hashtags on every thread mutation
- `backend/application/usecase/thread/thread_service_command.go:135, 178, 209, 252, 302, 324`
- `backend/infrastructure/repository/hashtag/hashtag_repository.go:256-258`

Trending recomputes are 5-min TTL'd but invalidation on every mutation guarantees misses in busy periods.

**Fix:** drop eager invalidation in `thread_service_command.go`. Let the 5-min TTL expire naturally. Or move to goroutine.

### P-M8. `GetFollowingSet` deserializes 100 KB JSON per cache hit
- `backend/infrastructure/repository/follow/follow_repository.go:145-183`

Caches `map[uint]bool` of every followee under one key. 10k followees ≈ 100 KB blob deserialized even when caller needs ~20 ids.

**Fix:** store as Redis SET (`SADD user:following:<uid> <id>...`) and use `SMISMEMBER` (already exists at `infrastructure/cache/cache.go:227`).

### P-M9. `S3Storage.ensureBucket` round trips per upload
- `backend/infrastructure/storage/s3_client.go:90-126, 128-131`

Every `Upload` calls `BucketExists` — extra S3 round trip in dev mode.

**Fix:** memoize known buckets in a `sync.Map` on the struct; short-circuit after first success per bucket.

### P-M10. Chat preload returns full user rows (incl. password hash)
- `backend/infrastructure/repository/chat/chat_repository.go:24-73`

`Preload("Sender").Preload("ReplyTo.Sender")` with no projection. Relies on JSON tags to strip sensitive fields.

**Fix:**
```go
Preload("Sender", func(db *gorm.DB) *gorm.DB {
    return db.Select("user_id, user_name, name, is_premium")
})
```
Same for `ReplyTo.Sender`.

### P-M11. `UpsertHashtags` re-reads IDs after upsert
- `backend/infrastructure/repository/hashtag/hashtag_repository.go:37-52, 200-222`

Two round trips: hand-built variadic INSERT, then `WHERE name IN ?` SELECT to read back IDs.

**Fix:** use `clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "name"}}}` on the upsert.

### P-M12. Notification senderProfile fans into 3 IN-clause lookups
- `backend/infrastructure/repository/notification/notification_repository.go:46-54`
- `backend/application/usecase/notification/notification_service.go:144-156`

Per page load: 3 separate IN-queries to `users`, `user_profiles`, `user_faculties`.

**Fix:** single LEFT JOIN:
```sql
SELECT u.*, p.avatar_bucket, p.avatar_object_name, uf.faculty_id, f.faculty_name
FROM users u
LEFT JOIN user_profiles p ON p.user_id = u.user_id
LEFT JOIN user_faculties uf ON uf.user_id = u.user_id
LEFT JOIN faculties f ON f.id = uf.faculty_id
WHERE u.user_id IN (?)
```

## LOW

### P-L1. Duplicate `feed:trending:global` delete
- `backend/infrastructure/repository/thread/thread_repository_write.go:204`

`invalidateTrendingFeed` already deletes the key; the literal `r.cache.Delete(ctx, "feed:trending:global")` next line is redundant.
**Fix:** remove the duplicate.

### P-L2. Auth `FindByEmail`/`FindByUserName` preload Role
- `backend/infrastructure/repository/user/user_repository.go:89-99`

Login flow extras a SELECT for Role data it doesn't need.
**Fix:** add `FindByEmailMinimal` that selects only `user_id, password, role_id, name, user_name, email, is_premium`.

### P-L3. `verification_codes.expires_at` not indexed
- `backend/infrastructure/scheduler/verification_code_purge.go`

Hourly `DELETE WHERE expires_at < NOW()` does seq scan.
**Fix:** `CREATE INDEX IF NOT EXISTS idx_verification_codes_expires_at ON verification_codes(expires_at);`

### P-L4. Scheduler loops don't honor shutdown
- `backend/infrastructure/scheduler/hashtag_weekly_reset.go:11-26`
- `backend/infrastructure/scheduler/feed_score_scheduler.go:25-34`

Naked `for { sleep }` loops hold a DB conn during sleep and stall SIGTERM.
**Fix:** accept `context.Context`, use:
```go
select {
case <-ctx.Done(): return
case <-ticker.C: ...
}
```

### P-L5. DB pool 50/25 is low for both pools
- `backend/infrastructure/config/db.go:43-46`

Under 200+ concurrent requests × 2-4 queries each, pool exhausts.
**Fix:** bump readDB to `MaxOpenConns=100, MaxIdleConns=50`. Verify Postgres `max_connections`.

### P-L6. `DeleteByPattern("feed:foryou:*")` walks full Redis keyspace
- `backend/infrastructure/scheduler/feed_score_scheduler.go:79`

Same as P-H11 — drop it.

### P-L7. `ExistsByEmail`/`ExistsByUserName` not cached during async signup validation
- `backend/infrastructure/repository/user/user_repository.go:101-111`

Frontend keystroke-validation → many DB Count calls.
**Fix:** cache for 30s under `user:exists:email:<hash>` / `user:exists:username:<lower>`.

### P-L8. `feed_score_scheduler` writes `feed_score = 0` for all old rows every tick
- `backend/infrastructure/scheduler/feed_score_scheduler.go:53`

WAL churn + table bloat.
**Fix:** `UPDATE ... SET feed_score = 0 WHERE created_at <= NOW() - INTERVAL '7 days' AND feed_score <> 0`.

### P-L9. Community queries always self-join for repost fallback
- `backend/infrastructure/repository/thread/thread_repository_query.go:269-272, 309-315`

`LEFT JOIN threads original ON original.thread_id = threads.original_thread_id` + `COALESCE(...)` cannot be index-served. Dead weight when query already excludes REPOST.
**Fix:** branch on `threadTypes` — skip the self-join when REPOST is not in the filter set.

### P-L10. `validateThreadEventExists` uses Count(*) on UUID PK
- `backend/infrastructure/repository/thread/thread_repository_write.go:128-140`

Cheap but unnecessary.
**Fix:** `r.db.readDB.Select("id").Where("id = ?", *eventID).First(&threadmodel.Event{}).Error`.

---

# Quick-Win Order (Highest ROI First)

| # | Item | Effort | Impact |
|---|------|--------|--------|
| 1 | P-H3 — goroutine-wrap `SendInteractionNotification` (1-line helper) | 1 min | every like/follow/repost/comment unblocks |
| 2 | P-H4 — goroutine-wrap follower fan-out invalidation | 5 min | request stops waiting on N×Redis |
| 3 | P-H2 — swap OTP send to existing `QueueMail` | 10 min | login/register p99 drops drastically |
| 4 | P-H1 — goroutine-wrap chat fan-out | 5 min | chat send latency stable for group rooms |
| 5 | P-H7, P-H8, P-M3, P-M6 indexes — single migration commit | 15 min | DB CPU drops, sort spills gone |
| 6 | P-H5 — `EXISTS` rewrite for hot existence checks | 30 min | every interaction faster |
| 7 | P-H6 — fix R/W split in `follow_repository` | 5 min | correctness + read replica safety |
| 8 | C1, C2, C3, H2, H3 — security one-liners | 30 min | shut critical attack surface |
| 9 | H1, H4, H5 — drop query-string JWT, restrict `/email/send`, pin HS256 | 20 min | eliminate token leak + open relay |

---

# Files Touched Summary

**Indexes (one migration file):** `backend/infrastructure/migration/migrate.go`
**Security one-liners:** `auth_middleware.go`, `auth_service.go` (×2), `verification_service.go`, `socket.go`, `smtp_client.go`, `order_controller.go`, `auth_routes.go`, `email_routes.go`
**Perf one-liners:** `notification_helper.go`, `thread_repository_write.go` (4 sites), `verification_service.go`, `chat_service.go`, `follow_repository.go`
**Repo method swaps:** `_shared/interaction_query.go`, `user_repository.go`, `chat_repository.go`, `notification_repository.go`, `thread_summary_service.go`

End of report.

---

# Re-scan Addendum — 2026-05-23

Second pass over `/backend`. Goal: confirm prior findings still present, surface new ones.

## Status of prior findings

All 19 security items (**C1–C3, H1–H5, M1–M7, L1–L5**) and all 20 perf items (**P-H1–P-H11, P-M1–P-M12, P-L1–P-L10**) remain **OPEN**. Zero fixes applied since the original audit. Spot-check evidence (current line numbers):

- C1 `jwtSecret()` fallback: `webapi/middleware/auth_middleware.go:119-124`, `application/usecase/auth/auth_service.go:388-393`, `webapi/socket/socket.go:187-192`
- C2 `otpHMACSecret()` fallback: `application/usecase/verification/verification_service.go:32-40`
- H1 query-string JWT: `webapi/socket/socket.go:72-77, 126-129`
- H3 non-constant-time signature compare: `webapi/controller/order/order_controller.go:155`
- M1 unchecked `.(float64)`: `webapi/utils/ctrl_helpers.go:30,34`
- M3 email field on public profile: `webapi/controller/user_profile/user_profile_controller.go:216`
- P-H6 writes on readDB: `infrastructure/repository/follow/follow_repository.go:266,361`
- P-H9 N serial UPDATEs: `application/usecase/chat/chat_service.go:272-276`
- P-M10 chat Preload full user: `infrastructure/repository/chat/chat_repository.go:34,48,58,69`

Audited-clean items (SQL injection, bcrypt, OTP `hmac.Equal`, UUID file names, AES-GCM, `password json:"-"`) are still clean — no regressions.

## NEW perf findings

### P-H12. `ChatRoomService.CreateGroup` — N+1 user lookup + serial AddMember
- `application/usecase/chat_room/chat_room_service.go:87-98`

Per memberID: one `userRepo.FindByID` then one `roomRepo.AddMember`. For a 50-person group that's 100 round trips on the request path.

**Fix:** batch validate (`SELECT user_id FROM users WHERE user_id IN (?)`), then one INSERT with multi-row VALUES (`db.Create(&members)` where members is `[]RoomMember`).

### P-H13. `ChatRoomService.AddMemberToGroup` — same N+1 pattern
- `application/usecase/chat_room/chat_room_service.go:119-134`

Same as P-H12: per-user `FindByID` + `IsMember` + `AddMember` in a loop.

**Fix:** one IN-query for user existence, one IN-query for current membership, one bulk INSERT for the diff. Final `FindByID(roomID)` reload at `:135` stays.

### P-M13. `ExpireSubscriptions` — serial `UpdatePremium` per owner
- `application/usecase/premium_subscription/premium_subscription_service.go:78-82`

After `DeactivateByOwnerIDs(ownerIDs)` (one bulk UPDATE), loops `userRepo.UpdatePremium(id, false)` per owner.

**Fix:** add `userRepo.UpdatePremiumByIDs(ids []uint, premium bool)` → single `UPDATE users SET is_premium = ? WHERE user_id IN (?)`. Background job, not request-path, but still O(N) UPDATEs + per-row cache patch.

### P-M14. `thread_summary` worker goroutine swallows panics
- `application/usecase/thread_summary/thread_summary_service.go:60-83`

`go func() { s.group.Do(... s.generateSummary(...) ...) }()` has no `defer recover()`. OpenAI client panic or downstream nil deref crashes the worker (no `recover` middleware on background goroutines).

**Fix:** wrap the body in `defer func() { if r := recover(); r != nil { log.Printf(...) } }()`.

## NEW security findings

Deep scan covering IDOR, mass assignment, SSRF, path traversal, WebSocket auth, race conditions, refresh-token reuse, account enumeration, admin gating, JWT claim verification, premium/payment bypass found **no new issues** beyond the original audit. Checks that passed:

- Controller `:id` params: ownership enforced via `claims.UserID` on mutations (thread delete, chat unsend, community member remove, notification delete, bookmark/like).
- WebSocket room joins call `roomService.IsMember(roomID, userID)` (`socket/chat_handler.go`).
- Mass assignment: BodyParser DTOs do not expose `user_id`/`role_id`/`is_premium`/`created_at`.
- Order amount fixed server-side (no client `amount` field).
- Like/Bookmark/Repost protected by composite PK `(user_id, thread_id)`.
- Admin endpoints behind `middleware.AdminOnly()` consistently.

## Updated Quick-Win Order

Append after row 9:

| # | Item | Effort | Impact |
|---|------|--------|--------|
| 10 | P-H12 + P-H13 — batch `CreateGroup` / `AddMemberToGroup` | 30 min | group-chat ops degenerate from O(N) to O(1) |
| 11 | P-M14 — `defer recover` in `thread_summary` goroutine | 1 min | worker crash protection |
| 12 | P-M13 — bulk `UpdatePremiumByIDs` for expiry job | 10 min | nightly job O(1) instead of O(owners) |

End of addendum.
