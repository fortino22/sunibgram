import { create } from 'zustand';
import type { ThreadItem } from '@/lib/services/threads';

interface ComposeStore {
  open: boolean;
  openCompose: () => void;
  closeCompose: () => void;
  preselectedEventId?: string;
  preselectedEventTitle?: string;
  preselectedIsOfficialEvent?: boolean;
  openComposeForEvent: (id: string, title: string) => void;
  openOfficialEventPost: (id: string, title: string) => void;
  replyThread: ThreadItem | null;
  openReply: (thread: ThreadItem) => void;
  closeReply: () => void;
}

export const useComposeStore = create<ComposeStore>((set) => ({
  open: false,
  openCompose: () => set({ open: true, preselectedEventId: undefined, preselectedEventTitle: undefined, preselectedIsOfficialEvent: undefined }),
  closeCompose: () => set({ open: false, preselectedEventId: undefined, preselectedEventTitle: undefined, preselectedIsOfficialEvent: undefined }),
  preselectedEventId: undefined,
  preselectedEventTitle: undefined,
  preselectedIsOfficialEvent: undefined,
  openComposeForEvent: (id, title) => set({ open: true, preselectedEventId: id, preselectedEventTitle: title, preselectedIsOfficialEvent: false }),
  openOfficialEventPost: (id, title) => set({ open: true, preselectedEventId: id, preselectedEventTitle: title, preselectedIsOfficialEvent: true }),
  replyThread: null,
  openReply: (thread) => set({ replyThread: thread }),
  closeReply: () => set({ replyThread: null }),
}));
