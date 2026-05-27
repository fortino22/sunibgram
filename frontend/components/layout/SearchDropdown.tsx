'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Search, X } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { useExploreSearch } from '@/hooks/useExplore';
import {
  getRecentSearches,
  addRecentQuery,
  addRecentUser,
  removeRecentSearch,
  clearRecentSearches,
  recentKey,
  type RecentSearch,
} from '@/lib/searchHistory';

export function SearchDropdown() {
  const router = useRouter();
  const [value, setValue] = useState('');
  const [open, setOpen] = useState(false);
  const [debouncedQ, setDebouncedQ] = useState('');
  const [recent, setRecent] = useState<RecentSearch[]>([]);

  useEffect(() => {
    const t = setTimeout(() => setDebouncedQ(value.trim()), 300);
    return () => clearTimeout(t);
  }, [value]);

  const { data, isLoading } = useExploreSearch(debouncedQ);
  const users = data?.users.slice(0, 6) ?? [];

  const trimmed = value.trim();
  const showResults = open && trimmed.length > 0;
  const showRecent = open && trimmed.length === 0 && recent.length > 0;
  const showDropdown = showResults || showRecent;

  function refreshRecent() {
    setRecent(getRecentSearches());
  }

  function navigate(path: string) {
    setValue('');
    setOpen(false);
    router.push(path);
  }

  function goSearch(q: string, tab?: string) {
    const v = q.trim();
    if (!v) return;
    addRecentQuery(v);
    const params = new URLSearchParams({ q: v });
    if (tab) params.set('tab', tab);
    navigate(`/explore?${params.toString()}`);
  }

  function openUser(user: { userId: number; username: string; name: string; profile: string }) {
    addRecentUser(user);
    navigate(`/profile/${user.userId}`);
  }

  // Same look as the explore search bar — full width (no shrink), full pill.
  const inputRow: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    width: '100%',
    background: 'var(--bg-secondary)',
    border: open ? '1.5px solid var(--accent)' : '1px solid var(--border)',
    borderRadius: 999,
    padding: open ? '7.5px 13.5px' : '8px 14px',
    transition: 'border-color 0.15s',
  };

  const row: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: 12,
    width: '100%',
    padding: '11px 16px',
    background: 'transparent',
    border: 'none',
    cursor: 'pointer',
    textAlign: 'left',
    fontFamily: 'inherit',
  };

  const removeBtn: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    background: 'none',
    border: 'none',
    cursor: 'pointer',
    color: 'var(--accent)',
    flexShrink: 0,
    padding: 4,
    marginRight: -4,
  };

  return (
    <div style={{ position: 'relative' }}>
      {/* Input — full-width pill */}
      <div style={inputRow}>
        <Search size={15} style={{ color: 'var(--text-tertiary)', flexShrink: 0 }} />
        <input
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onFocus={() => { refreshRecent(); setOpen(true); }}
          onBlur={() => setTimeout(() => setOpen(false), 150)}
          onKeyDown={(e) => { if (e.key === 'Enter' && trimmed) goSearch(value); }}
          placeholder="Search"
          style={{
            border: 'none',
            background: 'transparent',
            outline: 'none',
            fontSize: 14,
            color: 'var(--text)',
            width: '100%',
            fontFamily: 'inherit',
          }}
        />
        {value && (
          <button
            type="button"
            onMouseDown={(e) => { e.preventDefault(); setValue(''); }}
            aria-label="Clear"
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: 18,
              height: 18,
              borderRadius: '50%',
              background: 'var(--text-tertiary)',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--bg)',
              flexShrink: 0,
            }}
          >
            <X size={10} />
          </button>
        )}
      </div>

      {/* Floating dropdown card — detached, full width of the panel */}
      {showDropdown && (
        <div
          style={{
            position: 'absolute',
            top: 'calc(100% + 6px)',
            left: 0,
            right: 0,
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 16,
            boxShadow: '0 8px 28px rgba(0,0,0,0.45)',
            zIndex: 50,
            overflow: 'hidden',
            paddingBottom: 6,
          }}
        >
          {showRecent && (
            <>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 16px 6px' }}>
                <span style={{ fontSize: 17, fontWeight: 800, color: 'var(--text)' }}>Recent</span>
                <button
                  type="button"
                  onMouseDown={(e) => { e.preventDefault(); clearRecentSearches(); refreshRecent(); }}
                  style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: 14, fontWeight: 600, color: 'var(--accent)', fontFamily: 'inherit', padding: 0 }}
                >
                  Clear all
                </button>
              </div>

              {recent.map((item) => (
                <div
                  key={recentKey(item)}
                  role="button"
                  tabIndex={0}
                  onMouseDown={() => (item.type === 'query' ? goSearch(item.text) : openUser(item))}
                  style={row}
                  onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
                  onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
                >
                  {item.type === 'query' ? (
                    <>
                      <Search size={17} style={{ color: 'var(--text-tertiary)', flexShrink: 0 }} />
                      <span style={{ flex: 1, minWidth: 0, fontSize: 15, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {item.text}
                      </span>
                    </>
                  ) : (
                    <>
                      <Avatar username={item.username} avatarUrl={item.profile || null} size="sm" />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontSize: 15, fontWeight: 600, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {item.name}
                        </div>
                        <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@{item.username}</div>
                      </div>
                    </>
                  )}
                  <button
                    type="button"
                    aria-label="Remove"
                    onMouseDown={(e) => { e.preventDefault(); e.stopPropagation(); removeRecentSearch(item); refreshRecent(); }}
                    style={removeBtn}
                  >
                    <X size={17} />
                  </button>
                </div>
              ))}
            </>
          )}

          {showResults && (
            <>
              {/* Search for "query" */}
              <button
                type="button"
                onMouseDown={() => goSearch(value)}
                style={row}
                onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
                onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
              >
                <Search size={16} style={{ color: 'var(--text-secondary)', flexShrink: 0 }} />
                <span style={{ fontSize: 15, color: 'var(--text)' }}>
                  Search for <strong>&ldquo;{trimmed}&rdquo;</strong>
                </span>
              </button>

              {isLoading && (
                <div style={{ padding: '12px 16px', fontSize: 14, color: 'var(--text-tertiary)' }}>
                  Searching…
                </div>
              )}

              {!isLoading && users.map((user) => (
                <button
                  key={user.userId}
                  type="button"
                  onMouseDown={() => openUser({ userId: user.userId, username: user.username, name: user.name, profile: user.profile })}
                  style={row}
                  onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
                  onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
                >
                  <Avatar username={user.username} avatarUrl={user.profile || null} size="sm" />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 15, fontWeight: 600, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {user.name}
                    </div>
                    <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@{user.username}</div>
                  </div>
                </button>
              ))}

              {/* Go to @query */}
              <button
                type="button"
                onMouseDown={() => goSearch(value, 'people')}
                style={row}
                onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
                onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
              >
                <span style={{ fontSize: 15, color: 'var(--text)', paddingLeft: 28 }}>
                  Go to <strong>@{trimmed}</strong>
                </span>
              </button>
            </>
          )}
        </div>
      )}
    </div>
  );
}
