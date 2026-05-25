'use client';

import { useState, useEffect } from 'react';
import { X, ChevronLeft } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useExplorePeople, useSuggestedPeople } from '@/hooks/useExplore';
import { useCreatePrivateRoom, useCreateGroupRoom } from '@/hooks/chat/useCreateChatRoom';
import { useAuthStore } from '@/stores/authStore';
import type { ExplorePerson } from '@/lib/services/explore';

interface NewChatModalProps {
  onClose: () => void;
  onRoomCreated: (roomId: number) => void;
}

export function NewChatModal({ onClose, onRoomCreated }: NewChatModalProps) {
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [selected, setSelected] = useState<ExplorePerson[]>([]);
  const [step, setStep] = useState<'select' | 'groupName'>('select');
  const [groupName, setGroupName] = useState('');

  const currentUser = useAuthStore((s) => s.user);
  const createPrivate = useCreatePrivateRoom();
  const createGroup = useCreateGroupRoom();

  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search.trim()), 300);
    return () => clearTimeout(t);
  }, [search]);

  const hasQuery = debouncedSearch.length > 0;
  const searchQuery = useExplorePeople(debouncedSearch);
  const suggestedQuery = useSuggestedPeople(10);

  // Keep suggested data visible while first search is loading to avoid empty flash
  const rawPeople: ExplorePerson[] = hasQuery
    ? (searchQuery.data ?? suggestedQuery.data ?? [])
    : (suggestedQuery.data ?? []);
  const people = rawPeople.filter((p) => p.userId !== currentUser?.id);
  // Only show skeleton when there's truly no data to display at all
  const showSkeleton = hasQuery
    ? (searchQuery.isLoading && !suggestedQuery.data)
    : suggestedQuery.isLoading;

  const isSelected = (userId: number) => selected.some((p) => p.userId === userId);

  const toggleSelect = (person: ExplorePerson) => {
    setSelected((prev) => {
      const exists = prev.find((p) => p.userId === person.userId);
      if (exists) return prev.filter((p) => p.userId !== person.userId);
      return [...prev, person];
    });
  };

  const isPending = createPrivate.isPending || createGroup.isPending;

  const handleChat = () => {
    if (selected.length === 0 || isPending) return;
    if (selected.length === 1) {
      createPrivate.mutate(selected[0].userId, {
        onSuccess: (data) => onRoomCreated(data.room_id),
      });
    } else {
      setStep('groupName');
    }
  };

  const handleCreateGroup = () => {
    if (!groupName.trim() || selected.length < 2 || isPending) return;
    createGroup.mutate(
      { name: groupName.trim(), memberIds: selected.map((p) => p.userId) },
      { onSuccess: (data) => onRoomCreated(data.room_id) }
    );
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 100,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          width: 560,
          maxWidth: 'calc(100vw - 32px)',
          maxHeight: '80vh',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '16px 20px',
            borderBottom: '1px solid var(--border)',
            flexShrink: 0,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            {step === 'groupName' && (
              <button
                onClick={() => setStep('select')}
                style={{
                  background: 'none',
                  border: 'none',
                  cursor: 'pointer',
                  padding: 4,
                  borderRadius: 6,
                  display: 'flex',
                }}
              >
                <ChevronLeft size={18} style={{ color: 'var(--text-muted)' }} />
              </button>
            )}
            <span style={{ fontWeight: 600, fontSize: 16, color: 'var(--text)' }}>
              {step === 'select' ? 'New message' : 'New group'}
            </span>
          </div>
          <button
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              padding: 4,
              borderRadius: 6,
              display: 'flex',
            }}
          >
            <X size={18} style={{ color: 'var(--text-muted)' }} />
          </button>
        </div>

        {step === 'select' ? (
          <>
            {/* To: row — chips + inline search */}
            <div
              style={{
                padding: '10px 16px',
                borderBottom: '1px solid var(--border)',
                flexShrink: 0,
                display: 'flex',
                alignItems: 'flex-start',
                flexWrap: 'wrap',
                gap: 6,
                minHeight: 48,
              }}
            >
              <span
                style={{
                  fontSize: 14,
                  color: 'var(--text-muted)',
                  paddingTop: 5,
                  flexShrink: 0,
                  lineHeight: 1,
                }}
              >
                To:
              </span>

              {selected.map((p) => (
                <button
                  key={p.userId}
                  onClick={() => toggleSelect(p)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 4,
                    background: 'var(--accent)',
                    color: '#fff',
                    border: 'none',
                    borderRadius: 20,
                    padding: '3px 8px 3px 4px',
                    fontSize: 13,
                    fontWeight: 500,
                    cursor: 'pointer',
                  }}
                >
                  <Avatar
                    username={p.username}
                    avatarUrl={p.userProfilePic || null}
                    size={20}
                  />
                  <span>{p.name || p.username}</span>
                  <X size={11} style={{ marginLeft: 2, opacity: 0.8 }} />
                </button>
              ))}

              <input
                type="text"
                placeholder={selected.length === 0 ? 'Search…' : ''}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                autoFocus
                style={{
                  flex: 1,
                  minWidth: 80,
                  background: 'none',
                  border: 'none',
                  outline: 'none',
                  fontSize: 14,
                  color: 'var(--text)',
                  fontFamily: 'inherit',
                  paddingTop: 4,
                }}
              />
            </div>

            {/* Section label */}
            <div
              style={{
                padding: '8px 20px 4px',
                fontSize: 11,
                fontWeight: 600,
                color: 'var(--text-muted)',
                textTransform: 'uppercase',
                letterSpacing: '0.06em',
                flexShrink: 0,
              }}
            >
              {hasQuery ? 'Results' : 'Suggested'}
            </div>

            {/* User list */}
            <div style={{ flex: 1, overflowY: 'auto', padding: '4px 8px 8px' }}>
              {showSkeleton ? (
                Array.from({ length: 5 }, (_, i) => (
                  <div
                    key={i}
                    style={{
                      height: 60,
                      borderRadius: 10,
                      background: 'var(--skeleton, oklch(16% 0.005 260))',
                      marginBottom: 4,
                      animation: 'pulse 1.5s ease-in-out infinite',
                    }}
                  />
                ))
              ) : people.length === 0 ? (
                <div
                  style={{
                    textAlign: 'center',
                    padding: '32px 20px',
                    color: 'var(--text-muted)',
                    fontSize: 13,
                  }}
                >
                  {hasQuery ? 'No users found' : 'No suggestions available'}
                </div>
              ) : (
                people.map((u) => {
                  const sel = isSelected(u.userId);
                  return (
                    <button
                      key={u.userId}
                      onClick={() => toggleSelect(u)}
                      style={{
                        width: '100%',
                        display: 'flex',
                        alignItems: 'center',
                        gap: 12,
                        padding: '10px 12px',
                        background: 'none',
                        border: 'none',
                        borderRadius: 10,
                        cursor: 'pointer',
                        textAlign: 'left',
                        transition: 'background 0.1s',
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.background = 'var(--sidebar-hover)';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.background = 'none';
                      }}
                    >
                      <Avatar
                        username={u.username}
                        avatarUrl={u.userProfilePic || null}
                        size={42}
                      />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div
                          style={{
                            fontWeight: 500,
                            fontSize: 14,
                            color: 'var(--text)',
                            display: 'flex',
                            alignItems: 'center',
                            gap: 4,
                          }}
                        >
                          <span
                            style={{
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap',
                            }}
                          >
                            {u.name || u.username}
                          </span>
                          {u.isPremium && <VerifiedBadge size={15} />}
                        </div>
                        <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
                          @{u.username}
                        </div>
                      </div>

                      {/* Circular checkbox */}
                      <div
                        style={{
                          width: 24,
                          height: 24,
                          borderRadius: '50%',
                          border: sel ? 'none' : '2px solid var(--border)',
                          background: sel ? 'var(--accent)' : 'transparent',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          flexShrink: 0,
                          transition: 'all 0.15s',
                        }}
                      >
                        {sel && (
                          <svg width="13" height="13" viewBox="0 0 13 13" fill="none">
                            <path
                              d="M2.5 6.5l3 3 5-5"
                              stroke="#fff"
                              strokeWidth="2"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />
                          </svg>
                        )}
                      </div>
                    </button>
                  );
                })
              )}
            </div>

            {/* Bottom Chat / Next button */}
            <div
              style={{
                padding: '12px 16px',
                borderTop: '1px solid var(--border)',
                flexShrink: 0,
              }}
            >
              <button
                onClick={handleChat}
                disabled={selected.length === 0 || isPending}
                style={{
                  width: '100%',
                  padding: '13px',
                  borderRadius: 12,
                  background:
                    selected.length === 0 ? 'var(--border)' : 'var(--accent)',
                  border: 'none',
                  color: selected.length === 0 ? 'var(--text-muted)' : '#fff',
                  fontWeight: 600,
                  fontSize: 15,
                  cursor: selected.length === 0 || isPending ? 'not-allowed' : 'pointer',
                  transition: 'background 0.15s',
                  opacity: isPending ? 0.7 : 1,
                }}
              >
                {isPending
                  ? 'Creating…'
                  : selected.length > 1
                  ? 'Next'
                  : 'Chat'}
              </button>
            </div>
          </>
        ) : (
          /* Step 2: Group name */
          <>
            <div
              style={{
                padding: '20px',
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                gap: 20,
              }}
            >
              {/* Member preview */}
              <div>
                <div
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    color: 'var(--text-muted)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.06em',
                    marginBottom: 10,
                  }}
                >
                  Members ({selected.length})
                </div>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                  {selected.map((p) => (
                    <div
                      key={p.userId}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 6,
                        background: 'var(--bg)',
                        border: '1px solid var(--border)',
                        borderRadius: 20,
                        padding: '4px 10px 4px 4px',
                      }}
                    >
                      <Avatar
                        username={p.username}
                        avatarUrl={p.userProfilePic || null}
                        size={24}
                      />
                      <span
                        style={{
                          fontSize: 13,
                          color: 'var(--text)',
                          fontWeight: 500,
                        }}
                      >
                        {p.name || p.username}
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Group name input */}
              <div>
                <div
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    color: 'var(--text-muted)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.06em',
                    marginBottom: 8,
                  }}
                >
                  Group Name
                </div>
                <input
                  type="text"
                  placeholder="e.g. Study Group, Project Team…"
                  value={groupName}
                  onChange={(e) => setGroupName(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') handleCreateGroup(); }}
                  autoFocus
                  maxLength={60}
                  style={{
                    width: '100%',
                    background: 'var(--bg)',
                    border: '1px solid var(--border)',
                    borderRadius: 10,
                    padding: '11px 14px',
                    fontSize: 14,
                    color: 'var(--text)',
                    fontFamily: 'inherit',
                    outline: 'none',
                    boxSizing: 'border-box',
                  }}
                  onFocus={(e) => {
                    e.currentTarget.style.borderColor = 'var(--accent)';
                  }}
                  onBlur={(e) => {
                    e.currentTarget.style.borderColor = 'var(--border)';
                  }}
                />
                <div
                  style={{
                    fontSize: 12,
                    color: 'var(--text-muted)',
                    textAlign: 'right',
                    marginTop: 4,
                  }}
                >
                  {groupName.length}/60
                </div>
              </div>
            </div>

            <div
              style={{
                padding: '12px 16px',
                borderTop: '1px solid var(--border)',
                flexShrink: 0,
              }}
            >
              <button
                onClick={handleCreateGroup}
                disabled={!groupName.trim() || isPending}
                style={{
                  width: '100%',
                  padding: '13px',
                  borderRadius: 12,
                  background: !groupName.trim() ? 'var(--border)' : 'var(--accent)',
                  border: 'none',
                  color: !groupName.trim() ? 'var(--text-muted)' : '#fff',
                  fontWeight: 600,
                  fontSize: 15,
                  cursor: !groupName.trim() || isPending ? 'not-allowed' : 'pointer',
                  transition: 'background 0.15s',
                  opacity: isPending ? 0.7 : 1,
                }}
              >
                {isPending ? 'Creating…' : 'Create Group'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
