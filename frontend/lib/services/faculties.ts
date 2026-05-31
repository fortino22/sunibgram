import { api } from '../api';

export interface FacultyMessageResponse {
  message: string;
}

export interface Faculty {
  faculty_id: number;
  faculty_name: string;
  [key: string]: unknown;
}

export interface FacultyResponse {
  data: Faculty;
  message?: string;
  [key: string]: unknown;
}

export interface FacultyListResponse {
  data: Faculty[];
  [key: string]: unknown;
}

export interface FacultyParams {
  facultyName: string;
}

export interface AdminUser {
  id: number;
  username?: string;
  email?: string;
  role_id?: number;
  [key: string]: unknown;
}

export interface AdminUserResponse {
  data: AdminUser;
  [key: string]: unknown;
}

export interface AdminUsersResponse {
  data: AdminUser[];
  page?: number;
  limit?: number;
  total?: number;
  [key: string]: unknown;
}

export interface Role {
  id: number;
  name: string;
  [key: string]: unknown;
}

export interface RoleResponse {
  data: Role;
  message?: string;
  [key: string]: unknown;
}

export interface RoleListResponse {
  data: Role[];
  [key: string]: unknown;
}

export interface AdminPaginationParams {
  page?: number;
  limit?: number;
}

function adminQuery(params: AdminPaginationParams = {}): string {
  const search = new URLSearchParams();
  if (params.page !== undefined) {
    search.set('page', String(params.page));
  }
  if (params.limit !== undefined) {
    search.set('limit', String(params.limit));
  }
  const value = search.toString();
  return value ? `?${value}` : '';
}

// Public endpoint — token optional (omit it pre-auth, e.g. on the signup form;
// passing '' would make the api wrapper throw before fetching).
export function getFaculties(token?: string): Promise<FacultyListResponse> {
  return api.get<FacultyListResponse>('/api/faculties', token);
}

export function getFaculty(id: number | string, token: string): Promise<FacultyResponse> {
  return api.get<FacultyResponse>(`/api/faculties/${id}`, token);
}

export function createFaculty(params: FacultyParams, token: string): Promise<FacultyResponse> {
  return api.post<FacultyResponse>('/api/faculties', { faculty_name: params.facultyName }, token);
}

export function updateFaculty(
  id: number | string,
  params: FacultyParams,
  token: string
): Promise<FacultyResponse> {
  return api.put<FacultyResponse>(`/api/faculties/${id}`, { faculty_name: params.facultyName }, token);
}

export function deleteFaculty(id: number | string, token: string): Promise<FacultyMessageResponse> {
  return api.delete<FacultyMessageResponse>(`/api/faculties/${id}`, token);
}

export function getAdminUsers(
  params: AdminPaginationParams = {},
  token: string
): Promise<AdminUsersResponse> {
  return api.get<AdminUsersResponse>(`/api/admin/users/${adminQuery(params)}`, token);
}

export function getAdminUser(id: number | string, token: string): Promise<AdminUserResponse> {
  return api.get<AdminUserResponse>(`/api/admin/users/${id}`, token);
}

export function getRoles(token: string): Promise<RoleListResponse> {
  return api.get<RoleListResponse>('/api/admin/roles/', token);
}

export function createRole(name: string, token: string): Promise<RoleResponse> {
  return api.post<RoleResponse>('/api/admin/roles/', { name }, token);
}

export function getRole(id: number | string, token: string): Promise<RoleResponse> {
  return api.get<RoleResponse>(`/api/admin/roles/${id}`, token);
}

export function updateRole(id: number | string, name: string, token: string): Promise<RoleResponse> {
  return api.put<RoleResponse>(`/api/admin/roles/${id}`, { name }, token);
}

export function deleteRole(id: number | string, token: string): Promise<FacultyMessageResponse> {
  return api.delete<FacultyMessageResponse>(`/api/admin/roles/${id}`, token);
}
