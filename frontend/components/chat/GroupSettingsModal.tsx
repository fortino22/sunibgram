'use client';

import { useState, useEffect, useRef } from 'react';
import { X, Camera, Users, UserPlus, LogOut, Search, ChevronLeft } from 'lucide-react';
import { toast } from 'sonner';
import { compressImage, IMAGE_MAX_BYTES } from '@/lib/compressImage';
import { Avatar } from '@/components/shared/Avatar';
import { useUpdateGroupRoom, useAddGroupMembers, useLeaveGroup } from '@/hooks/chat/useCreateChatRoom';
import { useGroupMembers } from '@/hooks/chat/useGroupMembers';
import { useExplorePeople, useSuggestedPeople } from '@/hooks/useExplore';
import { useAuthStore } from '@/stores/authStore';
import type { ChatRoom, GroupMember } from '@/lib/services/chat';
import type { ExplorePerson } from '@/lib/services/explore';

interface GroupSettingsModalProps {
  room: ChatRoom;
  onClose: () => void;
  onLeave: () => void;
}

type Section = 'main' | 'addMembers';

export function GroupSettingsModal({ room, onClose, onLeave }: GroupSettingsModalProps) {
  const currentUser = useAuthStore((s) => s.user);

  // ── Info form state ──────────────────────────────────────────────────────
  const [name, setName] = useState(room.name ?? '');
  const [description, setDescription] = useState(room.description ?? '');
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(room.avatar_url ?? null);
  const avatarInputRef = useRef<HTMLInputElement>(null);

  // ── Section routing ──────────────────────────────────────────────────────
  const [section, setSection] = useState<Section>('main');

  // ── Add-member search state ──────────────────────────────────────────────
  const [addSearch, setAddSearch] = useState('');
  const [addDebounced, setAddDebounced] = useState('');
  const [toAdd, setToAdd] = useState<ExplorePerson[]>([]);

  // ── Leave confirmation ───────────────────────────────────────────────────
  const [confirmLeave, setConfirmLeave] = useState(false);

  const { data: members = [], isLoading: membersLoading } = useGroupMembers(room.room_id);
  const memberIdSet = new Set(members.map((m) => m.user_id));

  const update = useUpdateGroupRoom(room.room_id);
  const addMembers = useAddGroupMembers(room.room_id);
  const leave = useLeaveGroup(room.room_id);

  // 300 ms debounce for add-member search
  useEffect(() => {
    const t = setTimeout(() => setAddDebounced(addSearch.trim()), 300);
    return () => clearTimeout(t);
  }, [addSearch]);

  const hasQuery = addDebounced.length > 0;
  const searchQuery = useExplorePeople(addDebounced);
  const suggestedQuery = useSuggestedPeople(10);
  // Fall back to suggested data while search loads to avoid empty flash
  const searchPeople: ExplorePerson[] = hasQuery
    ? (searchQuery.data ?? suggestedQuery.data ?? [])
    : (suggestedQuery.data ?? []);
  // Filter out people already in the group and the current user
  const addCandidates = searchPeople.filter(
    (p) => !memberIdSet.has(p.userId) && p.userId !== currentUser?.id
  );
  const isAddLoading = hasQuery
    ? (searchQuery.isLoading && !suggestedQuery.data)
    : suggestedQuery.isLoading;

  const hasChanges =
    avatarFile !== null ||
    name.trim() !== (room.name ?? '') ||
    description.trim() !== (room.description ?? '');

  // ── Handlers ─────────────────────────────────────────────────────────────

  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const compressed = await compressImage(file, {
      maxDimension: 1200,
      maxSizeBytes: IMAGE_MAX_BYTES.avatar,
    });
    if (compressed.size > IMAGE_MAX_BYTES.avatar) {
      toast.error('Avatar image is too large even after compression. Try a smaller image.');
      return;
    }
    setAvatarFile(compressed);
    setAvatarPreview(URL.createObjectURL(compressed));
  };

  const handleSave = () => {
    if (!name.trim() || update.isPending) return;
    update.mutate(
      { name: name.trim(), description: description.trim() || undefined, avatar: avatarFile },
      { onSuccess: onClose }
    );
  };

  const toggleAddCandidate = (person: ExplorePerson) => {
    setToAdd((prev) => {
      const exists = prev.find((p) => p.userId === person.userId);
      if (exists) return prev.filter((p) => p.userId !== person.userId);
      return [...prev, person];
    });
  };

  const handleAddMembers = () => {
    if (toAdd.length === 0 || addMembers.isPending) return;
    addMembers.mutate(
      toAdd.map((p) => p.userId),
      {
        onSuccess: () => {
          setToAdd([]);
          setAddSearch('');
          setSection('main');
        },
      }
    );
  };

  const handleLeave = () => {
    if (leave.isPending) return;
    leave.mutate(undefined, { onSuccess: onLeave });
  };

  // ── Shared styles ─────────────────────────────────────────────────────────
  const inputStyle: React.CSSProperties = {
    width: '100%',
    background: 'var(--bg)',
    border: '1px solid var(--border)',
    borderRadius: 10,
    padding: '10px 14px',
    fontSize: 14,
    color: 'var(--text)',
    fontFamily: 'inherit',
    outline: 'none',
    boxSizing: 'border-box',
  };

  const sectionLabel: React.CSSProperties = {
    fontSize: 11,
    fontWeight: 600,
    color: 'var(--text-muted)',
    textTransform: 'uppercase',
    letterSpacing: '0.06em',
    marginBottom: 8,
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
          width: 480,
          maxWidth: 'calc(100vw - 32px)',
          maxHeight: '88vh',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* ── Header ── */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '14px 20px',
            borderBottom: '1px solid var(--border)',
            flexShrink: 0,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            {section === 'addMembers' && (
              <button
                onClick={() => { setSection('main'); setToAdd([]); setAddSearch(''); }}
                style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 4, display: 'flex', borderRadius: 6 }}
              >
                <ChevronLeft size={18} style={{ color: 'var(--text-muted)' }} />
              </button>
            )}
            <span style={{ fontWeight: 600, fontSize: 16, color: 'var(--text)' }}>
              {section === 'main' ? 'Group Settings' : 'Add Members'}
            </span>
          </div>
          <button
            onClick={onClose}
            style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 4, borderRadius: 6, display: 'flex' }}
          >
            <X size={18} style={{ color: 'var(--text-muted)' }} />
          </button>
        </div>

        {/* ── Body ── */}
        <div style={{ flex: 1, overflowY: 'auto' }}>

          {section === 'main' && (
            <>
              {/* ── Avatar + form ── */}
              <div style={{ padding: '20px 20px 0' }}>
                {/* Avatar picker */}
                <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 20 }}>
                  <div
                    style={{ position: 'relative', flexShrink: 0, cursor: 'pointer' }}
                    onClick={() => avatarInputRef.current?.click()}
                    title="Change group photo"
                  >
                    {avatarPreview ? (
                      <img
                        src={avatarPreview}
                        alt="group avatar"
                        style={{ width: 64, height: 64, borderRadius: '50%', objectFit: 'cover' }}
                      />
                    ) : (
                      <div
                        style={{
                          width: 64,
                          height: 64,
                          borderRadius: '50%',
                          background: 'var(--accent)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: 26,
                          color: '#fff',
                          fontWeight: 700,
                        }}
                      >
                        {(name || room.name || 'G').charAt(0).toUpperCase()}
                      </div>
                    )}
                    {/* Camera overlay */}
                    <div
                      style={{
                        position: 'absolute',
                        inset: 0,
                        borderRadius: '50%',
                        background: 'rgba(0,0,0,0.38)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        opacity: 0,
                        transition: 'opacity 0.15s',
                      }}
                      onMouseEnter={(e) => { e.currentTarget.style.opacity = '1'; }}
                      onMouseLeave={(e) => { e.currentTarget.style.opacity = '0'; }}
                    >
                      <Camera size={20} color="#fff" />
                    </div>
                    <input
                      ref={avatarInputRef}
                      type="file"
                      accept="image/jpeg,image/png,image/webp"
                      style={{ display: 'none' }}
                      onChange={handleAvatarChange}
                    />
                  </div>

                  <div>
                    <div style={{ fontWeight: 600, fontSize: 15, color: 'var(--text)' }}>
                      {name || room.name || 'Group Chat'}
                    </div>
                    <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 2 }}>
                      {room.member_count} members · Click photo to change
                    </div>
                  </div>
                </div>

                {/* Name */}
                <div style={{ marginBottom: 14 }}>
                  <div style={sectionLabel}>Group Name</div>
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    maxLength={60}
                    style={inputStyle}
                    onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--accent)'; }}
                    onBlur={(e) => { e.currentTarget.style.borderColor = 'var(--border)'; }}
                  />
                  <div style={{ fontSize: 11, color: 'var(--text-muted)', textAlign: 'right', marginTop: 3 }}>
                    {name.length}/60
                  </div>
                </div>

                {/* Description */}
                <div style={{ marginBottom: 16 }}>
                  <div style={sectionLabel}>
                    Description{' '}
                    <span style={{ fontWeight: 400, textTransform: 'none' }}>(optional)</span>
                  </div>
                  <textarea
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    maxLength={200}
                    rows={3}
                    placeholder="What's this group about?"
                    style={{ ...inputStyle, resize: 'vertical', lineHeight: 1.5 }}
                    onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--accent)'; }}
                    onBlur={(e) => { e.currentTarget.style.borderColor = 'var(--border)'; }}
                  />
                  <div style={{ fontSize: 11, color: 'var(--text-muted)', textAlign: 'right', marginTop: 3 }}>
                    {description.length}/200
                  </div>
                </div>

                <button
                  className="btn-solid"
                  onClick={handleSave}
                  disabled={!name.trim() || !hasChanges || update.isPending}
                  style={{
                    width: '100%',
                    padding: '11px',
                    borderRadius: 10,
                    background: !name.trim() || !hasChanges ? 'var(--border)' : 'var(--accent)',
                    border: 'none',
                    color: !name.trim() || !hasChanges ? 'var(--text-muted)' : '#fff',
                    fontWeight: 600,
                    fontSize: 14,
                    cursor: !name.trim() || !hasChanges || update.isPending ? 'not-allowed' : 'pointer',
                    fontFamily: 'inherit',
                    opacity: update.isPending ? 0.7 : 1,
                    transition: 'background 0.15s, filter 0.12s ease',
                    marginBottom: 20,
                  }}
                >
                  {update.isPending ? 'Saving…' : 'Save Changes'}
                </button>
              </div>

              <div style={{ height: 1, background: 'var(--border)' }} />

              {/* ── Members list ── */}
              <div style={{ padding: '16px 20px' }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                    <Users size={14} style={{ color: 'var(--text-muted)' }} />
                    <span style={sectionLabel as React.CSSProperties}>
                      Members ({room.member_count})
                    </span>
                  </div>
                  <button
                    onClick={() => setSection('addMembers')}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 4,
                      background: 'none',
                      border: '1px solid var(--border)',
                      borderRadius: 8,
                      padding: '4px 10px',
                      cursor: 'pointer',
                      fontSize: 12,
                      color: 'var(--accent)',
                      fontWeight: 600,
                      fontFamily: 'inherit',
                    }}
                  >
                    <UserPlus size={13} />
                    Add
                  </button>
                </div>

                {membersLoading ? (
                  Array.from({ length: 3 }, (_, i) => (
                    <div
                      key={i}
                      style={{
                        height: 48,
                        borderRadius: 8,
                        background: 'var(--skeleton, oklch(16% 0.005 260))',
                        marginBottom: 4,
                        animation: 'pulse 1.5s ease-in-out infinite',
                      }}
                    />
                  ))
                ) : (
                  members.map((m: GroupMember) => (
                    <MemberRow
                      key={m.user_id}
                      member={m}
                      isMe={m.user_id === currentUser?.id}
                    />
                  ))
                )}
              </div>

              <div style={{ height: 1, background: 'var(--border)' }} />

              {/* ── Leave group ── */}
              <div style={{ padding: '16px 20px 20px' }}>
                {confirmLeave ? (
                  <div>
                    <div style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 12, textAlign: 'center' }}>
                      Are you sure you want to leave this group?
                    </div>
                    <div style={{ display: 'flex', gap: 10 }}>
                      <button
                        onClick={() => setConfirmLeave(false)}
                        style={{
                          flex: 1,
                          padding: '10px',
                          borderRadius: 10,
                          background: 'var(--bg)',
                          border: '1px solid var(--border)',
                          color: 'var(--text)',
                          fontWeight: 600,
                          fontSize: 14,
                          cursor: 'pointer',
                          fontFamily: 'inherit',
                        }}
                      >
                        Cancel
                      </button>
                      <button
                        onClick={handleLeave}
                        disabled={leave.isPending}
                        style={{
                          flex: 1,
                          padding: '10px',
                          borderRadius: 10,
                          background: '#ef4444',
                          border: 'none',
                          color: '#fff',
                          fontWeight: 600,
                          fontSize: 14,
                          cursor: leave.isPending ? 'not-allowed' : 'pointer',
                          fontFamily: 'inherit',
                          opacity: leave.isPending ? 0.7 : 1,
                        }}
                      >
                        {leave.isPending ? 'Leaving…' : 'Leave Group'}
                      </button>
                    </div>
                  </div>
                ) : (
                  <button
                    onClick={() => setConfirmLeave(true)}
                    style={{
                      width: '100%',
                      padding: '11px',
                      borderRadius: 10,
                      background: 'none',
                      border: '1px solid #ef4444',
                      color: '#ef4444',
                      fontWeight: 600,
                      fontSize: 14,
                      cursor: 'pointer',
                      fontFamily: 'inherit',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      gap: 7,
                      transition: 'background 0.15s',
                    }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'rgba(239,68,68,0.08)'; }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'none'; }}
                  >
                    <LogOut size={15} />
                    Leave Group
                  </button>
                )}
              </div>
            </>
          )}

          {section === 'addMembers' && (
            <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
              {/* Selected chips */}
              {toAdd.length > 0 && (
                <div
                  style={{
                    padding: '10px 16px',
                    borderBottom: '1px solid var(--border)',
                    display: 'flex',
                    flexWrap: 'wrap',
                    gap: 6,
                  }}
                >
                  {toAdd.map((p) => (
                    <button
                      key={p.userId}
                      onClick={() => toggleAddCandidate(p)}
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
                      <Avatar username={p.username} avatarUrl={p.userProfilePic || null} size={20} />
                      <span>{p.name || p.username}</span>
                      <X size={11} style={{ marginLeft: 2, opacity: 0.8 }} />
                    </button>
                  ))}
                </div>
              )}

              {/* Search */}
              <div style={{ padding: '10px 16px', borderBottom: '1px solid var(--border)', flexShrink: 0 }}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 8,
                    background: 'var(--bg)',
                    border: '1px solid var(--border)',
                    borderRadius: 10,
                    padding: '8px 12px',
                  }}
                >
                  <Search size={14} style={{ color: 'var(--text-muted)' }} />
                  <input
                    type="text"
                    placeholder="Search by name or username…"
                    value={addSearch}
                    onChange={(e) => setAddSearch(e.target.value)}
                    autoFocus
                    style={{
                      flex: 1,
                      background: 'none',
                      border: 'none',
                      outline: 'none',
                      fontSize: 14,
                      color: 'var(--text)',
                      fontFamily: 'inherit',
                    }}
                  />
                </div>
              </div>

              {/* Candidate list */}
              <div style={{ flex: 1, overflowY: 'auto', padding: '4px 8px' }}>
                {isAddLoading ? (
                  Array.from({ length: 4 }, (_, i) => (
                    <div
                      key={i}
                      style={{
                        height: 56,
                        borderRadius: 10,
                        background: 'var(--skeleton, oklch(16% 0.005 260))',
                        marginBottom: 4,
                        animation: 'pulse 1.5s ease-in-out infinite',
                      }}
                    />
                  ))
                ) : addCandidates.length === 0 ? (
                  <div style={{ textAlign: 'center', padding: '32px 20px', color: 'var(--text-muted)', fontSize: 13 }}>
                    {hasQuery ? 'No users found' : 'No suggestions available'}
                  </div>
                ) : (
                  addCandidates.map((u) => {
                    const sel = toAdd.some((p) => p.userId === u.userId);
                    return (
                      <button
                        key={u.userId}
                        onClick={() => toggleAddCandidate(u)}
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
                        onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--sidebar-hover)'; }}
                        onMouseLeave={(e) => { e.currentTarget.style.background = 'none'; }}
                      >
                        <Avatar username={u.username} avatarUrl={u.userProfilePic || null} size={40} />
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div style={{ fontWeight: 500, fontSize: 14, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                            {u.name || u.username}
                          </div>
                          <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>@{u.username}</div>
                        </div>
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
                              <path d="M2.5 6.5l3 3 5-5" stroke="#fff" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                            </svg>
                          )}
                        </div>
                      </button>
                    );
                  })
                )}
              </div>

              {/* Add button */}
              <div style={{ padding: '12px 16px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
                <button
                  className="btn-solid"
                  onClick={handleAddMembers}
                  disabled={toAdd.length === 0 || addMembers.isPending}
                  style={{
                    width: '100%',
                    padding: '12px',
                    borderRadius: 12,
                    background: toAdd.length === 0 ? 'var(--border)' : 'var(--accent)',
                    border: 'none',
                    color: toAdd.length === 0 ? 'var(--text-muted)' : '#fff',
                    fontWeight: 600,
                    fontSize: 15,
                    cursor: toAdd.length === 0 || addMembers.isPending ? 'not-allowed' : 'pointer',
                    fontFamily: 'inherit',
                    opacity: addMembers.isPending ? 0.7 : 1,
                    transition: 'background 0.15s, filter 0.12s ease',
                  }}
                >
                  {addMembers.isPending
                    ? 'Adding…'
                    : toAdd.length > 0
                    ? `Add ${toAdd.length} member${toAdd.length > 1 ? 's' : ''}`
                    : 'Select people to add'}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function MemberRow({ member, isMe }: { member: GroupMember; isMe: boolean }) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '8px 0',
      }}
    >
      <Avatar
        username={member.username}
        avatarUrl={member.avatar_url || null}
        size={38}
      />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 500, fontSize: 14, color: 'var(--text)', display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {member.name || member.username}
          </span>
          {isMe && (
            <span style={{ fontSize: 10, color: 'var(--text-muted)', fontWeight: 400, flexShrink: 0 }}>You</span>
          )}
        </div>
        <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>@{member.username}</div>
      </div>
      {member.faculty_name && (
        <span style={{ fontSize: 11, color: 'var(--text-muted)', flexShrink: 0, maxWidth: 100, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {member.faculty_name}
        </span>
      )}
    </div>
  );
}
