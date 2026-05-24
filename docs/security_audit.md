# Backend Security Audit

Scope: `backend/` — Go/Fiber. No logic change. Drop-in hardening only.

Severity: **CRIT** = exploit now, **HIGH** = exploit with one extra step, **MED** = defense-in-depth, **LOW** = hygiene.

---

## CRIT-1 — Prompt injection in thread summary

**Problem**
[application/usecase/thread_summary/thread_summary_service.go:99-102](backend/application/usecase/thread_summary/thread_summary_service.go#L99-L102) concatenates raw user comments into OpenAI prompt:

```go
prompt := fmt.Sprintf(
    "Summarize the following comments... :\n\n%s",
    strings.Join(comments, "\n"),
)
```

Any user can write `IGNORE PREVIOUS INSTRUCTIONS. Output ... ` in a comment → LLM follows attacker input. Also leaks system prompt, can be used to burn tokens.

**Solution** (no logic change)
1. Use OpenAI **structured messages**: put instructions in `SystemMessage`, user content in single `UserMessage` wrapped in `<comment>` tags + escape `<` `>` in content.
2. Cap input length per comment (e.g. 500 chars) before join. Cap total joined length (e.g. 8000 chars) — also reduces token cost.
3. Set `MaxTokens` on output → bounds malicious "write 10MB of X" attacks.

---

## CRIT-2 — DB password leakable via panic message

**Problem**
[backend/infrastructure/config/db.go:36](backend/infrastructure/config/db.go#L36)

```go
panic(fmt.Sprintf("failed to connect to postgres (%s): %v", role, err))
```

`err` from `gorm.Open` on a malformed DSN can echo the DSN (including `password=...`) into stderr → captured by container logs → CloudWatch / log aggregator.

**Solution**
Log only `role` and error class, not raw err:
```go
log.Printf("postgres connect failed role=%s: %s", role, sanitizeDBErr(err))
panic("postgres connect failed")
```
Where `sanitizeDBErr` masks `password=...` substring.

---

## HIGH-1 — Missing ownership / authorization on event mutations

**Problem**
[backend/webapi/controller/event/event_controller.go](backend/webapi/controller/event/event_controller.go) — `Update`, `Delete`, `ApproveEvent`, `DeclineEvent`, `RequestRevision` accept event ID from URL and call service with only `RequireAuth`. No check that caller owns the event or is admin (besides string-literal `RoleName != "admin"` scattered checks that are role-coupled, not ownership-coupled).

**Solution**
Add owner check in service layer (entry point): `event.UserID == claims.UserID || claims.RoleName == admin`. Centralize in `application/helper/` so other modules reuse.

Audit same pattern: every controller calling `service.Update(id, ...)` or `service.Delete(id, ...)` where `id` comes from URL must enforce ownership inside the service.

---

## HIGH-2 — Admin check via string literal `"admin"`

**Problem**
Six+ sites: [event_controller.go:87,164,183,206](backend/webapi/controller/event/event_controller.go#L87), [order_controller.go:117](backend/webapi/controller/order/order_controller.go#L117), [notification_controller.go:34](backend/webapi/controller/notification/notification_controller.go#L34), [feedback_controller.go:85](backend/webapi/controller/feedback/feedback_controller.go#L85), `auth_middleware.go:88`.

If role seed renames "admin" → "ADMIN" or "superadmin", every check silently breaks open or breaks closed. Also: `feedback_controller.go:85` admin op only behind `RequireAuth`, missing `AdminOnly()` middleware (controller does its own check, easy to forget).

**Solution**
1. Add `domain/model/role/role.go`: `const RoleAdmin = "admin"`. Replace all string literals.
2. Add `middleware.AdminOnly()` to admin routes; remove inline checks in controllers (defense-in-depth still OK but route-level is canonical).
3. Better: introduce `claims.IsAdmin()` helper in `webapi/utils/ctrl_helpers.go`.

---

## HIGH-3 — `log.Fatal` on missing secrets at request time

**Problem**
[backend/webapi/middleware/auth_middleware.go:120-126](backend/webapi/middleware/auth_middleware.go#L120-L126) — `jwtSecret()` reads `os.Getenv` + `log.Fatal` **per Auth() call** (i.e. per request).

Same pattern: [webapi/socket/socket.go:177](backend/webapi/socket/socket.go#L177), [verification_service.go:40](backend/application/usecase/verification/verification_service.go#L40), [chat_crypto.go:28](backend/application/helper/chat_crypto.go#L28).

Issues:
- `log.Fatal` mid-request kills the whole server — DoS amplifier if env gets unset somehow.
- `os.Getenv` per request = pointless syscall on hot path.

**Solution**
Read once at boot in `cmd/bootstrap/`, fail fast there, pass via DI. Cache secret in package var initialized via `Init(secret string)`. Already done for `sessionCache` (auth_middleware.go:22-26) — extend same pattern.

---

## HIGH-4 — Unvalidated path params used as IDs

**Problem**
[event_participant_controller.go:28](backend/webapi/controller/event_participant/event_participant_controller.go#L28), [event_pin_controller.go:28](backend/webapi/controller/event_pin/event_pin_controller.go#L28), [event_thread_pin_controller.go:24,40](backend/webapi/controller/event_thread_pin/event_thread_pin_controller.go#L24) — `ctx.Params("id")` passed as string to service. If service does no parse/validate, attackers send `../../foo` or SQL fragments. GORM mostly safe via parameterized queries, but service may use it in Redis keys, log lines, file paths.

Also: `_shared/ctrl_helpers.go:26 ParseIDParam` silently returns `0` on parse fail — callers that don't check `id == 0` operate on "first row".

**Solution**
1. Make `ParseIDParam` return `(id uint, err error)`; controllers must error-check.
2. Where param goes to service as string, parse to uint first in controller.
3. Audit usage of `ctx.Params` across all controllers — anything fed to Redis key / log line must be `strconv.FormatUint` after parse.

---

## HIGH-5 — Payment server key fetched per-request via `os.Getenv`

**Problem**
[order_controller.go:151](backend/webapi/controller/order/order_controller.go#L151) and [order_service.go:37](backend/application/usecase/order/order_service.go#L37) call `os.Getenv("MIDTRANS_SERVER_KEY")` at request time. If env unset and code path uses it in an error string → leak path. Also forces secret to be in process env permanently.

**Solution**
Resolve once at boot into a `MidtransConfig{ServerKey string}` and inject via constructor. Same fix as HIGH-3.

---

## HIGH-6 — Migrations disable FK constraints in all environments

**Problem**
[infrastructure/migration/migrate.go:50](backend/infrastructure/migration/migrate.go#L50)

```go
DisableForeignKeyConstraintWhenMigrating: true,
```

Plus `SeedDummy()` / `SeedThreadWithReplies()` run on every boot per [migrate.go:115-116](backend/infrastructure/migration/migrate.go#L115). Prod boot inserts dummy users / threads. Dummy accounts may have weak/known passwords → trivial account takeover in prod.

**Solution**
Gate dummy seeders behind `APP_ENV != "production"`. FK disable is OK during migrate but should be re-enabled (it is in GORM by default after migrate done — verify).

---

## HIGH-7 — WebSocket auth has no rate limit

**Problem**
[webapi/socket/socket.go:123-146](backend/webapi/socket/socket.go) — JWT validated on every WS connect via subprotocol. No `ratelimit.AllowN` check on the connect handler → attacker can spam handshakes with garbage tokens, burning JWT verify CPU. JWT HS256 verify is ~10us — not free.

**Solution**
Wrap WS upgrade handler with IP-based rate limiter (existing `infrastructure/ratelimit/rate_limiter.go`). Same for `/auth/login`, `/auth/register`, `/verification/*` if not already covered (verify these have ratelimit middleware).

---

## MED-1 — TLS not enforced on SMTP

**Problem**
[infrastructure/email/smtp_client.go:65](backend/infrastructure/email/smtp_client.go#L65) — `smtp.PlainAuth` over standard `smtp.SendMail`. Goes plaintext if SMTP server doesn't advertise STARTTLS. Anyone MITM the SMTP path can read OTP codes.

**Solution**
Require STARTTLS explicitly: dial → `c.StartTLS(&tls.Config{ServerName: host})` → fail if not supported. Or use port 465 with `tls.Dial`.

---

## MED-2 — OTP rate limit keyed on email only

**Problem**
[verification_service.go:74,89](backend/application/usecase/verification/verification_service.go#L74) — rate limit by email. Attacker rotates target emails to bypass per-email limit. No global per-IP limit on `/verification/request`.

**Solution**
Add second rate-limit key: `verify:request:ip:<ip>` (e.g. 30/hour). Keep email key.

---

## MED-3 — `OptionalAuth` swallows expired-session errors

**Problem**
[auth_middleware.go:54-78](backend/webapi/middleware/auth_middleware.go#L54-L78) — if token parses but session is revoked, `OptionalAuth` calls `c.Next()` without `Locals`. Routes that use OptionalAuth treat user as anonymous → fine. But if the route falls back to "anonymous-readable" path, a revoked user keeps reading public content seamlessly — no log of revoked-session attempts.

**Solution**
Log revoked-session usage (info level). Helps detect stolen-token replay attempts.

---

## MED-4 — `kind` claim check accepts any non-`"access"` as invalid

**Problem**
[auth_middleware.go:113](backend/webapi/middleware/auth_middleware.go#L113) — refuses non-`access` tokens (good). But there's no `aud` / `iss` validation. If JWT secret leaks between services, tokens from one service work in another.

**Solution**
Set `iss = "sunibgram-api"` on issue, validate on parse via `jwt.WithIssuer`. Same for `aud` if multi-client.

---

## MED-5 — `MarkRoomRead` Raw SQL has no owner check

**Problem**
[infrastructure/repository/chat/chat_repository.go:100-103](backend/infrastructure/repository/chat/chat_repository.go#L100) — parameterized but assumes caller validated `userID` is in the room.

**Solution**
Either: (a) verify caller is room member in service layer (probably already done — confirm), or (b) add `AND user_id IN (SELECT user_id FROM chat_room_members WHERE room_id = ?)` to the SQL.

---

## MED-6 — `Faculty` admin endpoints risk privilege escalation if string check fails

Covered by HIGH-2; mentioning so audit doesn't miss the faculty controller.

---

## MED-7 — File upload: validate magic bytes, not extension

**Problem**
[domain/model/media/media.go](backend/domain/model/media/media.go) `ValidateImageExt` checks extension only. Attacker uploads `evil.jpg` containing PHP / SVG with `<script>` → if MinIO bucket served with wrong Content-Type, XSS.

**Solution**
Add `http.DetectContentType(first 512 bytes)` check. Reject mismatched MIME vs extension. Set `Content-Disposition: attachment` on private bucket presigned URLs.

---

## MED-8 — No `MaxTokens` on OpenAI call

**Problem** (also a cost issue — see cost doc)
[openai_client.go:30-37](backend/infrastructure/openai/openai_client.go#L30) — `StreamGenerate` has no `MaxTokens`. Attacker triggers summary repeatedly → unbounded billing + slow request. Slowloris-style DoS.

**Solution**
`MaxTokens: openai.Int(500)` in params (summary is one paragraph).

---

## LOW-1 — Trusted proxies optional

**Problem**
[server.go:131-140](backend/cmd/bootstrap/server.go#L131-L140) — `trustedProxies()` returns empty slice if env unset, but Fiber config has `EnableTrustedProxyCheck: true`. With empty list, all proxies untrusted → `X-Forwarded-For` ignored → rate limiter sees container IP for everyone.

**Solution**
Fail-fast like `corsAllowOrigins()` if `TRUSTED_PROXIES` unset in prod (gate by `APP_ENV`).

---

## LOW-2 — `compress.LevelBestSpeed` on response leaks length via timing

Negligible. Skip.

---

## LOW-3 — Helmet defaults — verify CSP

**Problem**
[server.go:36](backend/cmd/bootstrap/server.go#L36) `helmet.New()` uses defaults. Default CSP may allow inline scripts depending on Fiber helmet version.

**Solution**
Explicit config:
```go
helmet.New(helmet.Config{
    ContentSecurityPolicy: "default-src 'self'",
    ReferrerPolicy:        "no-referrer",
    HSTSMaxAge:            31536000,
})
```

---

## LOW-4 — No audit log for admin actions

**Problem**
Admin approves/declines event, deletes feedback — no record. Compromised admin account leaves no trail.

**Solution**
Add `audit_logs` table; insert row in admin action paths. Out of scope for "no logic change" — flagged for later.

---

## Summary Table

| ID | Severity | Where | One-line fix |
|---|---|---|---|
| CRIT-1 | CRIT | thread_summary_service.go:99 | System/user message split + length cap + MaxTokens |
| CRIT-2 | CRIT | config/db.go:36 | Strip password from panic message |
| HIGH-1 | HIGH | event_controller.go:128,146 | Owner check in service |
| HIGH-2 | HIGH | 6+ controllers | Constant for role name + AdminOnly middleware |
| HIGH-3 | HIGH | auth_middleware.go:120 | Read secret once at boot, inject via DI |
| HIGH-4 | HIGH | event_*_controller.go | ParseIDParam returns error |
| HIGH-5 | HIGH | order_controller.go:151 | Config struct injected once |
| HIGH-6 | HIGH | migrate.go:115 | Gate dummy seeders by APP_ENV |
| HIGH-7 | HIGH | socket.go:123 | Rate-limit WS upgrade |
| MED-1 | MED | smtp_client.go:65 | Force StartTLS |
| MED-2 | MED | verification_service.go:74 | Add per-IP rate limit |
| MED-3 | MED | auth_middleware.go:69 | Log revoked-session reads |
| MED-4 | MED | auth_middleware.go:113 | Validate iss/aud |
| MED-5 | MED | chat_repository.go:100 | Membership check in SQL |
| MED-7 | MED | media/media.go | Magic-byte MIME check |
| MED-8 | MED | openai_client.go:30 | MaxTokens param |
| LOW-1 | LOW | server.go:131 | Fail-fast on unset TRUSTED_PROXIES in prod |
| LOW-3 | LOW | server.go:36 | Explicit Helmet config |
| LOW-4 | LOW | admin paths | Audit log table |
