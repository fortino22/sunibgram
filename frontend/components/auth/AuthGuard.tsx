'use client';

import { useEffect, useRef, type ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { useQueryClient } from '@tanstack/react-query';
import { type AuthUser, useAuthStore } from '@/stores/authStore';
import { refreshOnce } from '@/lib/auth/refreshClient';

interface AuthGuardProps {
  children: ReactNode;
}

interface JwtPayload {
  user_id: number;
  username: string;
  email: string;
  role_id: number;
  role_name: string;
}

function isJwtPayload(value: unknown): value is JwtPayload {
  if (typeof value !== 'object' || value === null) return false;

  const payload = value as Record<string, unknown>;
  return (
    typeof payload.user_id === 'number' &&
    typeof payload.username === 'string' &&
    typeof payload.email === 'string' &&
    typeof payload.role_id === 'number' &&
    typeof payload.role_name === 'string'
  );
}

function decodeJwtUser(token: string): AuthUser | null {
  try {
    const payload: unknown = JSON.parse(atob(token.split('.')[1]));
    if (!isJwtPayload(payload)) return null;

    return {
      id: payload.user_id,
      username: payload.username,
      email: payload.email,
      roleId: payload.role_id,
      roleName: payload.role_name,
    };
  } catch {
    return null;
  }
}

const TRANSIENT_BACKOFF_MS = [400, 800, 1600, 3200];

export function AuthGuard({ children }: AuthGuardProps) {
  const isAuthenticated = useAuthStore((s: { isAuthenticated: boolean }) => s.isAuthenticated);
  const initialized = useAuthStore((s: { initialized: boolean }) => s.initialized);
  const router = useRouter();
  const queryClient = useQueryClient();
  const attemptedRef = useRef(false);

  useEffect(() => {
    if (attemptedRef.current) return;
    attemptedRef.current = true;

    let cancelled = false;

    const attempt = async () => {
      const { token, setToken, setUser, setInitialized, logout } = useAuthStore.getState();

      if (token) {
        setInitialized(true);
        return;
      }

      for (let i = 0; i <= TRANSIENT_BACKOFF_MS.length; i++) {
        if (cancelled) return;

        const result = await refreshOnce();
        if (cancelled) return;

        if (result.status === 'ok') {
          setToken(result.accessToken, result.sid);
          setUser(decodeJwtUser(result.accessToken));
          setInitialized(true);
          return;
        }

        if (result.status === 'revoked') {
          logout();
          queryClient.clear(); // revoked session: drop stale cache so re-login is clean
          setInitialized(true);
          router.replace('/login');
          return;
        }

        // transient — backoff + retry, do NOT logout
        const backoff = TRANSIENT_BACKOFF_MS[i];
        if (backoff === undefined) break;
        const delay = result.retryAfterMs ?? backoff;
        await new Promise<void>((resolve) => setTimeout(resolve, delay));
      }

      if (cancelled) return;
      // exhausted retries — mark initialized but do NOT logout; keep cookie
      // user sees loading -> login prompt only after explicit revoke
      setInitialized(true);
      router.replace('/login');
    };

    void attempt();

    return () => {
      cancelled = true;
      attemptedRef.current = false; // allow remount to retry a cancelled attempt
    };
  }, [router, queryClient]);

  useEffect(() => {
    if (initialized && !isAuthenticated) {
      router.replace('/login');
    }
  }, [initialized, isAuthenticated, router]);

  if (!initialized || !isAuthenticated) {
    return (
      <div
        style={{
          minHeight: '100vh',
          background: 'var(--bg)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <div
          style={{
            width: 32,
            height: 32,
            borderRadius: '50%',
            border: '2px solid var(--border)',
            borderTopColor: 'var(--accent)',
            animation: 'spin 0.7s linear infinite',
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  return <>{children}</>;
}
