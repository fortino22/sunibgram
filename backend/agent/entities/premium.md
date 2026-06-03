---
summary: Premium unlocks creating communities and events. Paid purchase is coming soon; for now users can activate a one-time 1-month free trial.
last_synced: 2026-06-03
source:
  - backend/application/usecase/premium_subscription/premium_subscription_service.go
tags: [entity, premium, billing]
---

# Premium

The upgrade that unlocks creating communities and events. Manage it at
`/settings/subscription` (see [[user-routes]]).

## Buying premium
**Coming soon.** Paid premium purchase is not available yet — do not tell users they can
buy it or quote a price. If asked to buy, say paid premium is coming soon and point them to
the free trial below.

## Free trial
**1 month**, activatable **once per account** — a second attempt is rejected. This is the way
to get premium right now.

## What it unlocks
- Creating a [[communities|community]].
- Creating an [[events|event]].

When a user asks to do a premium-gated action, mention the premium requirement and suggest the
**1-month free trial** (purchase is coming soon). Exact figures are mirrored in [[limits]].
