---
summary: One-off Rp 10,000 upgrade (paid via Midtrans) that unlocks creating communities and events; a one-time 1-month free trial is available.
last_synced: 2026-06-03
source:
  - backend/application/usecase/order/order_service.go
  - backend/application/usecase/premium_subscription/premium_subscription_service.go
tags: [entity, premium, billing]
---

# Premium

The paid upgrade. Manage it at `/settings/subscription` (see [[user-routes]]).

## Price & payment
**Rp 10,000** (IDR), a one-off order paid via **Midtrans** (item "Sunibgram Premium").

## Free trial
**1 month**, activatable **once per account** — a second attempt is rejected.

## What it unlocks
- Creating a [[communities|community]].
- Creating an [[events|event]].

When a user asks to do a premium-gated action, mention the premium requirement rather
than promising it works for everyone. Exact figures are mirrored in [[limits]].
