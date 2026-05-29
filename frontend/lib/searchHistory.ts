// Recent explore searches — device-local, no backend. SSR-safe.
// Holds two kinds of entries: typed queries and clicked users.
// Scoped per user so different accounts on the same browser don't share history.

const KEY_PREFIX = 'sunibgram:recent-searches';
const LEGACY_KEY = 'sunibgram:recent-searches';
const MAX_ITEMS = 8;

export type RecentSearch =
  | { type: 'query'; text: string }
  | { type: 'user'; userId: number; username: string; name: string; profile: string };

/** Stable identity for dedupe / removal. */
export function recentKey(item: RecentSearch): string {
  return item.type === 'query' ? `q:${item.text.trim().toLowerCase()}` : `u:${item.userId}`;
}

function storageKey(userId: number): string {
  return `${KEY_PREFIX}:${userId}`;
}

function dropLegacyKey(): void {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.removeItem(LEGACY_KEY);
  } catch {
    // ignore
  }
}

function coerce(raw: unknown): RecentSearch | null {
  // Legacy format: plain string → query item.
  if (typeof raw === 'string') {
    const text = raw.trim();
    return text ? { type: 'query', text } : null;
  }
  if (raw && typeof raw === 'object') {
    const o = raw as Record<string, unknown>;
    if (o.type === 'query' && typeof o.text === 'string' && o.text.trim()) {
      return { type: 'query', text: o.text.trim() };
    }
    if (o.type === 'user' && typeof o.userId === 'number') {
      return {
        type: 'user',
        userId: o.userId,
        username: typeof o.username === 'string' ? o.username : '',
        name: typeof o.name === 'string' ? o.name : '',
        profile: typeof o.profile === 'string' ? o.profile : '',
      };
    }
  }
  return null;
}

function read(userId: number): RecentSearch[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = window.localStorage.getItem(storageKey(userId));
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.map(coerce).filter((v): v is RecentSearch => v !== null);
  } catch {
    return [];
  }
}

function write(userId: number, items: RecentSearch[]): void {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(storageKey(userId), JSON.stringify(items));
  } catch {
    // ignore quota / serialization errors
  }
}

function add(userId: number, item: RecentSearch): void {
  const key = recentKey(item);
  const next = [item, ...read(userId).filter((existing) => recentKey(existing) !== key)].slice(0, MAX_ITEMS);
  write(userId, next);
}

export function getRecentSearches(userId: number | null | undefined): RecentSearch[] {
  if (!userId) return [];
  dropLegacyKey();
  return read(userId);
}

export function addRecentQuery(userId: number | null | undefined, text: string): void {
  if (!userId) return;
  const value = text.trim();
  if (!value) return;
  add(userId, { type: 'query', text: value });
}

export function addRecentUser(
  userId: number | null | undefined,
  user: { userId: number; username: string; name: string; profile: string },
): void {
  if (!userId) return;
  add(userId, { type: 'user', userId: user.userId, username: user.username, name: user.name, profile: user.profile });
}

export function removeRecentSearch(userId: number | null | undefined, item: RecentSearch): void {
  if (!userId) return;
  const key = recentKey(item);
  write(userId, read(userId).filter((existing) => recentKey(existing) !== key));
}

export function clearRecentSearches(userId: number | null | undefined): void {
  if (!userId) return;
  write(userId, []);
}
