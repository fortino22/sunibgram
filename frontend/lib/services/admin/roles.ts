import { api } from '../../api';
import type { MessageResponse, Role } from './types';

export function listRoles(token: string): Promise<Role[]> {
  return api.get<Role[]>('/api/admin/roles/', token);
}

export function createRole(token: string, name: string): Promise<Role> {
  return api.post<Role>('/api/admin/roles/', { name }, token);
}

export function updateRole(token: string, id: number, name: string): Promise<Role> {
  return api.put<Role>(`/api/admin/roles/${id}`, { name }, token);
}

export function deleteRole(token: string, id: number): Promise<MessageResponse> {
  return api.delete<MessageResponse>(`/api/admin/roles/${id}`, token);
}
