# CLAUDE.md — University Social Media App (X/Threads Clone)

Source of truth for Claude Code. Read fully before writing code, creating files, or making architectural decisions.

---

## Project Summary

University-exclusive social media platform like X/Threads. Students post threads with text, photos, videos. Others comment, reply, repost, like, bookmark. **Portfolio project** — code quality, architecture, DX matter as much as features.

---

## Tech Stack (Non-Negotiable)

| Layer | Technology |
|---|---|
| Framework | Next.js 15 (App Router) |
| Language | TypeScript (strict mode) |
| Styling | Tailwind CSS v3 |
| UI Components | shadcn/ui |
| Server State | TanStack Query v5 |
| Client State | Zustand v4 |
| Forms | React Hook Form + Zod |
| Auth | NextAuth.js v5 |
| Real-time | Socket.io client |
| File Uploads | Uploadthing |
| Rich Text Editor | TipTap |

**Never suggest or use alternatives.** No Redux, SWR, Axios (use native fetch), Context API for global state, or unlisted libraries unless explicitly asked.

---

## Code Style Rules

### TypeScript
- **Always strict TypeScript.** No `any`. Use `unknown` if needed.
- Always define explicit return types.
- `interface` for object shapes, `type` for unions and primitives.
- Export types from `src/types/`.

```typescript
// ✅ Correct
interface Post {
  id: string;
  content: string;
  authorId: string;
  createdAt: Date;
  likeCount: number;
  repostCount: number;
}

// � Wrong
const post: any = {};
```

### Component Rules
- **Named exports** for all components. No default exports except `app/` pages (Next.js requirement).
- One component per file.
- Props always typed with `interface`.

```typescript
// ✅ Correct
interface PostCardProps {
  post: Post;
  showActions?: boolean;
}

export function PostCard({ post, showActions = true }: PostCardProps) { ... }

// � Wrong
export default function PostCard(props: any) { ... }
```

### File Naming
- Components: `PascalCase.tsx` (e.g., `PostCard.tsx`)
- Hooks: `camelCase.ts` prefixed with `use` (e.g., `usePosts.ts`)
- Utilities: `camelCase.ts` (e.g., `formatDate.ts`)
- Types: `camelCase.ts` (e.g., `post.ts`)

### Import Order (enforce always)
```typescript
// 1. React
import { useState, useEffect } from 'react';

// 2. Next.js
import Image from 'next/image';
import Link from 'next/link';

// 3. Third-party
import { useQuery } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';

// 4. Internal - absolute paths only
import { PostCard } from '@/components/PostCard';
import { usePosts } from '@/hooks/usePosts';
import type { Post } from '@/types/post';
```

---

## Project Structure

Follow exactly. No files outside this structure.

```
src/
├── app/                            # Next.js App Router
│   ├── layout.tsx                  # Root layout (providers, fonts)
│   ├── page.tsx                    # Home feed (authenticated)
│   ├── not-found.tsx
│   ├── error.tsx
│   ├── (auth)/                     # Route group — no shared layout with main
│   │   ├── layout.tsx
│   │   ├── login/
│   │   │   └── page.tsx
│   │   └── signup/
│   │       └── page.tsx
│   ├── post/
│   │   └── [id]/
│   │       └── page.tsx            # Single post + thread view
│   ├── user/
│   │   └── [username]/
│   │       └── page.tsx            # User profile
│   ├── bookmarks/
│   │   └── page.tsx
│   ├── notifications/
│   │   └── page.tsx
│   ├── @modal/                     # Parallel route for modals
│   │   ├── default.tsx             # Returns null
│   │   ├── compose/
│   │   │   └── page.tsx            # Post composition modal
│   │   └── reply/
│   │       └── [id]/
│   │           └── page.tsx        # Reply modal
│   └── api/
│       ├── auth/
│       │   └── [...nextauth]/route.ts
│       └── uploadthing/
│           └── route.ts
│
├── components/
│   ├── feed/
│   │   ├── Feed.tsx
│   │   ├── FeedSkeleton.tsx
│   │   └── InfiniteScrollTrigger.tsx
│   ├── post/
│   │   ├── PostCard.tsx
│   │   ├── PostActions.tsx         # Like, repost, reply, bookmark buttons
│   │   ├── PostHeader.tsx          # Avatar, username, timestamp
│   │   ├── PostContent.tsx         # Text, image, video
│   │   ├── PostComposer.tsx        # TipTap editor for creating posts
│   │   └── PostSkeleton.tsx
│   ├── comment/
│   │   ├── CommentThread.tsx
│   │   ├── CommentCard.tsx
│   │   └── CommentComposer.tsx
│   ├── layout/
│   │   ├── Sidebar.tsx             # Left nav sidebar
│   │   ├── RightPanel.tsx          # Right panel (who to follow, trends)
│   │   ├── MobileNav.tsx           # Bottom nav for mobile
│   │   └── TopBar.tsx
│   ├── user/
│   │   ├── Avatar.tsx
│   │   ├── UserCard.tsx
│   │   └── FollowButton.tsx
│   └── ui/                         # shadcn/ui components live here
│
├── hooks/
│   ├── usePosts.ts
│   ├── usePost.ts
│   ├── useComments.ts
│   ├── useCreatePost.ts
│   ├── useLikePost.ts
│   ├── useRepost.ts
│   ├── useBookmark.ts
│   ├── useFollow.ts
│   ├── useSocket.ts
│   └── useNotifications.ts
│
├── stores/
│   ├── authStore.ts                # Zustand: current user, session
│   └── uiStore.ts                  # Zustand: sidebar, theme, modals
│
├── lib/
│   ├── api.ts                      # Fetch wrapper with base URL + auth headers
│   ├── queryClient.ts              # TanStack Query client config
│   ├── socket.ts                   # Socket.io singleton client
│   ├── uploadthing.ts              # Uploadthing config
│   └── utils.ts                    # cn(), formatDate(), etc.
│
├── types/
│   ├── post.ts
│   ├── user.ts
│   ├── comment.ts
│   ├── notification.ts
│   └── index.ts                    # Re-exports all types
│
└── env.ts                           # Type-safe env vars (with Zod)
```

---

## Type Definitions

Canonical types. Use everywhere. Never redefine.

```typescript
// src/types/user.ts
export interface User {
  id: string;
  username: string;
  displayName: string;
  avatarUrl: string | null;
  bio: string | null;
  university: string;
  followerCount: number;
  followingCount: number;
  createdAt: Date;
}

// src/types/post.ts
export interface Post {
  id: string;
  content: string;
  mediaUrls: string[];
  author: User;
  parentId: string | null;        // null = original post, string = reply
  repostOfId: string | null;      // null = original, string = repost
  likeCount: number;
  repostCount: number;
  commentCount: number;
  bookmarkCount: number;
  isLiked: boolean;               // by current user
  isReposted: boolean;            // by current user
  isBookmarked: boolean;          // by current user
  createdAt: Date;
}

// src/types/comment.ts
export interface Comment {
  id: string;
  content: string;
  author: User;
  postId: string;
  parentCommentId: string | null; // null = top-level, string = reply to comment
  likeCount: number;
  isLiked: boolean;
  replies?: Comment[];
  createdAt: Date;
}

// src/types/notification.ts
export type NotificationType = 'like' | 'comment' | 'repost' | 'follow' | 'mention';

export interface Notification {
  id: string;
  type: NotificationType;
  actorId: string;
  actor: User;
  postId?: string;
  read: boolean;
  createdAt: Date;
}
```

---

## State Management Patterns

### TanStack Query (Server State)

All API data lives in TanStack Query. Never use `useState` for fetched data.

```typescript
// src/hooks/usePosts.ts
import { useInfiniteQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Post } from '@/types';

export function usePosts() {
  return useInfiniteQuery({
    queryKey: ['posts'],
    queryFn: ({ pageParam = 0 }) => api.get<Post[]>(`/posts?page=${pageParam}`),
    getNextPageParam: (lastPage, pages) =>
      lastPage.length === 20 ? pages.length : undefined,
    staleTime: 1000 * 60,
  });
}
```

**Query Key Conventions:**
```typescript
['posts']                        // all posts feed
['posts', postId]                // single post
['posts', postId, 'comments']    // comments for a post
['user', username]               // user profile
['user', username, 'posts']      // user's posts
['notifications']
['bookmarks']
```

### Zustand (Client State)

Only UI state and auth in Zustand. Never store fetched data here.

```typescript
// src/stores/authStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '@/types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  setUser: (user: User | null) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      setUser: (user) => set({ user, isAuthenticated: !!user }),
      logout: () => set({ user: null, isAuthenticated: false }),
    }),
    { name: 'auth-store' }
  )
);

// src/stores/uiStore.ts
interface UIState {
  sidebarOpen: boolean;
  theme: 'light' | 'dark';
  toggleSidebar: () => void;
  toggleTheme: () => void;
}

export const useUIStore = create<UIState>()((set) => ({
  sidebarOpen: true,
  theme: 'light',
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
  toggleTheme: () => set((s) => ({ theme: s.theme === 'light' ? 'dark' : 'light' })),
}));
```

---

## Mutations Pattern

All mutations use TanStack Query `useMutation` with optimistic updates.

```typescript
// src/hooks/useLikePost.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Post } from '@/types';

export function useLikePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ postId, isLiked }: { postId: string; isLiked: boolean }) =>
      isLiked
        ? api.delete(`/posts/${postId}/like`)
        : api.post(`/posts/${postId}/like`),

    onMutate: async ({ postId, isLiked }) => {
      await queryClient.cancelQueries({ queryKey: ['posts', postId] });

      const previousPost = queryClient.getQueryData<Post>(['posts', postId]);

      queryClient.setQueryData<Post>(['posts', postId], (old) => {
        if (!old) return old;
        return {
          ...old,
          isLiked: !isLiked,
          likeCount: isLiked ? old.likeCount - 1 : old.likeCount + 1,
        };
      });

      return { previousPost };
    },

    onError: (_err, { postId }, context) => {
      if (context?.previousPost) {
        queryClient.setQueryData(['posts', postId], context.previousPost);
      }
    },

    onSettled: (_data, _error, { postId }) => {
      queryClient.invalidateQueries({ queryKey: ['posts', postId] });
    },
  });
}
```

**Same optimistic update pattern for:** repost, bookmark, follow, comment like.

---

## Form Pattern

```typescript
// Always use React Hook Form + Zod together
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

const createPostSchema = z.object({
  content: z.string().min(1, 'Post cannot be empty').max(500, 'Too long'),
  mediaUrls: z.array(z.string().url()).max(4).optional(),
});

type CreatePostInput = z.infer<typeof createPostSchema>;

export function PostComposer() {
  const { register, handleSubmit, formState: { errors, isSubmitting } } =
    useForm<CreatePostInput>({
      resolver: zodResolver(createPostSchema),
    });

  const onSubmit = handleSubmit(async (data) => {
    await createPost(data);
  });

  return (
    <form onSubmit={onSubmit}>
      {/* Fields here */}
    </form>
  );
}
```

---

## API Client

All HTTP requests through `src/lib/api.ts`. Never raw fetch in components or hooks.

```typescript
// src/lib/api.ts
const BASE_URL = process.env.NEXT_PUBLIC_API_URL;

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include',
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message ?? 'Request failed');
  }

  return res.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
  patch: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'PATCH', body: JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
};
```

---

## Real-Time (Socket.io)

Socket.io for live notifications and real-time feed updates only.

```typescript
// src/lib/socket.ts — Singleton pattern
import { io, Socket } from 'socket.io-client';

let socket: Socket | null = null;

export function getSocket(): Socket {
  if (!socket) {
    socket = io(process.env.NEXT_PUBLIC_SOCKET_URL!, {
      withCredentials: true,
      autoConnect: false,
    });
  }
  return socket;
}

// src/hooks/useSocket.ts
import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { getSocket } from '@/lib/socket';
import { useAuthStore } from '@/stores/authStore';

export function useSocket() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();

  useEffect(() => {
    if (!isAuthenticated) return;

    const socket = getSocket();
    socket.connect();

    socket.on('post:liked', ({ postId }: { postId: string }) => {
      queryClient.invalidateQueries({ queryKey: ['posts', postId] });
    });

    socket.on('post:commented', ({ postId }: { postId: string }) => {
      queryClient.invalidateQueries({ queryKey: ['posts', postId, 'comments'] });
    });

    socket.on('notification:new', () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    });

    return () => {
      socket.disconnect();
    };
  }, [isAuthenticated, queryClient]);
}
```

---

## Component Patterns

### Loading States
Every data-fetching component must handle loading + error states.

```typescript
export function Feed() {
  const { data, isLoading, isError, fetchNextPage, hasNextPage } = usePosts();

  if (isLoading) return <FeedSkeleton />;
  if (isError) return <ErrorState message="Failed to load posts" />;

  const posts = data.pages.flatMap((page) => page);

  return (
    <div>
      {posts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}
      <InfiniteScrollTrigger
        onIntersect={fetchNextPage}
        disabled={!hasNextPage}
      />
    </div>
  );
}
```

### Skeleton Components
Every data component needs corresponding skeleton.

```typescript
// components/post/PostSkeleton.tsx
import { Skeleton } from '@/components/ui/skeleton';

export function PostSkeleton() {
  return (
    <div className="flex gap-3 p-4 border-b">
      <Skeleton className="h-10 w-10 rounded-full" />
      <div className="flex-1 space-y-2">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/4" />
      </div>
    </div>
  );
}
```

### Infinite Scroll Trigger
```typescript
// components/feed/InfiniteScrollTrigger.tsx
import { useEffect, useRef } from 'react';

interface InfiniteScrollTriggerProps {
  onIntersect: () => void;
  disabled?: boolean;
}

export function InfiniteScrollTrigger({ onIntersect, disabled }: InfiniteScrollTriggerProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (disabled || !ref.current) return;

    const observer = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting) onIntersect();
    });

    observer.observe(ref.current);
    return () => observer.disconnect();
  }, [onIntersect, disabled]);

  return <div ref={ref} className="h-1" />;
}
```

---

## UI Guidelines

### Design Tokens
```css
/* globals.css */
:root {
  --primary: #1d4ed8;          /* University blue */
  --primary-hover: #1e40af;
  --background: #ffffff;
  --surface: #f8fafc;
  --border: #e2e8f0;
  --text-primary: #0f172a;
  --text-secondary: #64748b;
  --destructive: #ef4444;
  --success: #22c55e;
}

.dark {
  --background: #0f172a;
  --surface: #1e293b;
  --border: #334155;
  --text-primary: #f8fafc;
  --text-secondary: #94a3b8;
}
```

### Layout
- Max content width: `640px` (feed), `600px` (post detail)
- Sidebar width: `275px` (desktop)
- Right panel width: `350px` (desktop)
- Mobile breakpoint: `768px` (hide sidebar, show bottom nav)

### shadcn/ui Components to Install
```bash
npx shadcn-ui@latest add button
npx shadcn-ui@latest add card
npx shadcn-ui@latest add input
npx shadcn-ui@latest add textarea
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add dropdown-menu
npx shadcn-ui@latest add avatar
npx shadcn-ui@latest add skeleton
npx shadcn-ui@latest add toast
npx shadcn-ui@latest add tabs
npx shadcn-ui@latest add separator
npx shadcn-ui@latest add tooltip
```

---

## Features Checklist

Implement in order:

### Phase 1 — Foundation
- [ ] Next.js 15 project setup with TypeScript strict mode
- [ ] Tailwind + shadcn/ui setup
- [ ] TanStack Query provider + Zustand stores
- [ ] Auth layout + login/signup pages with NextAuth.js
- [ ] Root layout with sidebar + right panel (desktop), bottom nav (mobile)

### Phase 2 — Core Features
- [ ] Home feed with infinite scroll (`/`)
- [ ] PostCard component with all actions (like, repost, comment, bookmark, share)
- [ ] Single post + thread view (`/post/[id]`)
- [ ] Create post modal with TipTap editor and image upload
- [ ] Reply to post (parallel route modal)
- [ ] Comment thread with nested replies (max depth 3)

### Phase 3 — Social Features
- [ ] User profile page (`/user/[username]`)
- [ ] Follow / Unfollow with optimistic update
- [ ] Repost + Quote repost
- [ ] Bookmarks page
- [ ] Notifications page

### Phase 4 — Real-Time
- [ ] Socket.io integration
- [ ] Live like/comment counts
- [ ] Real-time notification badge
- [ ] Live notification feed

### Phase 5 — Polish
- [ ] Dark mode toggle
- [ ] Image upload via Uploadthing
- [ ] Video post support
- [ ] Search page
- [ ] Responsive mobile layout

---

## Environment Variables

```env
# .env.local

# Backend
NEXT_PUBLIC_API_URL=http://localhost:3001
NEXT_PUBLIC_SOCKET_URL=http://localhost:3001

# Auth
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=generate-with-openssl-rand-base64-32

# Uploadthing
UPLOADTHING_SECRET=
NEXT_PUBLIC_UPLOADTHING_APP_ID=
```

---

## Setup Commands

```bash
# Create project
npx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir

# Install all dependencies
npm install @tanstack/react-query @tanstack/react-query-devtools zustand \
  react-hook-form @hookform/resolvers zod \
  next-auth socket.io-client \
  uploadthing @uploadthing/react \
  @tiptap/react @tiptap/starter-kit @tiptap/extension-placeholder \
  clsx tailwind-merge class-variance-authority

# Initialize shadcn/ui
npx shadcn-ui@latest init

# Install shadcn components
npx shadcn-ui@latest add button card input textarea dialog dropdown-menu avatar skeleton toast tabs separator tooltip
```

---

## Common Mistakes to Avoid

| � Don't | ✅ Do |
|---|---|
| Use `any` type | Use explicit types or `unknown` |
| Store API data in `useState` | Use TanStack Query |
| Use Context API for global state | Use Zustand |
| Use raw `fetch` in components | Use `api` from `lib/api.ts` |
| Create new socket on each render | Use singleton from `lib/socket.ts` |
| Skip loading/error states | Always handle all async states |
| Default export components | Named exports (except app/ pages) |
| Skip optimistic updates on mutations | Always add optimistic updates |
| Inline Zod schemas | Define schemas in `lib/validators.ts` |
| Mix server/client components blindly | Mark client components with `'use client'` |

---

## Portfolio Notes

Project demonstrates:
1. **Modern React architecture** — Server Components, parallel routes, streaming
2. **Professional state management** — TanStack Query for async, Zustand for UI
3. **Real-time capability** — WebSocket integration
4. **Type safety** — strict TypeScript end-to-end
5. **Scalable component design** — composable, single-responsibility
6. **Performance mindset** — optimistic updates, infinite scroll, skeletons, caching

When explaining: focus on **why** each decision was made, not just what was built.