'use client';

import { useEffect, type ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { type AuthUser, useAuthStore } from '@/stores/authStore';

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

interface RefreshResponse {
  access_token: string;
  sid?: string;
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

function isRefreshResponse(value: unknown): value is RefreshResponse {
  if (typeof value !== 'object' || value === null) return false;

  const response = value as Record<string, unknown>;
  return (
    typeof response.access_token === 'string' &&
    (response.sid === undefined || typeof response.sid === 'string')
  );
}

export function AuthGuard({ children }: AuthGuardProps) {
  const { token, isAuthenticated, initialized, setToken, setUser, setInitialized, logout } =
    useAuthStore();
  const router = useRouter();

  useEffect(() => {
    let cancelled = false;

    const refresh = async () => {
      if (token) {
        setInitialized(true);
        return;
      }

      try {
        const response = await fetch('/auth/refresh');

        if (!response.ok) {
          throw new Error('Refresh failed');
        }

        const data: unknown = await response.json();
        if (!isRefreshResponse(data)) {
          throw new Error('Invalid refresh response');
        }

        if (cancelled) return;

        setToken(data.access_token, data.sid ?? null);
        setUser(decodeJwtUser(data.access_token));
        setInitialized(true);
      } catch {
        if (cancelled) return;

        logout();
        setInitialized(true);
        router.replace('/login');
      }
    };

    void refresh();

    return () => {
      cancelled = true;
    };
  }, [logout, router, setInitialized, setToken, setUser, token]);

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
