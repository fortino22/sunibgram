---
summary: Each account is tied to a Binus faculty (chosen at signup), used for the faculty feed and faculty leaderboards; the live list comes from /faculties.
last_synced: 2026-06-03
source:
  - backend/infrastructure/seeder/faculty_seeder.go
tags: [entity, faculties]
---

# Faculties

Each account is tied to a Binus faculty, chosen at signup. It drives the **Faculty feed**
(see [[threads]]) and the per-faculty [[leaderboard]] boards.

## Seeded faculties
BINUS ASO School of Engineering · BINUS Business School · BINUS Graduate Program ·
BINUS Online · Doctor of Computer Science · Faculty of Digital Communication and Hotel
and Tourism · Faculty of Engineering · Faculty of Humanities · School of Accounting ·
School of Computer Science · School of Computing and Creative Arts · School of Design ·
School of Information Systems.

> This is the seeded default list. The live list comes from the `/faculties` endpoint and
> may change — don't promise a faculty exists; let the user pick from the app.
