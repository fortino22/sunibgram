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
  setToken: (token: string, sid: string) => void;
  logout: () => void;
}

function decodeJwtUser(token: string): AuthUser | null {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
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
      setToken: (token: string, sid: string) => {
        const user = decodeJwtUser(token);
        set({ token, sid, user, isAuthenticated: !!user });
      },
      logout: () => set({ token: null, sid: null, user: null, isAuthenticated: false }),
    }),
    { name: 'auth-store' }
  )
);
