import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface AuthUser {
  id: number;
  username: string;
  email: string;
  roleId: number;
  roleName: string;
}

interface AuthState {
  token: string | null;
  sid: string | null;
  user: AuthUser | null;
  isAuthenticated: boolean;
  initialized: boolean;
  setToken: (token: string, sid: string | null) => void;
  setUser: (user: AuthUser | null) => void;
  setInitialized: (value: boolean) => void;
  logout: () => void;
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

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      sid: null,
      user: null,
      isAuthenticated: false,
      initialized: false,
      setToken: (token: string, sid: string | null) => {
        const user = decodeJwtUser(token);
        set({ token, sid, user, isAuthenticated: token !== null });
      },
      setUser: (user: AuthUser | null) => set({ user }),
      setInitialized: (value: boolean) => set({ initialized: value }),
      logout: () =>
        set({ token: null, sid: null, user: null, isAuthenticated: false, initialized: true }),
    }),
    {
      name: 'auth-store',
      partialize: (state) => ({ user: state.user }),
    }
  )
);
