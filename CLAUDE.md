# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Sunibgram — university-exclusive social media platform (X/Threads clone). Binus University students only (`@binus.ac.id`). Features: threads, replies, reposts, likes, bookmarks, follows, communities.

**Monorepo:**
- `backend/` — Go/Fiber REST API (port 8080) → see [backend/CLAUDE.md](backend/CLAUDE.md)
- `frontend/` — Next.js app (port 3000) → see [frontend/CLAUDE.md](frontend/CLAUDE.md)
- `docker-compose.yml` — Postgres 16, Redis 7, MinIO, app container

---

## Dev Commands

**Infrastructure only:**
```bash
docker compose up postgres redis minio -d
```

**Backend locally:**
```bash
cd backend && go run ./cmd/main.go   # uses backend/.env
```

**Full stack via Docker:**
```bash
docker compose up --build            # uses backend/.env.docker
```

**Frontend:**
```bash
cd frontend
npm run dev      # dev server
npm run build    # production build
npm run lint     # eslint
```

**Build backend binary:**
```bash
cd backend && go build -o server ./cmd/main.go
```

**Backend lint/vet:**
```bash
cd backend && go vet ./...
```

**Backend tests:**
```bash
cd backend && go test ./...
```

---

## Environment Setup

Copy `backend/.env.example` → `backend/.env` and fill in values before running locally.

Required vars: `DB_*`, `JWT_SECRET`, `MINIO_*`, `ADMIN_EMAIL/PASSWORD/NAME/USERNAME`.

Frontend needs `frontend/.env.local` with:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_SOCKET_URL=http://localhost:8080
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=<openssl rand -base64 32>
```

---

## Architecture Summary

**Backend** — strict clean architecture, Go module `github.com/sunibgram`. See [backend/CLAUDE.md](backend/CLAUDE.md) for layer rules, patterns (R/W DB split, Redis caching, file validation), and adding new features.

**Frontend** — Next.js App Router, TypeScript strict. Prescribed stack: TanStack Query v5 (server state), Zustand v4 (client state), shadcn/ui, React Hook Form + Zod, NextAuth.js v5, Socket.io, Uploadthing, TipTap. See [frontend/CLAUDE.md](frontend/CLAUDE.md) for canonical types, patterns, and component structure.
