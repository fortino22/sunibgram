'use client';

import { useEffect } from 'react';
import { socketManager } from '@/lib/socket';
import { useAuthStore } from '@/stores/authStore';

export function useSocket() {
  const token = useAuthStore((s) => s.token);

  useEffect(() => {
    if (!token) return;
    const socketUrl =
      process.env.NEXT_PUBLIC_SOCKET_URL ||
      process.env.NEXT_PUBLIC_API_URL ||
      '';
    socketManager.connect(token, socketUrl);
    return () => socketManager.disconnect();
  }, [token]);

  return socketManager;
}
