import { api } from '@/lib/api';

interface UserCountResponse {
  total: number;
}

export async function getUserCount(): Promise<number> {
  const res = await api.get<UserCountResponse>('/api/users/count');
  return res.total;
}
