import { api } from '../../api';
import type { Category, MessageResponse } from './types';

interface ListResponse {
  data: Category[];
}

export async function listCategories(token: string): Promise<Category[]> {
  const res = await api.get<ListResponse | Category[]>('/api/categories', token);
  if (Array.isArray(res)) return res;
  return res.data ?? [];
}

export function createCategory(token: string, name: string): Promise<MessageResponse> {
  return api.post<MessageResponse>('/api/categories', { name }, token);
}

export function updateCategory(token: string, id: string, name: string): Promise<MessageResponse> {
  return api.put<MessageResponse>(`/api/categories/${id}`, { name }, token);
}

export function deleteCategory(token: string, id: string): Promise<MessageResponse> {
  return api.delete<MessageResponse>(`/api/categories/${id}`, token);
}
