import { create } from 'zustand';

interface ChatState {
  activeRoomId: number | null;
  setActiveRoomId: (id: number | null) => void;
}

export const useChatStore = create<ChatState>((set) => ({
  activeRoomId: null,
  setActiveRoomId: (id) => set({ activeRoomId: id }),
}));
