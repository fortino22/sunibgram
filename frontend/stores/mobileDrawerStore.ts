import { create } from 'zustand';

interface MobileDrawerStore {
  open: boolean;
  openDrawer: () => void;
  closeDrawer: () => void;
  toggleDrawer: () => void;
}

export const useMobileDrawerStore = create<MobileDrawerStore>((set) => ({
  open: false,
  openDrawer: () => set({ open: true }),
  closeDrawer: () => set({ open: false }),
  toggleDrawer: () => set((s) => ({ open: !s.open })),
}));
