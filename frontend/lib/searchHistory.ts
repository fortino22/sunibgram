// Recent explore searches — device-local, no backend. SSR-safe.
// Holds two kinds of entries: typed queries and clicked users.

const STORAGE_KEY = 'sunibgram:recent-searches';
const MAX_ITEMS = 8;

export type RecentSearch =
  | { type: 'query'; text: string }
  | { type: 'user'; userId: number; username: string; name: string; profile: string };

/** Stable identity for dedupe / removal. */
export function recentKey(item: RecentSearch): string {
  return item.type === 'query' ? `q:${item.text.trim().toLowerCase()}` : `u:${item.userId}`;
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

function read(): RecentSearch[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.map(coerce).filter((v): v is RecentSearch => v !== null);
  } catch {
    return [];
  }
}

function write(items: RecentSearch[]): void {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
  } catch {
    // ignore quota / serialization errors
  }
}

/** Prepend, dedupe by identity, cap at MAX_ITEMS. */
function add(item: RecentSearch): void {
  const key = recentKey(item);
  const next = [item, ...read().filter((existing) => recentKey(existing) !== key)].slice(0, MAX_ITEMS);
  write(next);
}

export function getRecentSearches(): RecentSearch[] {
  return read();
}

export function addRecentQuery(text: string): void {
  const value = text.trim();
  if (!value) return;
  add({ type: 'query', text: value });
}

export function addRecentUser(user: { userId: number; username: string; name: string; profile: string }): void {
  add({ type: 'user', userId: user.userId, username: user.username, name: user.name, profile: user.profile });
}

export function removeRecentSearch(item: RecentSearch): void {
  const key = recentKey(item);
  write(read().filter((existing) => recentKey(existing) !== key));
}

export function clearRecentSearches(): void {
  write([]);
}
