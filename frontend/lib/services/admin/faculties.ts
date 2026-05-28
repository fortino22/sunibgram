import { api } from '../../api';
import type { Faculty, MessageResponse, PaginatedResponse } from './types';

interface ListResponse {
  data: Faculty[];
  meta?: { total: number; page: number; limit: number };
}

export async function listFaculties(token: string, page = 1, limit = 100): Promise<Faculty[]> {
  const res = await api.get<ListResponse | Faculty[]>(`/api/faculties?page=${page}&limit=${limit}`, token);
  if (Array.isArray(res)) return res;
  return res.data ?? [];
}

export function listFacultiesPaginated(token: string, page = 1, limit = 20): Promise<PaginatedResponse<Faculty>> {
  return api.get<PaginatedResponse<Faculty>>(`/api/faculties?page=${page}&limit=${limit}`, token);
}

export function createFaculty(token: string, faculty_name: string): Promise<MessageResponse> {
  return api.post<MessageResponse>('/api/faculties', { faculty_name }, token);
}

export function updateFaculty(token: string, id: number, faculty_name: string): Promise<MessageResponse> {
  return api.put<MessageResponse>(`/api/faculties/${id}`, { faculty_name }, token);
}

export function deleteFaculty(token: string, id: number): Promise<MessageResponse> {
  return api.delete<MessageResponse>(`/api/faculties/${id}`, token);
}
