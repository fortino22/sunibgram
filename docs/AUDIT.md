# Sunibgram Backend — Security & DB Cost Audit

> **Scope:** Security vulnerabilities + high-cost DB queries.  
> **Constraint:** All fixes preserve existing code logic and API responses.  
> **Fix types:** (A) migrate.go SQL additions · (B) env/config · (C) minimal route wiring

---

## Part 1 — Security Vulnerabilities

---

### SEC-01 · CRITICAL — JWT Hardcoded Fallback Secret

**File:** [`webapi/middleware/auth_middleware.go:118–123`](webapi/middleware/auth_middleware.go#L118)  
**Also:** [`application/usecase/auth_service.go:382–387`](application/usecase/auth_service.go#L382)

```go
func jwtSecret() string {
    if s := os.Getenv("JWT_SECRET"); s != "" {
        return s
    }
    return "sunibgram-secret"  // ← known fallback
}
```

**Risk:** If `JWT_SECRET` is unset in any environment, every user can self-sign valid JWTs using the public fallback `"sunibgram-secret"`. Full authentication bypass.

**Fix (B) — env enforcement at startup:**  
Add to `cmd/main.go` before `fiber.New()`:

```go
if os.Getenv("JWT_SECRET") == "" {
    log.Fatal("JWT_SECRET env var must be set")
}
```

No logic change. App simply refuses to start without a real secret.

---

### SEC-02 · HIGH — Hardcoded Admin Email Backdoor

**File:** [`application/usecase/auth_service.go:268–275`](application/usecase/auth_service.go#L268)

```go
allowedEmails := map[string]bool{
    "devin@binus.ac.id":     true,
    "octa@binus.ac.id":      true,
    "chermiese@binus.ac.id": true,
}
if user.Role.Name != "admin" && !allowedEmails[user.Email] {
    return nil, model.TokenPair{}, errors.New("admin access required")
}
```

**Risk:** Three email addresses bypass the role check and can always log in as admin regardless of their role in the database. Hardcoded secrets in source code are exposed in git history forever.

**Fix (B) — move to env var:**  
Set `ADMIN_WHITELIST_EMAILS=devin@binus.ac.id,octa@binus.ac.id,chermiese@binus.ac.id` in `.env`.  
Load with `strings.Split(os.Getenv("ADMIN_WHITELIST_EMAILS"), ",")`.  
Long-term: give these users the `admin` role in DB and remove the whitelist entirely.

---

### SEC-03 · HIGH — No Rate Limit on `/forgot-password`

**File:** [`webapi/router/auth_routes.go:29`](webapi/router/auth_routes.go#L29)

```go
auth.Post("/forgot-password", ctrl.ForgotPassword)
```

**Risk:** No rate limiter on this route. Attacker can:
1. Enumerate valid `@binus.ac.id` emails from timing differences.
2. Spam OTP requests to exhaust SMTP quota.
3. Brute-force OTP codes before expiry (bcrypt slows comparison but doesn't stop automation at scale).

The rate limiter infrastructure already exists at [`infrastructure/ratelimit/rate_limiter.go`](infrastructure/ratelimit/rate_limiter.go). The OTP-fail limiter is already wired for `/login-otp`. Same pattern applies here.

**Fix (C) — wire existing limiter in route:**  
Apply `forgotPasswordLimiter` (e.g. 5 attempts / 15 min per IP) in `auth_routes.go` using the same `middleware.RateLimitMiddleware` pattern already used for OTP routes. Zero API result change.

---

### SEC-04 · MEDIUM — CORS Hardcoded to `localhost:3000`

**File:** [`cmd/main.go:166`](cmd/main.go#L166)

```go
AllowOrigins: "http://localhost:3000",
```

**Risk:** Production and staging deploys will reject all browser requests with CORS errors unless this is changed per-environment. Developers may be tempted to open `*` as a quick fix, which is worse.

**Fix (B) — read from env:**

```go
allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
if allowedOrigins == "" {
    allowedOrigins = "http://localhost:3000"
}
// AllowOrigins: allowedOrigins
```

Add `CORS_ALLOWED_ORIGINS=https://sunibgram.com` to `.env.docker` and production env. No API change.

---

## Part 2 — High-Cost Database Queries

All fixes in this section are **pure SQL additions to [`infrastructure/migration/migrate.go`](infrastructure/migration/migrate.go)** — the `AutoMigrate` function. No query logic changes, no API result changes.

---

### DB-01 · CRITICAL — Missing Index on `community_members(user_id)`

**Used in:** [`infrastructure/repository/thread_repository_query.go:279, 300, 369, 387`](infrastructure/repository/thread_repository_query.go#L279)

```go
Where("communities.visibility = ? OR communities.community_id IN (SELECT community_id FROM community_members WHERE user_id = ?)", ...)
```

This subquery appears in 4 feed queries: `FindAllCommunityThreads`, `findAllCommunityThreadsByQuery`, and two variants. On every cache miss, each query runs this subquery against an unindexed `user_id` column — a full scan of `community_members`.

**Existing indexes:** Only `idx_communities_category_id` exists on communities. No index on `community_members.user_id`.

**Fix:**

```go
db.Exec("CREATE INDEX IF NOT EXISTS idx_community_members_user_id ON community_members(user_id)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_community_members_community_user ON community_members(community_id, user_id)")
```

---

### DB-02 · HIGH — Missing Index on `threads(parent_id)` and `threads(root_id)`

**Used in:**
- [`thread_repository_query.go:80`](infrastructure/repository/thread_repository_query.go#L80) — `WHERE threads.parent_id = ?` (FindChain, children lookup)
- [`thread_repository_query.go:107–108`](infrastructure/repository/thread_repository_query.go#L107) — `WHERE threads.thread_id = ? OR threads.root_id = ?` (FindByRootID, full thread chain)

Both queries run without caching (`FindChain` has no Redis cache layer). Full table scan on threads table for every thread detail page load.

**Fix:**

```go
db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_parent_id ON threads(parent_id)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_root_id ON threads(root_id)")
```

---

### DB-03 · HIGH — Missing Index on `user_relationships(following_id)`

**Used in:** [`thread_repository_user_feed.go:165, 178`](infrastructure/repository/thread_repository_user_feed.go#L165)

```go
Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id")
```

`FindFollowingFeed` JOINs `user_relationships` on `following_id`. The existing index `idx_user_rel_follower` only covers `follower_id`. The `following_id` join condition forces a sequential scan on `user_relationships` for every following feed request on cache miss.

**Fix:**

```go
db.Exec("CREATE INDEX IF NOT EXISTS idx_user_rel_following ON user_relationships(following_id)")
```

---

### DB-04 · HIGH — Missing Index on `threads(original_thread_id)`

**Used in:** [`thread_repository_query.go:272, 293`](infrastructure/repository/thread_repository_query.go#L272) (FindAllCommunityThreads, findAllCommunityThreadsByQuery)

```go
Joins("LEFT JOIN threads original ON original.thread_id = threads.original_thread_id")
```

The self-join to resolve repost community IDs (`COALESCE(threads.community_id, original.community_id)`) hits `original_thread_id` with no index. This join runs for both the COUNT and SELECT queries on every community feed cache miss.

**Fix:**

```go
db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_original_thread_id ON threads(original_thread_id)")
```

---

### DB-05 · HIGH — ILIKE Searches Without Trigram Indexes (Full Table Scans)

**Used in multiple places:**

| File | Line | Columns searched with ILIKE |
|---|---|---|
| [`explore_repository.go:302`](infrastructure/repository/explore_repository.go#L302) | threads.content | `content ILIKE '%query%'` |
| [`explore_repository.go:376`](infrastructure/repository/explore_repository.go#L376) | users.user_name, users.name, user_profiles.bio | 3 ILIKE conditions |
| [`explore_repository.go:139`](infrastructure/repository/explore_repository.go#L139) | communities.name, communities.description, categories.name | 3 ILIKE conditions |
| [`explore_repository.go:409`](infrastructure/repository/explore_repository.go#L409) | threads.content + DISTINCT ON | ILIKE + sequential sort |
| [`thread_repository_query.go:368, 386`](infrastructure/repository/thread_repository_query.go#L368) | communities.name, threads.content | 2 ILIKE conditions |

All use a `%query%` leading-wildcard pattern. PostgreSQL cannot use a B-tree index for leading wildcards — every ILIKE triggers a full sequential scan on the table. On a table with 100k+ threads this is catastrophic.

**Fix — enable `pg_trgm` and add GIN indexes:**

```go
// Enable trigram extension (requires superuser on first run)
db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm")

// Threads content search
db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_content_trgm ON threads USING gin(content gin_trgm_ops)")

// User search
db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING gin(user_name gin_trgm_ops)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_users_name_trgm ON users USING gin(name gin_trgm_ops)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_user_profiles_bio_trgm ON user_profiles USING gin(bio gin_trgm_ops)")

// Community search
db.Exec("CREATE INDEX IF NOT EXISTS idx_communities_name_trgm ON communities USING gin(name gin_trgm_ops)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_communities_desc_trgm ON communities USING gin(description gin_trgm_ops)")
```

GIN trigram indexes support `ILIKE '%query%'` and reduce search from O(n) to O(log n). PostgreSQL automatically uses them for `ILIKE` once `pg_trgm` is active — **zero query code changes needed**.

> **Note:** `pg_trgm` is included in standard PostgreSQL distributions. For the Docker Postgres 16 image it is available via `CREATE EXTENSION` with no additional packages.

---

### DB-06 · MEDIUM — `FindSuggestedPeople` — Triple Self-Join with Aggregation

**File:** [`infrastructure/repository/explore_repository.go:182–205`](infrastructure/repository/explore_repository.go#L182)

```sql
SELECT ... FROM user_relationships AS followed
JOIN user_relationships AS suggested ON suggested.follower_id = followed.following_id
JOIN users ON users.user_id = suggested.following_id
LEFT JOIN user_relationships AS existing ON existing.follower_id = ? AND existing.following_id = suggested.following_id
WHERE followed.follower_id = ? AND suggested.following_id <> ?
GROUP BY ... ORDER BY COUNT(DISTINCT followed.following_id) DESC
```

Three passes over `user_relationships` plus a `GROUP BY` + `COUNT(DISTINCT)` aggregate. TTL is 60 seconds per user — on a busy platform, this runs once per minute per active user.

**Fix — add composite covering index:**

```go
db.Exec("CREATE INDEX IF NOT EXISTS idx_user_rel_follower_following ON user_relationships(follower_id, following_id)")
```

This covering index lets all three self-join conditions resolve from the index without touching the heap.

---

### DB-07 · MEDIUM — `FindByRootID` Count Query Not Error-Checked

**File:** [`infrastructure/repository/thread_repository_query.go:106–110`](infrastructure/repository/thread_repository_query.go#L106)

```go
r.db.readDB.Model(&model.Thread{}).
    Joins("LEFT JOIN communities ON ...").
    Where("threads.thread_id = ? OR threads.root_id = ?", rootID, rootID).
    Count(&total)  // ← error ignored
```

The COUNT result is silently discarded on DB error. Users see `total: 0` on infrastructure failures instead of an error response. Not a crash, but produces misleading pagination metadata.

**Fix (B) — no code change, just document:** This is a silent error swallow. Track with an alert on `total == 0` when threads exist. The `idx_threads_root_id` from DB-02 also makes this query faster.

---

### DB-08 · LOW — `FindChain` Has No Cache

**File:** [`infrastructure/repository/thread_repository_query.go:67–87`](infrastructure/repository/thread_repository_query.go#L67)

`FindChain` runs 2 DB queries (anchor + children) on every thread detail page load with no Redis caching. With `idx_threads_parent_id` from DB-02, the children query becomes fast. The anchor query already benefits from `idx_threads_created_at`. No cache needed once indexes exist.

---

## Summary — All Fixes

### Security Fixes

| ID | Severity | File to Change | Change Type |
|---|---|---|---|
| SEC-01 | Critical | `cmd/main.go` | Add startup env check (1 line) |
| SEC-02 | High | `.env` / `cmd/main.go` | Move emails to env var |
| SEC-03 | High | `webapi/router/auth_routes.go` | Wire existing rate limiter |
| SEC-04 | Medium | `cmd/main.go` + `.env` files | Read CORS from env var |

### DB Index Additions (all in `infrastructure/migration/migrate.go`)

```go
// DB-01: community_members lookup (subquery used in 4 feed queries)
db.Exec("CREATE INDEX IF NOT EXISTS idx_community_members_user_id ON community_members(user_id)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_community_members_community_user ON community_members(community_id, user_id)")

// DB-02: thread chain/reply queries
db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_parent_id ON threads(parent_id)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_root_id ON threads(root_id)")

// DB-03: following feed JOIN
db.Exec("CREATE INDEX IF NOT EXISTS idx_user_rel_following ON user_relationships(following_id)")

// DB-04: community thread self-join
db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_original_thread_id ON threads(original_thread_id)")

// DB-05: ILIKE search — requires pg_trgm extension
db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm")
db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_content_trgm ON threads USING gin(content gin_trgm_ops)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING gin(user_name gin_trgm_ops)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_users_name_trgm ON users USING gin(name gin_trgm_ops)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_user_profiles_bio_trgm ON user_profiles USING gin(bio gin_trgm_ops)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_communities_name_trgm ON communities USING gin(name gin_trgm_ops)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_communities_desc_trgm ON communities USING gin(description gin_trgm_ops)")

// DB-06: suggested people aggregation
db.Exec("CREATE INDEX IF NOT EXISTS idx_user_rel_follower_following ON user_relationships(follower_id, following_id)")
```

**All `CREATE INDEX IF NOT EXISTS` statements are safe to add — they are idempotent and do not affect any query result or business logic.**

---

## Priority Order

1. **SEC-01** — JWT secret fallback. Trivial fix, critical blast radius.
2. **DB-05** — `pg_trgm` + GIN indexes. Explore/search is the most user-facing bottleneck and currently does full table scans.
3. **DB-01** — `community_members(user_id)` index. Hits 4 feed queries on every page load.
4. **SEC-02** — Remove hardcoded admin emails from source.
5. **DB-02 / DB-03 / DB-04** — Thread chain and following feed indexes.
6. **SEC-03** — Rate limit `/forgot-password`.
7. **DB-06** — Suggested people covering index.
8. **SEC-04** — CORS env config.
