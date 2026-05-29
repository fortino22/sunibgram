async function request<T>(
  path: string,
  options: RequestInit & { token?: string } = {}
): Promise<T> {
  const { token, ...fetchOptions } = options;
  const baseUrl = (process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  const url = `${baseUrl}${normalizedPath}`;

  const headers: Record<string, string> = {};

 
  if (token !== undefined) {
    if (typeof token !== 'string' || token.length === 0) {
      throw new Error('Authentication token is required for this request.');
    }
    headers['Authorization'] = `Bearer ${token}`;
  }

  if (!(fetchOptions.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(url, {
    ...fetchOptions,
    headers: { ...headers, ...(fetchOptions.headers as Record<string, string> | undefined) },
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: 'Request failed' }));
    throw new Error((error as { message?: string }).message ?? 'Request failed');
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string, token?: string) =>
    request<T>(path, { method: 'GET', token }),

  post: <T>(path: string, body?: unknown, token?: string) =>
    request<T>(path, {
      method: 'POST',
      body: body instanceof FormData ? body : JSON.stringify(body),
      token,
    }),

  put: <T>(path: string, body?: unknown, token?: string) =>
    request<T>(path, {
      method: 'PUT',
      body: JSON.stringify(body),
      token,
    }),

  patch: <T>(path: string, body?: unknown, token?: string) =>
    request<T>(path, {
      method: 'PATCH',
      body: body instanceof FormData ? body : JSON.stringify(body),
      token,
    }),

  delete: <T>(path: string, token?: string) =>
    request<T>(path, { method: 'DELETE', token }),
};

export const isInactiveCommunityError = (err: unknown): boolean =>
  err instanceof Error && err.message.toLowerCase().includes('community is inactive');

export const INACTIVE_COMMUNITY_TOAST =
  'Please resubscribe your premium to perform this action.';
