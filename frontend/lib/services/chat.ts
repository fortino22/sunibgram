import { api } from '@/lib/api';

export interface ChatSender {
  id: number;
  username: string;
}

export interface ChatMedia {
  media_id: number;
  url: string;
  media_type: string;
  file_size: number;
}

export interface ChatMessage {
  chat_id: number;
  room_id: number;
  sender: ChatSender;
  message: string | null;
  media: ChatMedia[];
  is_read: boolean;
  deleted_at: string | null;
  created_at: string;
}

export interface ChatRoom {
  room_id: number;
  name: string | null;
  description: string | null;
  type: 'private' | 'group';
  member_count?: number;
  created_at: string;
  updated_at: string;
}

export interface ChatRoomsResponse {
  data: ChatRoom[];
  meta: { total: number; page: number; limit: number };
}

export interface ChatMessagesResponse {
  data: ChatMessage[];
  meta: { total: number; page: number; limit: number };
}

export interface MessageEditedPayload {
  chat_id: number;
  room_id: number;
  sender_id: number;
  message: string;
}

export interface MessageDeletedPayload {
  chat_id: number;
  room_id: number;
  sender_id: number;
  deleted_by: number;
}

export interface MessageReadPayload {
  room_id: number;
  chat_ids: number[];
  read_by: number;
}

export interface TypingPayload {
  room_id: number;
  user_id: number;
}

export function getMyRooms(token: string, page = 1, limit = 30): Promise<ChatRoomsResponse> {
  return api.get<ChatRoomsResponse>(`/api/chats/rooms?page=${page}&limit=${limit}`, token);
}

export function getMessages(roomId: number, token: string, page = 1, limit = 40): Promise<ChatMessagesResponse> {
  return api.get<ChatMessagesResponse>(
    `/api/chats/rooms/${roomId}/messages?page=${page}&limit=${limit}`,
    token
  );
}

export async function sendMessage(
  roomId: number,
  message: string | null,
  files: File[],
  token: string
): Promise<void> {
  const form = new FormData();
  if (message) form.append('message', message);
  files.forEach((f) => form.append('files', f));
  await api.post<unknown>(`/api/chats/rooms/${roomId}/messages`, form, token);
}

export function createPrivateRoom(
  targetUserId: number,
  token: string
): Promise<{ room_id: number }> {
  return api.post<{ room_id: number }>(
    '/api/chats/rooms/private',
    { target_user_id: targetUserId },
    token
  );
}

export function createGroupRoom(
  name: string,
  memberIds: number[],
  description: string | undefined,
  token: string
): Promise<{ room_id: number }> {
  return api.post<{ room_id: number }>(
    '/api/chats/rooms/group',
    { name, member_ids: memberIds, description },
    token
  );
}

export async function unsendMessage(chatId: number, token: string): Promise<void> {
  await api.delete<unknown>(`/api/chats/messages/${chatId}`, token);
}

export async function editMessage(chatId: number, message: string, token: string): Promise<void> {
  await api.patch<unknown>(`/api/chats/messages/${chatId}`, { message }, token);
}

export async function markMessagesRead(
  roomId: number,
  chatIds: number[],
  token: string
): Promise<void> {
  await api.post<unknown>('/api/chats/messages/read', { room_id: roomId, chat_ids: chatIds }, token);
}
