'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import dynamic from 'next/dynamic';
import { useRouter } from 'next/navigation';
import { ArrowLeft, CalendarDays, ChevronDown, MapPin, MoreHorizontal, Pencil, Pin, PinOff, Plus, Trash2, Users } from 'lucide-react';
import { toast } from 'sonner';
import { Avatar } from '@/components/shared/Avatar';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import { EmptyState } from '@/components/shared/EmptyState';
import { DangerConfirmModal } from '@/components/shared/DangerConfirmModal';
import { renderRichContent } from '@/components/thread/MentionText';
import { CreateCommunityEventDialog } from './CreateCommunityEventDialog';
import { MentionHighlightTextarea } from './MentionHighlightTextarea';
import { MentionAutocomplete } from '@/components/thread/MentionAutocomplete';

const EmojiPicker = dynamic(() => import('emoji-picker-react'), { ssr: false });
import { useCommunityEvent, useCommunityEventComments } from '@/hooks/useCommunityEvents';
import {
  useAddCommunityEventComment,
  useEditCommunityEventComment,
  useDeleteCommunityEventComment,
  usePinCommunityEventComment,
  useSetCommunityEventReaction,
  useDeleteCommunityEvent,
} from '@/hooks/useCommunityEventMutations';
import type { MentionUser } from '@/lib/services/threads';
import { encodeMentions, decodeMentionsForEditing } from '@/lib/mentions';
import { useAuthStore } from '@/stores/authStore';
import { useMyProfile } from '@/hooks/useMyProfile';
import type {
  CommunityEventCommentData,
  CommunityEventDetail as CommunityEventDetailData,
} from '@/lib/services/communityEvents';

const REACTIONS = ['👍', '❤️', '🎉', '😂', '😮', '😢'];

// Backend stores UTC as "YYYY-MM-DD HH:mm:ss" with no zone — append "Z" so it's
// parsed as UTC, then render a short X-style relative label ("5m", "2h", "3d").
function commentTimeAgo(value: string): string {
  if (!value) return '';
  const d = new Date(value.replace(' ', 'T') + 'Z');
  if (Number.isNaN(d.getTime())) return '';
  const sec = Math.max(0, Math.floor((Date.now() - d.getTime()) / 1000));
  if (sec < 60) return 'now';
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h`;
  const day = Math.floor(hr / 24);
  if (day < 7) return `${day}d`;
  const wk = Math.floor(day / 7);
  if (wk < 5) return `${wk}w`;
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

function formatRange(start: string, end: string): string {
  const s = new Date(start.replace(' ', 'T') + 'Z');
  const e = new Date(end.replace(' ', 'T') + 'Z');
  if (Number.isNaN(s.getTime())) return '';
  const dt: Intl.DateTimeFormatOptions = { weekday: 'short', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' };
  const t: Intl.DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit' };
  const sameDay = s.toDateString() === e.toDateString();
  return sameDay
    ? `${s.toLocaleString(undefined, dt)} – ${e.toLocaleTimeString(undefined, t)}`
    : `${s.toLocaleString(undefined, dt)} → ${e.toLocaleString(undefined, dt)}`;
}

function ReactionBar({ event, eventId }: { event: CommunityEventDetailData; eventId: number }) {
  const setReaction = useSetCommunityEventReaction(eventId);
  const [pickerOpen, setPickerOpen] = useState(false);
  const pickerRef = useRef<HTMLDivElement>(null);

  const counts = useMemo(() => {
    const map = new Map<string, number>();
    event.reaction_summary.forEach((r) => map.set(r.emoji, r.count));
    return map;
  }, [event.reaction_summary]);

  // Default set + any emoji already in play (others' picks + the viewer's own).
  const emojis = useMemo(() => {
    const set = new Set<string>(REACTIONS);
    event.reaction_summary.forEach((r) => set.add(r.emoji));
    if (event.my_reaction) set.add(event.my_reaction);
    return Array.from(set);
  }, [event.reaction_summary, event.my_reaction]);

  useEffect(() => {
    if (!pickerOpen) return;
    const handler = (e: MouseEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(e.target as Node)) setPickerOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [pickerOpen]);

  return (
    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, alignItems: 'center' }}>
      {emojis.map((emoji) => {
        const active = event.my_reaction === emoji;
        const count = counts.get(emoji) ?? 0;
        return (
          <button
            key={emoji}
            className="ce-chip"
            onClick={() => setReaction.mutate(emoji)}
            disabled={setReaction.isPending}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '6px 12px',
              borderRadius: 999,
              border: `1px solid ${active ? 'var(--text)' : 'var(--border)'}`,
              background: active ? 'var(--bg-subtle)' : 'transparent',
              cursor: 'pointer',
              fontSize: 15,
            }}
            aria-pressed={active}
          >
            <span>{emoji}</span>
            {count > 0 && <span style={{ fontSize: 13, fontWeight: active ? 700 : 400, color: 'var(--text-secondary)' }}>{count}</span>}
          </button>
        );
      })}

      {/* Pick more */}
      <div ref={pickerRef} style={{ position: 'relative' }}>
        <button
          type="button"
          className="ce-chip"
          onClick={() => setPickerOpen((v) => !v)}
          aria-label="Pick another emoji"
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: 34,
            height: 34,
            borderRadius: 999,
            border: '1px solid var(--border)',
            background: pickerOpen ? 'var(--bg-subtle)' : 'transparent',
            color: 'var(--text-secondary)',
            cursor: 'pointer',
          }}
        >
          <Plus size={16} />
        </button>
        {pickerOpen && (
          <div style={{ position: 'absolute', top: 'calc(100% + 6px)', left: 0, zIndex: 50 }}>
            <EmojiPicker
              onEmojiClick={(data) => { setReaction.mutate(data.emoji); setPickerOpen(false); }}
              skinTonesDisabled
              height={340}
              width={300}
            />
          </div>
        )}
      </div>
    </div>
  );
}

function MenuItem({
  icon,
  label,
  onClick,
  danger,
}: {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  danger?: boolean;
}) {
  return (
    <button
      type="button"
      role="menuitem"
      onClick={onClick}
      onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-hover)'; }}
      onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        width: '100%',
        padding: '9px 14px',
        border: 'none',
        background: 'transparent',
        cursor: 'pointer',
        fontSize: 14,
        fontWeight: 500,
        textAlign: 'left',
        color: danger ? 'var(--coral)' : 'var(--text)',
      }}
    >
      {icon}
      {label}
    </button>
  );
}

function CommentRow({
  comment,
  eventId,
  isAdmin,
  currentUserId,
}: {
  comment: CommunityEventCommentData;
  eventId: number;
  isAdmin: boolean;
  currentUserId?: number;
}) {
  const pinMutation = usePinCommunityEventComment(eventId);
  const deleteMutation = useDeleteCommunityEventComment(eventId);
  const isAuthor = comment.author.user_id === currentUserId;
  const canEdit = isAdmin || isAuthor;
  const canDelete = isAdmin || isAuthor;
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return undefined;
    function onDown(e: MouseEvent | TouchEvent) {
      const target = e.target;
      if (!(target instanceof Node) || menuRef.current?.contains(target)) return;
      setMenuOpen(false);
    }
    document.addEventListener('mousedown', onDown);
    document.addEventListener('touchstart', onDown);
    return () => {
      document.removeEventListener('mousedown', onDown);
      document.removeEventListener('touchstart', onDown);
    };
  }, [menuOpen]);

  return (
    <div
      className="ce-comment"
      style={{
        display: 'flex',
        gap: 10,
        padding: 12,
        borderRadius: 12,
        // Pinned rows stay highlighted; non-pinned rely on .ce-comment:hover.
        background: comment.is_pinned ? 'var(--bg-subtle)' : undefined,
        border: comment.is_pinned ? '1px solid var(--border)' : '1px solid transparent',
      }}
    >
      <Avatar username={comment.author.username} avatarUrl={comment.author.avatar_url || null} size="sm" />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
          <span style={{ fontWeight: 600, fontSize: 14, color: 'var(--text)' }}>{comment.author.name}</span>
          <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@{comment.author.username}</span>
          {comment.created_at && (
            <>
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>·</span>
              <span
                style={{ fontSize: 13, color: 'var(--text-tertiary)' }}
                title={new Date(comment.created_at.replace(' ', 'T') + 'Z').toLocaleString()}
              >
                {commentTimeAgo(comment.created_at)}
              </span>
            </>
          )}
          {comment.is_pinned && (
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 3, fontSize: 11, fontWeight: 600, color: 'var(--text-secondary)' }}>
              <Pin size={11} /> Pinned
            </span>
          )}
        </div>
        <p style={{ margin: '4px 0 0', fontSize: 14, color: 'var(--text)', lineHeight: 1.5, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
          {renderRichContent(comment.content)}
        </p>
      </div>
      {canEdit && (
        <div className="ce-actions" ref={menuRef} style={{ position: 'relative', flexShrink: 0, opacity: menuOpen ? 1 : undefined }}>
          <button
            className="ce-icon-btn"
            onClick={() => setMenuOpen((v) => !v)}
            aria-label="Comment actions"
            aria-expanded={menuOpen}
            style={{ border: 'none', cursor: 'pointer', color: 'var(--text-tertiary)', padding: 6, borderRadius: 8 }}
          >
            <MoreHorizontal size={16} />
          </button>
          {menuOpen && (
            <div
              role="menu"
              style={{
                position: 'absolute',
                right: 0,
                top: '100%',
                zIndex: 50,
                marginTop: 4,
                minWidth: 170,
                borderRadius: 12,
                border: '1px solid var(--border)',
                background: 'var(--bg-card)',
                boxShadow: '0 10px 28px rgba(0,0,0,0.32)',
                padding: '6px 0',
                overflow: 'hidden',
              }}
            >
              {isAdmin && (
                <MenuItem
                  icon={comment.is_pinned ? <PinOff size={16} /> : <Pin size={16} />}
                  label={comment.is_pinned ? 'Unpin' : 'Pin'}
                  onClick={() => {
                    pinMutation.mutate({ commentId: comment.id, pinned: comment.is_pinned });
                    setMenuOpen(false);
                  }}
                />
              )}
              {canEdit && (
                <MenuItem
                  icon={<Pencil size={16} />}
                  label="Update"
                  onClick={() => { setEditOpen(true); setMenuOpen(false); }}
                />
              )}
              {canDelete && (
                <MenuItem
                  icon={<Trash2 size={16} />}
                  label="Delete"
                  danger
                  onClick={() => { setConfirmDelete(true); setMenuOpen(false); }}
                />
              )}
            </div>
          )}
        </div>
      )}

      <DangerConfirmModal
        open={confirmDelete}
        onClose={() => setConfirmDelete(false)}
        title="Delete comment?"
        description="This permanently removes this comment."
        confirmLabel="Delete"
        busy={deleteMutation.isPending}
        onConfirm={() => deleteMutation.mutate(comment.id, { onSuccess: () => setConfirmDelete(false) })}
      />
      <EditCommentModal eventId={eventId} comment={comment} open={editOpen} onClose={() => setEditOpen(false)} />
    </div>
  );
}

function EveryoneRow({ onSelect }: { onSelect: () => void }) {
  return (
    <button
      type="button"
      className="ce-everyone-row"
      onMouseDown={(e) => { e.preventDefault(); onSelect(); }}
      style={{ display: 'flex', alignItems: 'center', gap: 10, width: '100%', padding: '8px 10px', borderRadius: 10, border: 'none', background: 'transparent', cursor: 'pointer', textAlign: 'left' }}
    >
      <span style={{ width: 32, height: 32, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'color-mix(in oklab, var(--accent) 14%, transparent)', color: 'var(--accent)', flexShrink: 0 }}>
        <Users size={16} />
      </span>
      <div style={{ display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        <span style={{ fontWeight: 700, fontSize: 14, color: 'var(--text)' }}>Everyone</span>
        <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@all · notifies every tagged person</span>
      </div>
    </button>
  );
}

// Reusable mention-aware comment editor. Used both for adding a new comment and
// for editing an existing one (inside the edit modal) so the mention/caret logic
// lives in one place.
function CommentField({
  initial,
  submitLabel = 'Comment',
  pending,
  onSubmit,
  onCancel,
  resetAfterSubmit,
  autoFocus,
}: {
  initial?: { text: string; mentionMap: Map<string, number> };
  submitLabel?: string;
  pending: boolean;
  onSubmit: (content: string) => Promise<void>;
  onCancel?: () => void;
  resetAfterSubmit?: boolean;
  autoFocus?: boolean;
}) {
  const [text, setText] = useState(initial?.text ?? '');
  const [mentionState, setMentionState] = useState<{ query: string; start: number; end: number } | null>(null);
  const [mentionIndex, setMentionIndex] = useState(0);
  const mentionMap = useRef<Map<string, number>>(new Map(initial?.mentionMap ?? []));
  const mentionResultsRef = useRef<MentionUser[]>([]);
  const taRef = useRef<HTMLTextAreaElement>(null);
  const [taEl, setTaEl] = useState<HTMLTextAreaElement | null>(null);

  useEffect(() => {
    if (!autoFocus) return;
    const el = taRef.current;
    if (!el) return;
    el.focus();
    el.setSelectionRange(el.value.length, el.value.length);
  }, [autoFocus]);

  // Caret-based @mention detection. Walk back to a whitespace boundary, then the
  // first char of that run must be the trigger "@". The query (everything after
  // it) may itself contain email chars and a single inner "@", so a pasted email
  // like "john@binus.ac.id" is captured whole — the backend exact-matches it by
  // email and we insert the resolved @username, never a username collision.
  function detectMentionAt(value: string, caret: number) {
    let start = caret;
    while (start > 0 && !/\s/.test(value[start - 1])) start--;
    if (value[start] !== '@') return null;
    const query = value.slice(start + 1, caret);
    // Query is plain text (username-in-progress) or exactly one email "@".
    if ((query.match(/@/g)?.length ?? 0) > 1) return null;
    if (query.length > 60) return null;
    return { query, start, end: caret };
  }

  function syncMention(value: string, caret: number) {
    const d = detectMentionAt(value, caret);
    if (d) { setMentionState(d); setMentionIndex(0); }
    else setMentionState(null);
  }

  function handleChange(value: string) {
    setText(value);
    syncMention(value, taRef.current?.selectionStart ?? value.length);
  }

  function handleSelect() {
    const el = taRef.current;
    if (el) syncMention(el.value, el.selectionStart ?? el.value.length);
  }

  // Replace the active @query span, then drop the caret right after the insert.
  function applyInsertion(insertion: string) {
    const state = mentionState;
    if (!state) return;
    setText((prev) => prev.slice(0, state.start) + insertion + prev.slice(state.end));
    setMentionState(null);
    setMentionIndex(0);
    requestAnimationFrame(() => {
      const el = taRef.current;
      if (!el) return;
      el.focus();
      const pos = state.start + insertion.length;
      el.setSelectionRange(pos, pos);
    });
  }

  function insertMention(user: MentionUser) {
    mentionMap.current.set(user.username, Number(user.user_id));
    applyInsertion(`@${user.username} `);
  }

  // "Everyone · @all" — recognized by the literal "all" in the live highlight and
  // expanded to every invitee by the backend on submit (no map entry needed).
  function insertAll() {
    applyInsertion('@all ');
  }

  function handleMentionResults(results: MentionUser[]) {
    mentionResultsRef.current = results;
    setMentionIndex((i) => (results.length === 0 ? 0 : Math.min(i, results.length - 1)));
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (!mentionState) return;
    if (e.key === 'Escape') { e.preventDefault(); setMentionState(null); return; }
    const results = mentionResultsRef.current;
    if (results.length === 0) return;
    if (e.key === 'ArrowDown') { e.preventDefault(); setMentionIndex((i) => (i + 1) % results.length); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); setMentionIndex((i) => (i - 1 + results.length) % results.length); }
    else if (e.key === 'Enter' || e.key === 'Tab') {
      e.preventDefault();
      const target = results[Math.min(mentionIndex, results.length - 1)];
      if (target) insertMention(target);
    }
  }

  const showEveryone = mentionState !== null && 'all'.startsWith(mentionState.query.toLowerCase());

  async function submit() {
    const content = encodeMentions(text, mentionMap.current).trim();
    if (!content) return;
    try {
      await onSubmit(content);
      if (resetAfterSubmit) {
        setText('');
        mentionMap.current.clear();
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to save comment.');
    }
  }

  const disabled = pending || !text.trim();

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div
        className="ce-composer"
        style={{
          border: '1px solid var(--border)',
          borderRadius: 12,
          background: 'var(--bg-input, var(--bg-card))',
          overflow: 'hidden',
        }}
      >
        <MentionHighlightTextarea
          value={text}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          onSelect={handleSelect}
          placeholder="Write a comment… use @ to mention, or @all for everyone"
          recognize={(w) => w === 'all' || mentionMap.current.has(w)}
          textareaRef={taRef}
          onTextareaEl={setTaEl}
        />
      </div>

      {mentionState && (
        <MentionAutocomplete
          query={mentionState.query}
          anchorEl={taEl}
          mentionStart={mentionState.start}
          caretIndex={mentionState.end}
          activeIndex={mentionIndex}
          onActiveIndexChange={setMentionIndex}
          onSelect={insertMention}
          onClose={() => setMentionState(null)}
          onResults={handleMentionResults}
          topSlot={showEveryone ? <EveryoneRow onSelect={insertAll} /> : undefined}
        />
      )}
      <div style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center', gap: 8 }}>
        {onCancel && (
          <button
            type="button"
            className="btn-ghost"
            onClick={onCancel}
            disabled={pending}
            style={{
              padding: '8px 16px',
              borderRadius: 999,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
              fontSize: 14,
              fontWeight: 700,
              cursor: pending ? 'default' : 'pointer',
              opacity: pending ? 0.5 : 1,
            }}
          >
            Cancel
          </button>
        )}
        <button
          type="button"
          className="btn-solid"
          onClick={submit}
          disabled={disabled}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            padding: '8px 16px',
            borderRadius: 999,
            border: 'none',
            background: 'var(--text)',
            color: 'var(--bg)',
            fontSize: 14,
            fontWeight: 700,
            cursor: disabled ? 'default' : 'pointer',
            opacity: disabled ? 0.5 : 1,
          }}
        >
          {submitLabel}
        </button>
      </div>
    </div>
  );
}

function Composer({ eventId }: { eventId: number }) {
  const addComment = useAddCommunityEventComment(eventId);
  const { data: profile } = useMyProfile();
  return (
    <div style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
      <div style={{ paddingTop: 2, flexShrink: 0 }}>
        <Avatar username={profile?.username ?? ''} avatarUrl={profile?.avatar ?? null} size="sm" />
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <CommentField
          pending={addComment.isPending}
          resetAfterSubmit
          onSubmit={async (content) => {
            await addComment.mutateAsync(content);
          }}
        />
      </div>
    </div>
  );
}

function EditCommentModal({
  eventId,
  comment,
  open,
  onClose,
}: {
  eventId: number;
  comment: CommunityEventCommentData;
  open: boolean;
  onClose: () => void;
}) {
  const editComment = useEditCommunityEventComment(eventId);
  const initial = decodeMentionsForEditing(comment.content);
  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v && !editComment.isPending) onClose(); }}>
      <DialogContent className="max-w-[520px] !p-0 !rounded-2xl overflow-hidden">
        <div style={{ padding: '20px 22px' }}>
          <h2 style={{ fontSize: 18, fontWeight: 800, color: 'var(--text)', margin: 0, letterSpacing: '-0.01em' }}>
            Edit comment
          </h2>
          <p style={{ fontSize: 13, color: 'var(--text-tertiary)', margin: '4px 0 16px' }}>
            Mention people with @, or @all for everyone.
          </p>
          <CommentField
            key={comment.id}
            initial={initial}
            submitLabel="Save"
            autoFocus
            pending={editComment.isPending}
            onCancel={onClose}
            onSubmit={async (content) => {
              await editComment.mutateAsync({ commentId: comment.id, content });
              onClose();
            }}
          />
        </div>
      </DialogContent>
    </Dialog>
  );
}

export function CommunityEventDetail({ eventId }: { eventId: number }) {
  const router = useRouter();
  const currentUserId = useAuthStore((s) => s.user?.id);
  const { data: event, isLoading, isError } = useCommunityEvent(eventId);
  const { data: commentsData } = useCommunityEventComments(eventId);
  const [editOpen, setEditOpen] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [commentsOpen, setCommentsOpen] = useState(true);
  const deleteEvent = useDeleteCommunityEvent(event?.community_id ?? 0);

  if (isLoading) {
    return <div style={{ padding: 24 }}><div style={{ height: 200, borderRadius: 14, background: 'var(--bg-subtle)' }} /></div>;
  }
  if (isError || !event) {
    return <EmptyState title="Event not found" subtitle="This event doesn't exist or you don't have access to it." />;
  }

  const comments = commentsData?.data ?? [];
  // Pinned comment floats to the top; rest keep posted order (stable sort).
  const orderedComments = [...comments].sort((a, b) => Number(b.is_pinned) - Number(a.is_pinned));

  async function handleDelete() {
    if (!event) return;
    try {
      await deleteEvent.mutateAsync(event.id);
      toast.success('Event deleted.');
      router.push(`/communities/${event.community_id}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete event.');
    }
  }

  return (
    <div style={{ maxWidth: 720, margin: '0 auto', paddingBottom: 64 }}>
      {/* Sticky top bar — back + title, mirrors the community header */}
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 20,
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          padding: '10px 12px',
          background: 'color-mix(in oklab, var(--bg) 82%, transparent)',
          backdropFilter: 'blur(8px)',
          WebkitBackdropFilter: 'blur(8px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <button
          onClick={() => router.back()}
          aria-label="Back"
          className="btn-ghost"
          style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: 36, height: 36, borderRadius: 999, border: 'none', background: 'transparent', color: 'var(--text)', cursor: 'pointer', flexShrink: 0 }}
        >
          <ArrowLeft size={20} />
        </button>
        <span style={{ fontSize: 18, fontWeight: 800, color: 'var(--text)', letterSpacing: '-0.01em', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
          {event.title}
        </span>
      </div>

      {/* Banner (full-bleed) */}
      <div style={{ height: 220, background: 'var(--bg-subtle)', overflow: 'hidden' }}>
        {event.banner_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={event.banner_url} alt={event.title} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
        ) : (
          <div style={{ width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-tertiary)' }}>
            <CalendarDays size={40} />
          </div>
        )}
      </div>

      {/* Content */}
      <div style={{ padding: '16px 16px 0' }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12, marginBottom: 12 }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <h1 style={{ margin: 0, fontSize: 28, fontWeight: 800, color: 'var(--text)', letterSpacing: '-0.02em' }}>{event.title}</h1>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 14, color: 'var(--text-secondary)', marginTop: 8 }}>
            <CalendarDays size={15} />
            {formatRange(event.start_time, event.end_time)}
          </div>
          {event.location && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 14, color: 'var(--text-tertiary)', marginTop: 4 }}>
              <MapPin size={15} />
              {event.location}
            </div>
          )}
        </div>
        {event.is_admin && (
          <div style={{ display: 'flex', gap: 8, flexShrink: 0 }}>
            <button
              onClick={() => setEditOpen(true)}
              aria-label="Edit event"
              className="btn-ghost"
              style={{ border: '1px solid var(--border)', borderRadius: 999, width: 36, height: 36, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'transparent', color: 'var(--text)', cursor: 'pointer', flexShrink: 0 }}
            >
              <Pencil size={16} />
            </button>
            <button
              onClick={() => setConfirmDelete(true)}
              aria-label="Delete event"
              className="btn-ghost"
              style={{ border: '1px solid var(--border)', borderRadius: 999, width: 36, height: 36, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'transparent', color: 'var(--text)', cursor: 'pointer', flexShrink: 0 }}
            >
              <Trash2 size={16} />
            </button>
          </div>
        )}
      </div>

      {event.description && (
        <p style={{ fontSize: 15, color: 'var(--text)', lineHeight: 1.6, whiteSpace: 'pre-wrap', marginBottom: 16 }}>
          {event.description}
        </p>
      )}

      {/* Invitees */}
      {event.invitees.length > 0 && (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 16, flexWrap: 'wrap' }}>
          <Users size={15} color="var(--text-tertiary)" />
          {event.invitees.map((inv) => (
            <span key={inv.user_id} title={`@${inv.username}`}>
              <Avatar username={inv.username} avatarUrl={inv.avatar_url || null} size={28} />
            </span>
          ))}
        </div>
      )}

      {/* Reactions */}
      <div style={{ marginBottom: 24 }}>
        <ReactionBar event={event} eventId={eventId} />
      </div>

      {/* Comments */}
      <div style={{ borderTop: '1px solid var(--border)', paddingTop: 24, marginTop: 8 }}>
        <h2 style={{ fontSize: 16, fontWeight: 700, color: 'var(--text)', marginBottom: 12 }}>
          Comments ({event.comment_count})
        </h2>

        {event.can_comment ? (
          <div
            style={{
              position: 'sticky',
              top: 56,
              zIndex: 10,
              marginBottom: 16,
              paddingBottom: 12,
              background: 'color-mix(in oklab, var(--bg) 90%, transparent)',
              backdropFilter: 'blur(6px)',
              WebkitBackdropFilter: 'blur(6px)',
            }}
          >
            <Composer eventId={eventId} />
          </div>
        ) : (
          <p style={{ fontSize: 13, color: 'var(--text-tertiary)', marginBottom: 16 }}>
            Only tagged people can comment on this event.
          </p>
        )}

        {comments.length > 0 && (
          <button
            type="button"
            className="btn-ghost"
            onClick={() => setCommentsOpen((v) => !v)}
            aria-expanded={commentsOpen}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              marginBottom: 8,
              padding: '6px 12px',
              border: 'none',
              borderRadius: 999,
              background: 'transparent',
              cursor: 'pointer',
              color: 'var(--text)',
              fontSize: 14,
              fontWeight: 700,
            }}
          >
            <ChevronDown
              size={16}
              style={{ transition: 'transform 0.2s ease', transform: commentsOpen ? 'rotate(0deg)' : 'rotate(-90deg)' }}
            />
            {commentsOpen ? 'Hide comments' : 'Show comments'}
          </button>
        )}

        {comments.length === 0 ? (
          <p style={{ fontSize: 14, color: 'var(--text-tertiary)', padding: '12px 0' }}>No comments yet.</p>
        ) : (
          <div className={`ce-collapse${commentsOpen ? ' ce-collapse-open' : ''}`}>
            <div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6, paddingTop: 4, paddingBottom: 8 }}>
                {orderedComments.map((comment) => (
                  <CommentRow
                    key={comment.id}
                    comment={comment}
                    eventId={eventId}
                    isAdmin={event.is_admin}
                    currentUserId={currentUserId}
                  />
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
      </div>

      {event.is_admin && (
        <CreateCommunityEventDialog open={editOpen} onOpenChange={setEditOpen} communityId={event.community_id} event={event} />
      )}
      <DangerConfirmModal
        open={confirmDelete}
        onClose={() => setConfirmDelete(false)}
        title="Delete event?"
        description="This permanently removes the event, its comments, and reactions."
        confirmLabel="Delete"
        busy={deleteEvent.isPending}
        onConfirm={handleDelete}
      />
    </div>
  );
}
