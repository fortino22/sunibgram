import { create } from 'zustand';
import type { ThreadItem } from '@/lib/services/threads';

export interface AnchorRect {
  top: number;
  left: number;
  bottom: number;
  right: number;
  width: number;
}

interface SummaryStore {
  summaryThread: ThreadItem | null;
  anchorRect: AnchorRect | null;
  openSummary: (thread: ThreadItem, rect: AnchorRect) => void;
  closeSummary: () => void;
}

export const useSummaryStore = create<SummaryStore>((set) => ({
  summaryThread: null,
  anchorRect: null,
  openSummary: (thread, rect) => set({ summaryThread: thread, anchorRect: rect }),
  closeSummary: () => set({ summaryThread: null, anchorRect: null }),
}));
