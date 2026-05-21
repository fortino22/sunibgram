import { create } from 'zustand';
import type { ThreadItem } from '@/lib/services/threads';

interface ComposeStore {
  open: boolean;
  openCompose: () => void;
  closeCompose: () => void;
  replyThread: ThreadItem | null;
  openReply: (thread: ThreadItem) => void;
  closeReply: () => void;
}

export const useComposeStore = create<ComposeStore>((set) => ({
  open: false,
  openCompose: () => set({ open: true }),
  closeCompose: () => set({ open: false }),
  replyThread: null,
  openReply: (thread) => set({ replyThread: thread }),
  closeReply: () => set({ replyThread: null }),
}));
