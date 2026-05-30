'use client';

import { useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { Frown } from 'lucide-react';
import { useReportTrend } from '@/hooks/useTrendReports';
import type { TrendReportReasonCode } from '@/lib/services/trendReports';

export interface AnchorRect {
  top: number;
  left: number;
  bottom: number;
  right: number;
  width: number;
  height: number;
}

interface ReportTrendModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  hashtagName: string | null;
  anchorRect: AnchorRect | null;
  onReported?: () => void;
}

const REASONS: { code: TrendReportReasonCode; label: string }[] = [
  { code: 'not_relevant', label: 'The associated content is not relevant' },
  { code: 'spam', label: 'This trend is spam' },
  { code: 'abusive_or_harmful', label: 'This trend is abusive or harmful' },
  { code: 'not_interested', label: 'Not interested in this' },
  { code: 'duplicate', label: 'This trend is a duplicate' },
  { code: 'harmful_or_spammy', label: 'This trend is harmful or spammy' },
];

const POPOVER_WIDTH = 280;
const POPOVER_HEIGHT_ESTIMATE = 300;
const VIEWPORT_PAD = 8;

function computePosition(anchor: AnchorRect): { top: number; left: number } {
  const vw = typeof window !== 'undefined' ? window.innerWidth : 1024;
  const vh = typeof window !== 'undefined' ? window.innerHeight : 768;

  let left = anchor.right - POPOVER_WIDTH;
  if (left < VIEWPORT_PAD) left = VIEWPORT_PAD;
  if (left + POPOVER_WIDTH > vw - VIEWPORT_PAD) left = vw - POPOVER_WIDTH - VIEWPORT_PAD;

  let top = anchor.bottom + 6;
  if (top + POPOVER_HEIGHT_ESTIMATE > vh - VIEWPORT_PAD) {
    top = anchor.top - POPOVER_HEIGHT_ESTIMATE - 6;
    if (top < VIEWPORT_PAD) top = VIEWPORT_PAD;
  }
  return { top, left };
}

export function ReportTrendModal({ open, onOpenChange, hashtagName, anchorRect, onReported }: ReportTrendModalProps) {
  const mutation = useReportTrend();
  const popoverRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!open) return undefined;
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onOpenChange(false);
    }
    function onDown(e: MouseEvent) {
      if (!popoverRef.current) return;
      if (!popoverRef.current.contains(e.target as Node)) onOpenChange(false);
    }
    window.addEventListener('keydown', onKey);
    document.addEventListener('mousedown', onDown);
    return () => {
      window.removeEventListener('keydown', onKey);
      document.removeEventListener('mousedown', onDown);
    };
  }, [open, onOpenChange]);

  if (!open || !anchorRect) return null;
  if (typeof document === 'undefined') return null;

  const pos = computePosition(anchorRect);

  const handlePick = (reason_code: TrendReportReasonCode) => {
    if (!hashtagName) return;
    mutation.mutate(
      { hashtag_name: hashtagName, reason_code },
      {
        onSuccess: () => onReported?.(),
        onSettled: () => onOpenChange(false),
      }
    );
  };

  return createPortal(
    <div
      ref={popoverRef}
      role="dialog"
      aria-label="Report trend"
      style={{
        position: 'fixed',
        top: pos.top,
        left: pos.left,
        width: POPOVER_WIDTH,
        zIndex: 60,
        background: 'var(--bg-card)',
        border: '1px solid var(--border)',
        borderRadius: 12,
        boxShadow: '0 10px 30px rgba(0,0,0,0.45)',
        overflow: 'hidden',
      }}
    >
      <div style={{ padding: '10px 14px 4px' }}>
        <h2 className="text-[15px] font-extrabold" style={{ color: 'var(--text)' }}>
          What&apos;s happening
        </h2>
      </div>
      <div>
        {REASONS.map((reason) => (
          <button
            key={reason.code}
            type="button"
            disabled={mutation.isPending}
            onClick={() => handlePick(reason.code)}
            className="flex w-full items-center gap-3 text-left transition-colors disabled:opacity-50"
            style={{
              padding: '8px 14px',
              color: 'var(--text)',
              background: 'transparent',
              border: 'none',
              cursor: mutation.isPending ? 'not-allowed' : 'pointer',
            }}
            onMouseEnter={(e) => {
              if (mutation.isPending) return;
              (e.currentTarget as HTMLElement).style.background = 'var(--bg-hover)';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'transparent';
            }}
          >
            <Frown size={16} style={{ color: 'var(--text-tertiary)', flexShrink: 0 }} />
            <span className="text-[13px] font-semibold">{reason.label}</span>
          </button>
        ))}
      </div>
    </div>,
    document.body
  );
}
