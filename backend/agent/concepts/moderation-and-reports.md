---
summary: How reporting and enforcement work — report categories, trend-report reasons, the moderation queue, and the warning→termination ladder.
last_synced: 2026-06-03
source:
  - backend/infrastructure/seeder/violation_seeder.go
  - backend/infrastructure/seeder/trend_report_reason_seeder.go
tags: [concept, moderation, reports, safety]
---

# Moderation & reports

How content gets flagged and acted on. Acceptable-use rules live in [[rules]].

## How to report
Three-dot menu on any [[threads|thread]], message, or profile → **Report** → choose a
category → add a short explanation. Reports go to the moderation queue, usually reviewed
within a couple of working days.

## Report categories
| Category | What it covers |
|---|---|
| **Spam** | Repetitive, misleading, or unwanted promotional content. |
| **Harassment** | Bullying, intimidation, threats, or targeted abusive behavior. |
| **Hate Speech** | Content attacking people based on protected characteristics. |
| **Misinformation** | False or misleading information that may cause harm. |
| **Nudity/Sexual Content** | Explicit sexual content, nudity, or sexually suggestive material. |
| **Violence** | Graphic violence, threats, or content encouraging physical harm. |
| **Copyright Infringement** | Content that appears to violate intellectual property rights. |
| **Other** | Other policy violations not covered by the listed categories. |

## Reporting a trending topic
Trends can be reported separately. Reasons: **The associated content is not relevant**
(`not_relevant`) · **This trend is spam** (`spam`) · **This trend is abusive or harmful**
(`abusive_or_harmful`) · **Not interested in this** (`not_interested`) · **This trend is
a duplicate** (`duplicate`) · **This trend is harmful or spammy** (`harmful_or_spammy`).

## Enforcement ladder
Depending on severity and history: warnings → content removal → feature restrictions →
temporary suspension → permanent account termination.

## Community-level moderation
[[communities|Community]] owners and the admins they appoint enforce rules within their
own rooms. Sunibgram steps in only if a community is used to break platform rules. See
[[terms]].
