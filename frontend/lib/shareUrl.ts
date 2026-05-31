export function buildThreadShareUrl(threadId: number | string): string {
  const fromEnv = process.env.NEXT_PUBLIC_APP_URL;
  const base = fromEnv && fromEnv.length > 0
    ? fromEnv.replace(/\/+$/, '')
    : (typeof window !== 'undefined' ? window.location.origin : '');
  return `${base}/thread/${threadId}`;
}

export function buildProfileShareUrl(userId: number | string): string {
  const fromEnv = process.env.NEXT_PUBLIC_APP_URL;
  const base = fromEnv && fromEnv.length > 0
    ? fromEnv.replace(/\/+$/, '')
    : (typeof window !== 'undefined' ? window.location.origin : '');
  return `${base}/profile/${userId}`;
}
