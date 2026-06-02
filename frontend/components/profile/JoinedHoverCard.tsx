'use client';

import { useCallback, useRef, useState, type CSSProperties, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Calendar, MapPin, AtSign, Globe } from 'lucide-react';
import { useUsernameHistory } from '@/hooks/useUsernameHistory';

const OPEN_DELAY = 350;
const CLOSE_DELAY = 200;
const CARD_WIDTH = 280;
const EST_HEIGHT = 180;

interface JoinedHoverCardProps {
  userId: number | string;
  joinedMonth: string;
  /** The trigger element (the Joined link). */
  children: ReactNode;
  style?: CSSProperties;
}

type Placement = { left: number; top: number };

const labelStyle: CSSProperties = { fontSize: 13, fontWeight: 600, color: 'var(--text)', lineHeight: 1.3 };
const valueStyle: CSSProperties = { fontSize: 12, color: 'var(--text-tertiary)', marginTop: 1 };

function Row({ icon, label, value }: { icon: ReactNode; label: string; value: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
      <span style={{ color: 'var(--text-tertiary)', flexShrink: 0, marginTop: 1 }}>{icon}</span>
      <div style={{ minWidth: 0 }}>
        <div style={labelStyle}>{label}</div>
        <div style={valueStyle}>{value}</div>
      </div>
    </div>
  );
}

/**
 * X-style hover card for the profile "Joined" line. On hover-intent it lazily
 * fetches the username-change count and shows a compact "about this account"
 * preview. Clicking the wrapped link still navigates to the full About page.
 */
export function JoinedHoverCard({ userId, joinedMonth, children, style }: JoinedHoverCardProps) {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState<Placement | null>(null);
  const wrapRef = useRef<HTMLSpanElement>(null);
  const openTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const closeTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  // Lazy: only fetch the count once the card is opening.
  const { data: history } = useUsernameHistory(open ? userId : null);
  const count = history?.count ?? 0;

  const computePos = useCallback(() => {
    const el = wrapRef.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    const placeBelow = r.bottom + EST_HEIGHT + 12 < window.innerHeight;
    const top = placeBelow ? r.bottom + 8 : Math.max(8, r.top - EST_HEIGHT - 8);
    const left = Math.max(8, Math.min(r.left, window.innerWidth - CARD_WIDTH - 8));
    setPos({ left, top });
  }, []);

  const scheduleOpen = useCallback(() => {
    clearTimeout(closeTimer.current);
    openTimer.current = setTimeout(() => {
      computePos();
      setOpen(true);
    }, OPEN_DELAY);
  }, [computePos]);

  const scheduleClose = useCallback(() => {
    clearTimeout(openTimer.current);
    closeTimer.current = setTimeout(() => setOpen(false), CLOSE_DELAY);
  }, []);

  return (
    <span
      ref={wrapRef}
      onMouseEnter={scheduleOpen}
      onMouseLeave={scheduleClose}
      style={{ display: 'inline-flex', position: 'relative', ...style }}
    >
      {children}

      {open && pos && typeof document !== 'undefined' &&
        createPortal(
          <div
            onMouseEnter={() => clearTimeout(closeTimer.current)}
            onMouseLeave={scheduleClose}
            onClick={(e) => e.stopPropagation()}
            style={{
              position: 'fixed',
              left: pos.left,
              top: pos.top,
              width: CARD_WIDTH,
              zIndex: 100,
              background: 'var(--bg-card)',
              border: '1px solid var(--border)',
              borderRadius: 16,
              boxShadow: '0 12px 32px rgba(0,0,0,0.5)',
              padding: 16,
              animation: 'modal-panel-in 0.12s ease',
            }}
          >
            <div style={{ fontSize: 12, fontWeight: 700, color: 'var(--text-tertiary)', letterSpacing: '0.02em', textTransform: 'uppercase', marginBottom: 12 }}>
              About this account
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <Row icon={<Calendar size={16} strokeWidth={1.8} />} label="Date joined" value={joinedMonth} />
              <Row icon={<MapPin size={16} strokeWidth={1.8} />} label="Account based in" value="Indonesia" />
              <Row
                icon={<AtSign size={16} strokeWidth={1.8} />}
                label={`${count} username ${count === 1 ? 'change' : 'changes'}`}
                value={count > 0 ? 'Tap to view history' : 'Never changed'}
              />
              <Row icon={<Globe size={16} strokeWidth={1.8} />} label="Connected via" value="web" />
            </div>
          </div>,
          document.body,
        )}
    </span>
  );
}
