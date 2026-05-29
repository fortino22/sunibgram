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

export interface ChatReactionSummary {
  emoji: string;
  count: number;
  user_ids: number[];
}

export interface ChatReplyInfo {
  chat_id: number;
  sender: ChatSender;
  message: string | null;
  gif_url?: string | null;
}

export interface SharedThreadAuthor {
  user_id: number;
  username: string;
  name: string;
  is_premium: boolean;
}

export interface SharedThreadEmbed {
  thread_id: number;
  content: string;
  author?: SharedThreadAuthor | null;
  community_name?: string | null;
  created_at?: string | null;
  deleted: boolean;
}

export interface ChatMessage {
  chat_id: number;
  room_id: number;
  sender: ChatSender;
  reply_to?: ChatReplyInfo | null;
  message: string | null;
  gif_url?: string | null;
  shared_thread?: SharedThreadEmbed | null;
  media: ChatMedia[];
  reactions?: ChatReactionSummary[];
  is_read: boolean;
  is_edited?: boolean;
  edited_at?: string | null;
  deleted_at: string | null;
  created_at: string;
}

export interface ChatRoomParticipant {
  user_id: number;
  username: string;
  name: string;
  faculty_name: string;
  avatar_url?: string | null;
}

export interface LastMessagePreview {
  chat_id: number;
  sender_id: number;
  message: string | null;
  gif_url?: string | null;
  has_media: boolean;
  created_at: string;
}

export interface ChatRoom {
  room_id: number;
  name: string | null;
  description: string | null;
  avatar_url: string | null;
  type: 'private' | 'group';
  member_count: number;
  is_muted: boolean;
  unread_count: number;
  other_user?: ChatRoomParticipant | null;
  last_message?: LastMessagePreview | null;
  created_at: string;
  updated_at: string;
}

export interface GroupMember {
  user_id: number;
  username: string;
  name: string;
  faculty_name: string;
  avatar_url: string;
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
  is_edited: boolean;
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
  token: string,
  gifUrl?: string | null,
  replyToId?: number | null,
): Promise<void> {
  const form = new FormData();
  if (message) form.append('message', message);
  if (gifUrl) form.append('gif_url', gifUrl);
  if (replyToId) form.append('reply_to_id', String(replyToId));
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

export async function updateGroupRoom(
  roomId: number,
  name: string,
  description: string | undefined,
  avatar: File | null,
  token: string
): Promise<void> {
  const form = new FormData();
  form.append('name', name);
  if (description) form.append('description', description);
  if (avatar) form.append('avatar', avatar);
  await api.patch<unknown>(`/api/chats/rooms/${roomId}`, form, token);
}

export function getGroupMembers(roomId: number, token: string): Promise<{ data: GroupMember[] }> {
  return api.get<{ data: GroupMember[] }>(`/api/chats/rooms/${roomId}/members`, token);
}

export async function addGroupMembers(roomId: number, userIds: number[], token: string): Promise<void> {
  await api.post<unknown>(`/api/chats/rooms/${roomId}/members`, { user_ids: userIds }, token);
}

export async function toggleMuteRoom(roomId: number, token: string): Promise<{ is_muted: boolean }> {
  return api.post<{ is_muted: boolean }>(`/api/chats/rooms/${roomId}/mute`, {}, token);
}

export async function leaveGroup(roomId: number, token: string): Promise<void> {
  await api.delete<unknown>(`/api/chats/rooms/${roomId}/leave`, token);
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

export async function addReaction(chatId: number, emoji: string, token: string): Promise<void> {
  await api.post<unknown>(`/api/chats/messages/${chatId}/reactions`, { emoji }, token);
}

export async function removeReaction(chatId: number, emoji: string, token: string): Promise<void> {
  await api.delete<unknown>(
    `/api/chats/messages/${chatId}/reactions/${encodeURIComponent(emoji)}`,
    token
  );
}
