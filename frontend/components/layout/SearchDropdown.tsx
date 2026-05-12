'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { X } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { Icon } from '@/components/ui/Icon';
import { useExploreSearch } from '@/hooks/useExplore';

export function SearchDropdown() {
  const router = useRouter();
  const [value, setValue] = useState('');
  const [open, setOpen] = useState(false);
  const [debouncedQ, setDebouncedQ] = useState('');

  useEffect(() => {
    const t = setTimeout(() => setDebouncedQ(value.trim()), 300);
    return () => clearTimeout(t);
  }, [value]);

  const { data, isLoading } = useExploreSearch(debouncedQ);
  const users = data?.users.slice(0, 6) ?? [];
  const showDropdown = open && value.trim().length > 0;

  function navigate(path: string) {
    setValue('');
    setOpen(false);
    router.push(path);
  }

  function goSearch(tab?: string) {
    const params = new URLSearchParams({ q: value.trim() });
    if (tab) params.set('tab', tab);
    navigate(`/explore?${params.toString()}`);
  }

  const inputRow: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    background: open ? 'var(--bg-card)' : 'var(--bg-subtle)',
    border: `1.5px solid ${open ? 'var(--accent)' : 'var(--border)'}`,
    borderRadius: showDropdown ? '12px 12px 0 0' : 9999,
    padding: '8px 14px',
    transition: 'border-radius 0s, border-color 0.15s, background 0.15s',
    boxShadow: open ? '0 0 0 3px oklch(64% 0.22 270 / 0.15)' : 'none',
  };

  const dropdownRow = (last = false): React.CSSProperties => ({
    display: 'flex',
    alignItems: 'center',
    gap: 10,
    width: '100%',
    padding: '11px 14px',
    background: 'transparent',
    border: 'none',
    borderBottom: last ? 'none' : '1px solid var(--border)',
    cursor: 'pointer',
    textAlign: 'left',
    fontFamily: 'inherit',
  });

  return (
    <div style={{ position: 'relative' }}>
      {/* Input */}
      <div style={inputRow}>
        <Icon name="search" size={15} color="var(--text-tertiary)" />
        <input
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onFocus={() => setOpen(true)}
          onBlur={() => setTimeout(() => setOpen(false), 150)}
          onKeyDown={(e) => { if (e.key === 'Enter' && value.trim()) goSearch(); }}
          placeholder="Explore campus..."
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

      {/* Dropdown */}
      {showDropdown && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: 0,
            right: 0,
            background: 'var(--bg-card)',
            border: '1.5px solid var(--accent)',
            borderTop: 'none',
            borderRadius: '0 0 12px 12px',
            boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
            zIndex: 50,
            overflow: 'hidden',
          }}
        >
          {/* Search for "query" */}
          <button
            onMouseDown={() => goSearch()}
            style={{ ...dropdownRow(false), borderBottom: '1px solid var(--border)' }}
            onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
          >
            <Icon name="search" size={14} color="var(--text-secondary)" />
            <span style={{ fontSize: 14, color: 'var(--text)' }}>
              Search for <strong>&ldquo;{value.trim()}&rdquo;</strong>
            </span>
          </button>

          {/* User results */}
          {isLoading && (
            <div style={{ padding: '12px 14px', fontSize: 13, color: 'var(--text-tertiary)' }}>
              Searching…
            </div>
          )}

          {!isLoading && users.map((user, i) => (
            <button
              key={user.userId}
              onMouseDown={() => navigate(`/profile/${user.userId}`)}
              style={dropdownRow(i === users.length - 1 && true)}
              onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
              onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
            >
              <Avatar username={user.username} avatarUrl={user.profile || null} size="sm" />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {user.name}
                </div>
                <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>@{user.username}</div>
              </div>
            </button>
          ))}

          {/* Go to @query */}
          <button
            onMouseDown={() => goSearch('people')}
            style={{ ...dropdownRow(true), borderTop: '1px solid var(--border)' }}
            onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-subtle)'; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
          >
            <span style={{ fontSize: 14, color: 'var(--text)' }}>
              Go to <strong>@{value.trim()}</strong>
            </span>
          </button>
        </div>
      )}
    </div>
  );
}
