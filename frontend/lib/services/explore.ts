import { api } from '../api';

export interface ExploreUserResponse {
  userId: number;
  profile: string;
  username: string;
  name: string;
  is_premium: boolean;
  is_following: boolean;
}

export interface ExploreThreadResponse {
  content: string;
}

export interface ExploreResponse {
  users: ExploreUserResponse[];
  threads: ExploreThreadResponse[];
}

export function explore(query: string, token?: string): Promise<ExploreResponse> {
  const params = new URLSearchParams();
  params.set('q', query);
  return api.get<ExploreResponse>(`/api/explore?${params.toString()}`, token);
}
