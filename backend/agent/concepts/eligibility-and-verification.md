---
summary: Who can be on Sunibgram — current Binus students with a verified @binus.ac.id email, one account each, verified before posting; plus OTP rules.
last_synced: 2026-06-03
source:
  - frontend/app/about/help/page.tsx
  - backend/domain/model/user/user.go
tags: [concept, eligibility, verification, auth]
---

# Eligibility & verification

The hard rule that gates the whole platform.

## Who can be here
- **Current Binus students only.** A verified `@binus.ac.id` email is required.
- **One account per student.** No duplicate accounts.
- Verification must be completed **before you can post or reply**.
- Accounts that become ineligible (e.g. no longer a Binus student) may be suspended or
  terminated. See [[rules]] and [[terms]].

## Sign-up & login
Create an account at `/signup` (account info → profile details → email verification); log
in at `/login` (credentials → emailed OTP). Reset at `/forgot-password`. See
[[public-routes]].

## OTP & passwords
OTP expiry, send/verify rate limits, login-OTP lockout, and the password complexity rule
are all in [[limits]]. If an OTP never arrives: check spam, then request a new code — see
[[faq]].

## Username vs display name
Display name and profile details are editable any time (`/settings/account`); the
`@handle` is stable to prevent identity confusion — see [[faq]].
