# Public Routes (no login)

Routes reachable without signing in: authentication pages and the `/about` info
section. Safe to share with anyone, including prospective students.

> Signed-in, user-facing routes are in [user-routes.md](./user-routes.md).

---

## Auth

### `/login`
**Log in.** Two-step: enter credentials, then the OTP code emailed to you. On success,
goes to the home feed `/`.
**Suggest when:** "sign in", "log in", "I can't access my account".

### `/signup`
**Create an account.** Three steps: account info → profile details → email
verification. Requires a valid `@binus.ac.id` email; one account per student.
**Suggest when:** "sign up", "create an account", "how do I join".

### `/forgot-password?email=`
**Reset password.** Reset via an email OTP. Optional `?email=` pre-fills the address.
**Suggest when:** "forgot my password", "reset password", "can't remember password".

---

## About / Info

The `/about` section explains the product and policies. All pages are public.

| Route | Page | Suggest when |
|---|---|---|
| `/about` | About hub / landing — what Sunibgram is | "what is Sunibgram", "about the app" |
| `/about/what` | Why it exists + the 6 core principles | "why does this exist", "what do you stand for" |
| `/about/threads` | Threads guide — posting, replies, polls, reposts | "how do threads work", "how to post / poll" |
| `/about/communities` | Communities guide — public vs private, moderation | "how communities work", "public vs private" |
| `/about/events` | Events guide — official vs student-run, RSVPs | "how events work", "official vs unofficial events" |
| `/about/messages` | Messages guide — DMs, group rooms, read receipts | "how DMs work", "how does chat work" |
| `/about/leaderboard` | Leaderboard guide — weekly scoring, faculty boards | "how is the leaderboard scored" |
| `/about/use-cases` | Real scenarios (first-years, clubs, study groups…) | "how would I use this", "examples" |
| `/about/help` | Help Center + FAQ + support email | "help", "FAQ", "I need support" |
| `/about/team` | The team + independence (not official Binus) | "who built this", "is this official Binus" |
| `/about/privacy` | Privacy Center | "privacy policy", "how is my data used" |
| `/about/terms` | Terms & Conditions | "terms", "rules for using the app" |

For the policy text itself (so the agent can answer without sending the user away),
see [policies/](../policies/).
