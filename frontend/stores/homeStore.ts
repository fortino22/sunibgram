import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface PinnedCommunity {
  id: number;
  name: string;
}

interface HomeState {
  pinnedCommunity: PinnedCommunity | null;
  tab3Mode: 'faculty' | 'community';
  setPinnedCommunity: (community: PinnedCommunity | null) => void;
  setTab3Mode: (mode: 'faculty' | 'community') => void;
}

export const useHomeStore = create<HomeState>()(
  persist(
    (set) => ({
      pinnedCommunity: null,
      tab3Mode: 'faculty',
      setPinnedCommunity: (community) =>
        set((s) => ({
          pinnedCommunity: community,
          tab3Mode: community ? s.tab3Mode : 'faculty',
        })),
      setTab3Mode: (mode) => set({ tab3Mode: mode }),
    }),
    { name: 'home-store' }
  )
);
