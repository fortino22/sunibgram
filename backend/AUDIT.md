# Sunibgram Backend — Security & Performance Audit

Scope: `backend/` Go/Fiber service. Static review only. Suggestions improve security/cost **without changing business logic**. Each item lists file path, problem, and a drop-in fix.

Severity legend: `[CRIT]` exploit/loss now · `[HIGH]` serious risk or wallet pain · `[MED]` should fix · `[LOW]` polish.

---

## 1. Security Vulnerabilities

### 1.1 Hardcoded JWT secret fallback — `[CRIT]`
- File: [backend/application/usecase/auth_service.go:382-387](backend/application/usecase/auth_service.go#L382-L387), [backend/webapi/middleware/auth_middleware.go:118-123](backend/webapi/middleware/auth_middleware.go#L118-L123)
- Problem: `jwtSecret()` falls back to literal `"sunibgram-secret"` when `JWT_SECRET` env is unset. If deploy misses the env var, every attacker can forge admin tokens.
- Fix: fail-closed at boot. In `main.go` panic if `os.Getenv("JWT_SECRET") == ""`. Remove the string literal fallback.

### 1.2 SMTP TLS verification disabled — `[CRIT]`
- File: [backend/infrastructure/email/smtp.go:44](backend/infrastructure/email/smtp.go#L44)
- Problem: `InsecureSkipVerify: true` on `tls.Config`. SMTP traffic (with OTP / password-reset codes) is MITM-able.
- Fix: drop `InsecureSkipVerify` (default `false`). If self-signed cert is unavoidable for local dev, gate on `os.Getenv("APP_ENV") == "local"`.

### 1.3 Hardcoded admin allowlist — `[HIGH]`
- File: [backend/application/usecase/auth_service.go:268-275](backend/application/usecase/auth_service.go#L268-L275)
- Problem: `LoginProd` hardcodes `devin@`, `octa@`, `chermiese@` to bypass admin role check. Anyone who leaks/forks the repo learns who can log in.
- Fix: move list to env `ADMIN_ALLOWLIST="a@b,c@d"`, parse once at startup. Or attach the `admin` role to those users in seed and delete the bypass.

### 1.4 OTP generated with `math/rand` — `[HIGH]`
- File: [backend/application/usecase/verification_service.go:6,68](backend/application/usecase/verification_service.go#L68)
- Problem: `rand.Intn(1000000)` is non-cryptographic. Seed observable → OTP predictable.
- Fix: `crypto/rand` — `n, _ := rand.Int(rand.Reader, big.NewInt(1_000_000)); plain := fmt.Sprintf("%06d", n.Int64())`.

### 1.5 Rate limiter trusts raw `c.IP()` — `[HIGH]`
- File: [backend/webapi/router/auth_routes.go:15](backend/webapi/router/auth_routes.go#L15), [backend/cmd/main.go:185-193](backend/cmd/main.go#L185-L193)
- Problem: Fiber created without `EnableTrustedProxyCheck` / `TrustedProxies`. Attacker sets `X-Forwarded-For` → bypasses OTP send limit (3/10min).
- Fix:
  ```go
  app := fiber.New(fiber.Config{
      BodyLimit:               50 * 1024 * 1024,
      EnableTrustedProxyCheck: true,
      TrustedProxies:          []string{"127.0.0.1", "10.0.0.0/8"}, // your edge ranges
      ProxyHeader:             fiber.HeaderXForwardedFor,
  })
  ```

### 1.6 In-process rate limiter — `[HIGH]`
- File: [backend/infrastructure/ratelimit/rate_limiter.go](backend/infrastructure/ratelimit/rate_limiter.go)
- Problem: state in a `map[string]*bucket`. Two replicas = 2× limit. Also no eviction → memory leak (one entry per unique IP forever).
- Fix: back with Redis using `INCR key + EXPIRE` Lua atomic. If single-replica for now, add a goroutine that sweeps expired buckets every minute.

### 1.7 No throttle on `POST /api/auth/login/prod` & `LoginWithOTP` per IP — `[HIGH]`
- File: [backend/webapi/router/auth_routes.go:23-25](backend/webapi/router/auth_routes.go#L23-L25)
- Problem: Only `otpSendRateLimit` is wired. `/login` and `/login/prod` accept unlimited tries → credential stuffing.
- Fix: wrap with the same rate limiter keyed by IP+email, e.g. `loginLimiter := ratelimit.New(10, 5*time.Minute)`.

### 1.8 Weak password policy — `[MED]`
- File: [backend/application/usecase/auth_service.go:136-138,161-163](backend/application/usecase/auth_service.go#L136-L138)
- Problem: only `len(newPassword) < 6` check.
- Fix: enforce ≥10 chars + mixed-class. Move into `domain/model/user.go` `ValidatePassword()` so register + change + forgot all share it.

### 1.9 WebSocket token in URL query string — `[MED]`
- File: [backend/webapi/socket/socket.go:54-66](backend/webapi/socket/socket.go#L54-L66)
- Problem: `?token=<jwt>` ends up in reverse-proxy access logs and browser history.
- Fix: read JWT from `Sec-WebSocket-Protocol` subprotocol header or from a short-lived ticket exchanged via authenticated POST `/ws/ticket`.

### 1.10 CORS hardcoded to `http://localhost:3000` — `[MED]`
- File: [backend/cmd/main.go:188-193](backend/cmd/main.go#L188-L193)
- Problem: `AllowOrigins` is a string literal. Will silently break in staging/prod; deployer may relax it to `*` which combined with `AllowCredentials: true` is an explicit CORS spec violation but is sometimes accepted by browsers older than expected.
- Fix: read from `CORS_ALLOW_ORIGINS` env; refuse to start if empty.

### 1.11 No security headers middleware — `[MED]`
- File: [backend/cmd/main.go](backend/cmd/main.go)
- Problem: missing HSTS, X-Content-Type-Options, X-Frame-Options, Referrer-Policy.
- Fix: `app.Use(helmet.New())` (`github.com/gofiber/helmet/v2`).

### 1.12 No request timeouts — `[MED]`
- File: [backend/cmd/main.go:185-187](backend/cmd/main.go#L185-L187)
- Problem: no `ReadTimeout` / `WriteTimeout` / `IdleTimeout`. Slowloris exhausts goroutines.
- Fix: `ReadTimeout: 15*time.Second, WriteTimeout: 30*time.Second, IdleTimeout: 60*time.Second`.

### 1.13 Body limit 50 MB applied globally — `[MED]`
- File: [backend/cmd/main.go:186](backend/cmd/main.go#L186)
- Problem: chat send / thread post can be forced to allocate 50 MB. Image cap (`MaxImageUploadSize` 5 MB) only checked *after* full read.
- Fix: keep global at 25 MB; configure route-specific override only on doc/chat-media routes.

### 1.14 Session refresh = unlimited sliding window — `[MED]`
- File: [backend/application/usecase/auth_service.go:293-315](backend/application/usecase/auth_service.go#L293-L315)
- Problem: `RefreshToken` just extends the same `sid` TTL. Stolen sid → permanent access. No absolute lifetime.
- Fix: store `created_at` in session value; reject refresh if `now - created_at > 30 days`. Rotate `sid` on each refresh (delete old, mint new).

### 1.15 OTP rows never purged after use — `[LOW]`
- File: [backend/application/usecase/verification_service.go:119](backend/application/usecase/verification_service.go#L119), [backend/domain/model/verification_code.go](backend/domain/model/verification_code.go)
- Problem: `MarkUsed` flips a flag; row stays forever. Table grows unbounded.
- Fix: add scheduler cron `DELETE FROM verification_codes WHERE expires_at < NOW() - INTERVAL '7 days'`.

### 1.16 Error string leak — `[LOW]`
- File: throughout `webapi/middleware/auth_middleware.go:36`, `webapi/controller/*`
- Problem: `return c.Status(...).JSON(fiber.Map{"message": err.Error()})` passes raw internal errors (e.g. wrapped GORM/Redis errors) to clients.
- Fix: middleware/util that whitelists known message strings; everything else becomes `"internal error"` and is logged server-side.

---

## 2. High-Cost API / Performance Issues

### 2.1 230× `DeleteByPattern` calls — `[CRIT cost]`
- Files: [backend/application/usecase/user_profile_service.go:153-256](backend/application/usecase/user_profile_service.go#L153-L256), `follow_service.go`, `report_service.go`, others (230 hits total).
- Problem: each call runs Redis `SCAN cursor MATCH pattern COUNT 100` until cursor=0. SCAN cost is O(keyspace). A single avatar update fires ~15 scans → 15 full keyspace sweeps. Under load this is the single biggest CPU cost on Redis.
- Fix:
  1. Maintain **reverse-index sets** per entity: `idx:user:<id>:cache_keys` → SADD every concrete key on write; on invalidate, SMEMBERS + DEL pipeline. Constant-time replacement for SCAN.
  2. Pipeline the deletes: wrap N `DeleteByPattern` calls in `client.Pipeline()` so they share one network round trip.
  3. Add `COUNT 1000` to SCAN to cut iterations 10×.

### 2.2 DB connection pool unbounded — `[HIGH cost]`
- File: [backend/infrastructure/config/db.go](backend/infrastructure/config/db.go)
- Problem: GORM opened without `SetMaxOpenConns` / `SetMaxIdleConns` / `SetConnMaxLifetime`. Burst traffic = exhausts Postgres `max_connections`, DB rejects new connections → cascading 500s.
- Fix after `gorm.Open`:
  ```go
  sqlDB, _ := db.DB()
  sqlDB.SetMaxOpenConns(50)
  sqlDB.SetMaxIdleConns(25)
  sqlDB.SetConnMaxLifetime(30 * time.Minute)
  sqlDB.SetConnMaxIdleTime(5 * time.Minute)
  ```

### 2.3 New DB connection inside `Register` — `[HIGH cost]`
- File: [backend/application/usecase/auth_service.go:123](backend/application/usecase/auth_service.go#L123)
- Problem: `config.NewPostgresWrite().Create(...)` opens a brand-new GORM/pgx pool **per registration request**. Connection storm + leaks.
- Fix: inject `writeDB *gorm.DB` into `AuthService` constructor (already wired in `main.go`); use it instead of `NewPostgresWrite()`.

### 2.4 N+1 from nested Preload chain — `[HIGH cost]`
- File: [backend/infrastructure/repository/thread_repository.go:32-42](backend/infrastructure/repository/thread_repository.go#L32-L42)
- Problem: `preloadThreadDetails` runs 8 Preloads, several 3-4 levels deep (`OriginalThread.User.UserFaculty.Faculty`). For a page of 20 threads each repost triggers extra `IN (...)` queries. Combined with `applyThreadVisibility` joins = ~9 queries/page.
- Fix: project the join with a single `SELECT ... JOIN users JOIN user_faculties JOIN faculties` returning a flat row struct; reserve Preload only for genuine 1:N collections (`ThreadEvents`).

### 2.5 Recursive CTE has no depth bound — `[HIGH cost]`
- File: [backend/infrastructure/repository/thread_repository_query.go:100-114](backend/infrastructure/repository/thread_repository_query.go#L100-L114)
- Problem: `findAncestors` walks parent pointers with no `WHERE depth < N`. A deep/circular reply chain stalls Postgres.
- Fix: append `WHERE a.depth < 100` to the recursive arm and document the cap.

### 2.6 COUNT(*) on every paginated feed — `[HIGH cost]`
- File: [backend/infrastructure/repository/thread_repository_query.go:42](backend/infrastructure/repository/thread_repository_query.go#L42), [explore_repository.go:255](backend/infrastructure/repository/explore_repository.go#L255), many others.
- Problem: each list endpoint runs a duplicate `COUNT(*)` over the same joined visibility query. With multi-million-row threads table that's a full hashagg on every page request.
- Fix: keep the count behind a longer-TTL cache key separate from the rows (e.g. `count:thread:roots:viewer:%d` TTL 5 min). Or use Postgres `count_estimate(query)` for non-critical totals.

### 2.7 Per-key SET in tight loop — `[MED cost]`
- File: [backend/webapi/utils/feed_builder.go:117-127,228-237](backend/webapi/utils/feed_builder.go#L117-L127)
- Problem: avatar / profile-meta cache hydration calls `cacheRepo.Get` and `Cache.Set` once per user ID. A 20-user feed = 20 sequential Redis RTTs each.
- Fix: add `MGet([]string)` / `MSet(map)` helpers in `infrastructure/cache/cache.go` using `client.MGet` / `client.Pipeline`. Call once per batch.

### 2.8 `SAdd` + `Expire` is two round trips — `[MED cost]`
- File: [backend/infrastructure/cache/cache.go:98-118](backend/infrastructure/cache/cache.go#L98-L118)
- Problem: every membership write costs 2 RTTs.
- Fix: pipeline:
  ```go
  pipe := c.client.TxPipeline()
  pipe.SAdd(ctx, key, member)
  pipe.Expire(ctx, key, ttl)
  _, err := pipe.Exec(ctx)
  ```

### 2.9 60-second feed TTL — `[MED cost]`
- File: [backend/infrastructure/repository/thread_repository.go:26](backend/infrastructure/repository/thread_repository.go#L26)
- Problem: `threadFeedTTL = 60 * time.Second` forces full feed rebuild (with the 8-Preload chain) every minute per `(page, viewer)` combo. For a 10k-DAU app that's millions of cache misses/hour.
- Fix: bump to 5-10 minutes; rely on existing explicit invalidation (`invalidateVisibilityFeedsForViewer`) for freshness.

### 2.10 Feed cache stores fully hydrated threads — `[MED cost]`
- File: [backend/infrastructure/repository/thread_repository_query.go:58,174,250](backend/infrastructure/repository/thread_repository_query.go#L58)
- Problem: each cached entry is a JSON blob of preloaded threads + parents (large). Redis memory + JSON ser/deser dominate.
- Fix: cache only `[]thread_id` per query; hydrate from per-thread cache `thread:%d` (already exists). Cuts blob size ~10×.

### 2.11 `bcrypt.DefaultCost` for 6-digit OTPs — `[MED cost]`
- File: [backend/application/usecase/verification_service.go:69](backend/application/usecase/verification_service.go#L69), [backend/application/usecase/auth_service.go:221](backend/application/usecase/auth_service.go#L221)
- Problem: bcrypt at cost 10 takes ~70 ms per verify. OTP is 6 digits valid 10 min — bcrypt brings nothing over `hmac.Equal(hmac.Sum256(key, otp))` which is ~µs and constant-time.
- Fix: hash OTP with HMAC-SHA256 keyed by a server secret; compare with `hmac.Equal`. Keep bcrypt for actual user passwords.

### 2.12 `thread_id = ? OR root_id = ?` defeats index — `[MED cost]`
- File: [backend/infrastructure/repository/thread_repository_query.go:160,166](backend/infrastructure/repository/thread_repository_query.go#L160)
- Problem: `OR` on two columns prevents the planner from picking either index → sequential scan on big tables.
- Fix (no logic change): rewrite as `UNION ALL` of two equality queries, or add a partial index `CREATE INDEX idx_threads_root_or_self ON threads(root_id) WHERE root_id IS NOT NULL` + a separate single-row lookup for the anchor.

### 2.13 No HTTP compression — `[MED cost]`
- File: [backend/cmd/main.go](backend/cmd/main.go)
- Problem: JSON feeds shipped uncompressed; egress + latency 5-10× larger than needed.
- Fix: `app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))`.

### 2.14 Hourly subscription expiry inline goroutine — `[LOW cost]`
- File: [backend/cmd/main.go:239-247](backend/cmd/main.go#L239-L247)
- Problem: anonymous goroutine, no shutdown signal; ticker keeps DB connection on hourly tick. Fine, but if `ExpireSubscriptions` ever hangs you lose visibility.
- Fix: move into `infrastructure/scheduler/` alongside the others; add `context.Context` cancel on shutdown.

### 2.15 Email worker not scaled — `[LOW cost]`
- File: [backend/cmd/main.go:89](backend/cmd/main.go#L89)
- Problem: single `StartWorker(ctx)`. OTP send latency = SMTP RTT per email; queue piles up under bursts.
- Fix: start N workers (`for i := 0; i < 4; i++ { go emailService.StartWorker(ctx) }`).

### 2.16 SMTP `InsecureSkipVerify` already noted (see 1.2) plus no connection reuse — `[LOW cost]`
- File: [backend/infrastructure/email/smtp.go](backend/infrastructure/email/smtp.go)
- Problem: each send dials a fresh TCP+TLS handshake.
- Fix: maintain a pooled `*smtp.Client` keyed by host; reuse with `Reset`.

### 2.17 MinIO HTTP transport not tuned — `[LOW cost]`
- File: [backend/infrastructure/storage/minio_client.go:34-36](backend/infrastructure/storage/minio_client.go#L34-L36)
- Problem: custom transport sets `Proxy: nil` but leaves `MaxIdleConnsPerHost`, `IdleConnTimeout` at defaults (2). Many concurrent uploads = TCP churn.
- Fix: `MaxIdleConnsPerHost: 100, IdleConnTimeout: 90*time.Second, DisableCompression: true`.

---

## 3. Quick-Win Roll-Up

Sorted by effort vs. impact, highest payoff first.

| # | Item | Effort | Impact |
|---|---|---|---|
| 1.1 | Remove JWT secret fallback, fail-start | 5 min | CRIT |
| 1.2 | Drop `InsecureSkipVerify` on SMTP | 2 min | CRIT |
| 2.2 | Set DB pool limits | 10 min | HIGH cost |
| 2.3 | Inject `writeDB` into `AuthService`, drop `NewPostgresWrite()` in `Register` | 5 min | HIGH cost |
| 1.4 | OTP via `crypto/rand` | 5 min | HIGH |
| 1.5 | `EnableTrustedProxyCheck` + `TrustedProxies` | 5 min | HIGH |
| 1.7 | Rate-limit `/login` and `/login/prod` | 10 min | HIGH |
| 2.1 | Replace `DeleteByPattern` with reverse-index sets | 1-2 days | CRIT cost |
| 2.4 | Replace nested Preload chain with single projected join | 1 day | HIGH cost |
| 2.6 | Separate count cache | 2 hr | HIGH cost |
| 2.5 | Cap recursive CTE depth | 5 min | HIGH cost |
| 2.9 | Bump feed TTL to 5-10 min | 2 min | MED cost |
| 2.13 | Add `compress` middleware | 2 min | MED cost |
| 1.10 | CORS origins from env | 5 min | MED |
| 1.11 | Add `helmet` middleware | 5 min | MED |
| 1.12 | Add server timeouts | 5 min | MED |
| 2.7, 2.8 | Pipeline / MGET helpers in cache layer | 2 hr | MED cost |
| 1.3 | Move admin allowlist to env | 10 min | HIGH |
| 1.14 | Absolute session lifetime + sid rotation on refresh | 1 hr | MED |
| 1.15 | Cron purge old `verification_codes` | 15 min | LOW |

All fixes preserve current business logic — they tighten the same code paths.
