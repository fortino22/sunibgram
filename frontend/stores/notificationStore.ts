import { create } from 'zustand';
import type { NotificationItem } from '@/lib/services/notifications';

export interface ToastItem {
  id: string;
  notification: NotificationItem;
}

interface NotificationStore {
  unreadCount: number;
  increment: () => void;
  clearUnread: () => void;
  toasts: ToastItem[];
  addToast: (notification: NotificationItem) => void;
  removeToast: (id: string) => void;
}

export const useNotificationStore = create<NotificationStore>((set) => ({
  unreadCount: 0,
  increment: () => set((s) => ({ unreadCount: s.unreadCount + 1 })),
  clearUnread: () => set({ unreadCount: 0 }),
  toasts: [],
  addToast: (notification) =>
    set((s) => {
      const id = `toast-${notification.id}-${Date.now()}`;
      const next = [{ id, notification }, ...s.toasts].slice(0, 3);
      return { toasts: next };
    }),
  removeToast: (id) =>
    set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
}));
