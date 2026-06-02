# Limits & Restrictions

Concrete numeric limits and format rules the agent can quote when users ask "how
long / how big / how many / what file types?". Values verified against source code —
see the source path next to each group.

> Sizes are server-enforced maximums. The frontend also compresses images before
> upload, so the file that actually lands is usually well under the cap.

---

## Posts & threads
Source: `frontend/lib/services/threads.ts`, `backend/domain/model/thread*`.
- **Post / reply text:** max **500 characters** (Unicode code points).
- **Media per post:** max **4** images/videos.
- **Allowed media types:** images `jpg, jpeg, png, webp, gif`; videos `mp4, mov, avi`.
- **Profile bio:** max **500 characters** (`backend/.../auth/auth_service.go`).

## Polls
Source: `backend/domain/model/thread_poll/thread_poll.go`. Polls only on post-type
threads (not replies/reposts).
- **Options:** **2–4** per poll.
- **Option label:** max **25 characters**.
- **Question:** max **280 characters**.
- **Duration:** **5 minutes** to **7 days** (300–604,800 seconds).
- **Option image:** max **5 MB** (jpg/jpeg/png/webp).

## File uploads (size caps)
Source: `backend/domain/model/media/media.go`.
| Upload | Max size | Allowed types |
|---|---|---|
| Avatar | 5 MB | jpg, jpeg, png, webp |
| Banner | 5 MB | jpg, jpeg, png, webp |
| Post / poll image | 5 MB | jpg, jpeg, png, webp (post also allows gif + video) |
| Chat attachment | 20 MB | (chat media) |
| Event document | 20 MB | `.doc` / `.docx` (official events) |

## Bookmark collections
Source: `frontend/lib/services/bookmarkCollections.ts`.
- **Collections per user:** max **50**.
- **Collection name:** max **50 characters**.

## Communities
Source: `frontend/components/community/{CreateCommunityModal,EditCommunityModal}.tsx`.
- **Name:** max **100 characters**.
- **Description:** max **2000 characters**.
- **Rules:** max **10** per community, each rule max **500 characters**.

## Group chats
Source: `frontend/components/chat/GroupSettingsModal.tsx`.
- **Group name:** max **60 characters**.
- **Group description:** max **200 characters**.

## Feedback
Source: `frontend/lib/services/feedback.ts`.
- **Title:** max **200 characters**.
- **Body:** max **5000 characters**.
- Type is **Bug** or **Feature request**.

## Account & security
Source: `backend/domain/model/user/user.go`, `backend/.../verification`, `.../auth`.
- **Email:** must end with **`@binus.ac.id`**.
- **Password:** min **10 characters**, and must contain at least one **uppercase**,
  one **lowercase**, and one **digit**.
- **One account per student.**
- **OTP code:** expires after **10 minutes**.
- **OTP rate limits:** sending a code is limited to **3 per 10 minutes**; verifying is
  limited to **3 per 10 minutes**; **5 failed** login-OTP attempts per 10 minutes
  triggers a temporary lockout.

## Premium
Source: `backend/application/usecase/order/order_service.go`,
`.../premium_subscription/premium_subscription_service.go`.
- **Price:** **Rp 10,000** (IDR), one-off order paid via Midtrans
  (item "Sunibgram Premium").
- **Free trial:** **1 month**, activatable **once per account** (a second attempt is
  rejected).
- **Premium unlocks:** creating communities and creating events (see
  [features.md](./features.md)).
