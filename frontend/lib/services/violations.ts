import { api } from '@/lib/api';

export interface Violation {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface ViolationsResponse {
  data: Violation[];
}

export function getViolations(token: string): Promise<ViolationsResponse> {
  return api.get<ViolationsResponse>('/api/violations', token);
}
