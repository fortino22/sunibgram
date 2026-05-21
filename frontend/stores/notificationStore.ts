import { create } from 'zustand';

interface NotificationStore {
  unreadCount: number;
  increment: () => void;
  clearUnread: () => void;
}

export const useNotificationStore = create<NotificationStore>((set) => ({
  unreadCount: 0,
  increment: () => set((s) => ({ unreadCount: s.unreadCount + 1 })),
  clearUnread: () => set({ unreadCount: 0 }),
}));
