'use client';

import { useEffect, useRef, useState } from 'react';
import { Search, X, LoaderCircle } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { searchInviteeCandidates, type InviteeCandidate } from '@/lib/services/communityEvents';
import { useAuthStore } from '@/stores/authStore';

interface Props {
  communityId: number;
  selected: InviteeCandidate[];
  onChange: (next: InviteeCandidate[]) => void;
}

export function InviteePicker({ communityId, selected, onChange }: Props) {
  const token = useAuthStore((s) => s.token);
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<InviteeCandidate[]>([]);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const boxRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function onClick(e: MouseEvent) {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) setOpen(false);
    }
    window.addEventListener('mousedown', onClick);
    return () => window.removeEventListener('mousedown', onClick);
  }, [open]);

  useEffect(() => {
    if (!token) return;
    const q = query.trim();
    if (q.length === 0) {
      setResults([]);
      return;
    }
    setLoading(true);
    const t = setTimeout(async () => {
      try {
        const data = await searchInviteeCandidates(communityId, q, token);
        setResults(data);
        setOpen(true);
      } catch {
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 250);
    return () => clearTimeout(t);
  }, [query, communityId, token]);

  const selectedIds = new Set(selected.map((s) => s.user_id));
  const visibleResults = results.filter((r) => !selectedIds.has(r.user_id));

  function add(candidate: InviteeCandidate) {
    onChange([...selected, candidate]);
    setQuery('');
    setResults([]);
    setOpen(false);
  }

  function remove(userId: number) {
    onChange(selected.filter((s) => s.user_id !== userId));
  }

  return (
    <div ref={boxRef} style={{ position: 'relative' }}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '0 12px',
          borderRadius: 10,
          border: '1px solid var(--border)',
          background: 'var(--bg-input, var(--bg-card))',
        }}
      >
        <Search size={16} color="var(--text-tertiary)" />
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => results.length > 0 && setOpen(true)}
          placeholder="Search people by name or @username"
          style={{
            flex: 1,
            border: 'none',
            outline: 'none',
            background: 'transparent',
            color: 'var(--text)',
            fontSize: 14,
            padding: '10px 0',
          }}
        />
        {loading && <LoaderCircle size={15} className="animate-spin" color="var(--text-tertiary)" />}
      </div>

      {open && visibleResults.length > 0 && (
        <div
          className="tab-overflow-menu"
          style={{
            position: 'absolute',
            top: 'calc(100% + 6px)',
            left: 0,
            right: 0,
            zIndex: 40,
            maxHeight: 240,
            overflowY: 'auto',
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            borderRadius: 12,
            boxShadow: 'var(--shadow-dropdown)',
            padding: 6,
          }}
        >
          {visibleResults.map((r) => (
            <button
              key={r.user_id}
              type="button"
              className="tab-overflow-item"
              onClick={() => add(r)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                width: '100%',
                textAlign: 'left',
                padding: '8px 10px',
                background: 'none',
                border: 'none',
                borderRadius: 8,
                cursor: 'pointer',
              }}
            >
              <Avatar username={r.username} avatarUrl={r.avatar_url || null} size={32} />
              <div style={{ minWidth: 0 }}>
                <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)' }}>{r.name}</div>
                <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>@{r.username}</div>
              </div>
            </button>
          ))}
        </div>
      )}

      {selected.length > 0 && (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginTop: 10 }}>
          {selected.map((s) => (
            <span
              key={s.user_id}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 6,
                padding: '4px 8px 4px 4px',
                borderRadius: 999,
                background: 'var(--bg-subtle)',
                fontSize: 13,
                color: 'var(--text)',
              }}
            >
              <Avatar username={s.username} avatarUrl={s.avatar_url || null} size={22} />
              {s.name}
              <button
                type="button"
                onClick={() => remove(s.user_id)}
                aria-label={`Remove ${s.name}`}
                style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-tertiary)', display: 'flex' }}
              >
                <X size={13} />
              </button>
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
