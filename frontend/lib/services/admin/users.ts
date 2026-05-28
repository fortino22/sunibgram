import { api } from '../../api';
import type { AdminUser, PaginatedResponse } from './types';

export function listUsers(token: string, page = 1, limit = 20): Promise<PaginatedResponse<AdminUser>> {
  return api.get<PaginatedResponse<AdminUser>>(`/api/admin/users/?page=${page}&limit=${limit}`, token);
}

export function getUser(token: string, id: number | string): Promise<AdminUser> {
  return api.get<AdminUser>(`/api/admin/users/${id}`, token);
}
