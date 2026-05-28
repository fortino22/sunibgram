'use client';

import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useMentionSearch } from '@/hooks/thread/useMentionSearch';
import type { MentionUser } from '@/lib/services/threads';

interface MentionAutocompleteProps {
  query: string;
  anchorEl: HTMLTextAreaElement | null;
  mentionStart: number;
  caretIndex: number;
  activeIndex: number;
  onActiveIndexChange: (i: number) => void;
  onSelect: (user: MentionUser) => void;
  onClose: () => void;
  onResults?: (results: MentionUser[]) => void;
}

interface Position {
  top: number;
  left: number;
}

const MIRROR_COPIED_PROPS = [
  'box-sizing',
  'width',
  'height',
  'overflow-x',
  'overflow-y',
  'border-top-width',
  'border-right-width',
  'border-bottom-width',
  'border-left-width',
  'border-style',
  'padding-top',
  'padding-right',
  'padding-bottom',
  'padding-left',
  'font-style',
  'font-variant',
  'font-weight',
  'font-stretch',
  'font-size',
  'font-size-adjust',
  'line-height',
  'font-family',
  'text-align',
  'text-transform',
  'text-indent',
  'text-decoration',
  'letter-spacing',
  'word-spacing',
  'tab-size',
  'white-space',
  'word-break',
  'overflow-wrap',
] as const;

function getCaretCoordinates(el: HTMLTextAreaElement, position: number): { top: number; left: number; height: number } {
  const doc = el.ownerDocument;
  const div = doc.createElement('div');
  const style = div.style;
  const computed = window.getComputedStyle(el);

  style.position = 'absolute';
  style.visibility = 'hidden';
  style.top = '0';
  style.left = '0';
  style.whiteSpace = 'pre-wrap';
  style.wordWrap = 'break-word';

  for (const prop of MIRROR_COPIED_PROPS) {
    style.setProperty(prop, computed.getPropertyValue(prop));
  }

  div.textContent = el.value.substring(0, position);
  const span = doc.createElement('span');
  span.textContent = el.value.substring(position) || '.';
  div.appendChild(span);
  doc.body.appendChild(div);

  const top = span.offsetTop;
  const left = span.offsetLeft;
  const height = parseInt(computed.lineHeight, 10) || parseInt(computed.fontSize, 10) * 1.2;

  doc.body.removeChild(div);
  return { top, left, height };
}

export function MentionAutocomplete({
  query,
  anchorEl,
  mentionStart,
  caretIndex,
  activeIndex,
  onActiveIndexChange,
  onSelect,
  onClose,
  onResults,
}: MentionAutocompleteProps) {
  const { data, isLoading } = useMentionSearch(query, true);
  const results = data ?? [];
  const [pos, setPos] = useState<Position | null>(null);
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    onResults?.(data ?? []);
  }, [data, onResults]);

  useLayoutEffect(() => {
    if (!anchorEl) return;
    const update = () => {
      const r = anchorEl.getBoundingClientRect();
      const anchor = getCaretCoordinates(anchorEl, mentionStart);
      const caret = getCaretCoordinates(anchorEl, caretIndex);
      const top = r.top + caret.top - anchorEl.scrollTop + caret.height + 6;
      const left = r.left + anchor.left - anchorEl.scrollLeft;
      setPos({ top, left });
    };
    update();
    window.addEventListener('scroll', update, true);
    window.addEventListener('resize', update);
    return () => {
      window.removeEventListener('scroll', update, true);
      window.removeEventListener('resize', update);
    };
  }, [anchorEl, mentionStart, caretIndex]);

  useEffect(() => {
    const onDown = (e: MouseEvent) => {
      const t = e.target as Node;
      if (popoverRef.current?.contains(t)) return;
      if (anchorEl?.contains(t)) return;
      onClose();
    };
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, [anchorEl, onClose]);

  if (!pos || typeof document === 'undefined') return null;

  const popover = (
    <div
      ref={popoverRef}
      role="listbox"
      style={{
        position: 'fixed',
        top: pos.top,
        left: pos.left,
        width: 340,
        maxHeight: 360,
        overflowY: 'auto',
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
        borderRadius: 14,
        boxShadow: '0 4px 24px rgba(0,0,0,0.22)',
        zIndex: 1200,
        padding: 4,
      }}
    >
      {isLoading && results.length === 0 ? (
        <SkeletonRows />
      ) : results.length === 0 ? (
        <div
          style={{
            padding: '12px 14px',
            color: 'var(--text-tertiary)',
            fontSize: 14,
          }}
        >
          No users found
        </div>
      ) : (
        results.map((user, i) => (
          <button
            key={user.user_id}
            type="button"
            role="option"
            aria-selected={i === activeIndex}
            onMouseEnter={() => onActiveIndexChange(i)}
            onMouseDown={(e) => {
              e.preventDefault();
              onSelect(user);
            }}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              width: '100%',
              padding: '8px 10px',
              borderRadius: 10,
              border: 'none',
              background: i === activeIndex ? 'rgba(255,255,255,0.04)' : 'transparent',
              cursor: 'pointer',
              textAlign: 'left',
            }}
          >
            <Avatar username={user.username} avatarUrl={user.user_profile || null} size={32} />
            <div style={{ display: 'flex', flexDirection: 'column', minWidth: 0, flex: 1 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 4, minWidth: 0 }}>
                <span
                  style={{
                    fontWeight: 700,
                    fontSize: 14,
                    color: 'var(--text)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {user.name || user.username}
                </span>
                {user.is_premium && <VerifiedBadge size={14} />}
              </div>
              <span
                style={{
                  fontSize: 13,
                  color: 'var(--text-tertiary)',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                }}
              >
                @{user.username}
              </span>
            </div>
          </button>
        ))
      )}
    </div>
  );

  return createPortal(popover, document.body);
}

function SkeletonRows() {
  return (
    <>
      {[0, 1, 2].map((i) => (
        <div
          key={i}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: '8px 10px',
          }}
        >
          <div
            style={{
              width: 32,
              height: 32,
              borderRadius: '50%',
              background: 'var(--bg-subtle)',
            }}
          />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
            <div style={{ height: 10, width: '40%', borderRadius: 4, background: 'var(--bg-subtle)' }} />
            <div style={{ height: 10, width: '30%', borderRadius: 4, background: 'var(--bg-subtle)' }} />
          </div>
        </div>
      ))}
    </>
  );
}
