'use client';

import { useState } from 'react';
import { X, Bug, Hammer } from 'lucide-react';
import {
  FEEDBACK_BODY_MAX,
  FEEDBACK_TITLE_MAX,
  type CreateFeedbackInput,
  type FeedbackType,
} from '@/lib/services/feedback';

interface FeedbackComposerModalProps {
  busy?: boolean;
  errorMessage?: string | null;
  onSubmit: (input: CreateFeedbackInput) => void;
  onClose: () => void;
}

export function FeedbackComposerModal({
  busy,
  errorMessage,
  onSubmit,
  onClose,
}: FeedbackComposerModalProps) {
  const [type, setType] = useState<FeedbackType>('bug');
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');

  const trimmedTitle = title.trim();
  const trimmedBody = body.trim();
  const canSubmit =
    !busy &&
    trimmedTitle.length > 0 &&
    trimmedTitle.length <= FEEDBACK_TITLE_MAX &&
    trimmedBody.length > 0 &&
    trimmedBody.length <= FEEDBACK_BODY_MAX;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label="New feedback"
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 100,
        padding: 16,
        animation: 'modal-backdrop-in 0.12s ease',
      }}
      onClick={(e) => { if (e.target === e.currentTarget && !busy) onClose(); }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          width: '100%',
          maxWidth: 520,
          maxHeight: '90vh',
          overflowY: 'auto',
          animation: 'modal-panel-in 0.16s ease',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '14px 20px',
            borderBottom: '1px solid var(--border)',
            position: 'sticky',
            top: 0,
            background: 'var(--bg-card)',
            zIndex: 1,
          }}
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.02em' }}>
              Send feedback
            </span>
            <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>
              Report a bug or suggest a feature.
            </span>
          </div>
          <button
            type="button"
            onClick={onClose}
            disabled={busy}
            className="btn-ghost"
            aria-label="Close"
            style={{
              background: 'none',
              border: 'none',
              cursor: busy ? 'not-allowed' : 'pointer',
              color: 'var(--text)',
              padding: 6,
              borderRadius: '50%',
              display: 'inline-flex',
            }}
          >
            <X size={18} strokeWidth={2} />
          </button>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (!canSubmit) return;
            onSubmit({ type, title: trimmedTitle, body: trimmedBody });
          }}
          style={{ padding: '18px 20px 22px', display: 'flex', flexDirection: 'column', gap: 16 }}
        >
          <div
            role="tablist"
            aria-label="Feedback type"
            style={{
              display: 'grid',
              gridTemplateColumns: '1fr 1fr',
              gap: 8,
            }}
          >
            <TypeOption
              kind="bug"
              active={type === 'bug'}
              onClick={() => setType('bug')}
              disabled={busy}
            />
            <TypeOption
              kind="feature_request"
              active={type === 'feature_request'}
              onClick={() => setType('feature_request')}
              disabled={busy}
            />
          </div>

          <div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: 6,
              }}
            >
              <label htmlFor="feedback-title" style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}>
                Title
              </label>
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>
                {trimmedTitle.length}/{FEEDBACK_TITLE_MAX}
              </span>
            </div>
            <input
              id="feedback-title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={
                type === 'bug'
                  ? 'Short summary of the issue'
                  : 'Short summary of what you want'
              }
              maxLength={FEEDBACK_TITLE_MAX}
              autoFocus
              disabled={busy}
              style={{
                width: '100%',
                padding: '10px 12px',
                borderRadius: 10,
                border: '1px solid var(--border)',
                background: 'var(--bg-subtle)',
                color: 'var(--text)',
                fontSize: 14,
                boxSizing: 'border-box',
                fontFamily: 'inherit',
                outline: 'none',
              }}
            />
          </div>

          <div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: 6,
              }}
            >
              <label htmlFor="feedback-body" style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}>
                Details
              </label>
              <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>
                {trimmedBody.length}/{FEEDBACK_BODY_MAX}
              </span>
            </div>
            <textarea
              id="feedback-body"
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder={
                type === 'bug'
                  ? 'What happened? Steps to reproduce, what you expected, what actually happened.'
                  : 'Describe the feature, who benefits, and how it should work.'
              }
              maxLength={FEEDBACK_BODY_MAX}
              disabled={busy}
              style={{
                width: '100%',
                padding: '12px',
                borderRadius: 10,
                border: '1px solid var(--border)',
                background: 'var(--bg-subtle)',
                color: 'var(--text)',
                fontSize: 14,
                boxSizing: 'border-box',
                minHeight: 140,
                resize: 'vertical',
                fontFamily: 'inherit',
                outline: 'none',
                lineHeight: 1.55,
              }}
            />
          </div>

          {errorMessage && (
            <div style={{ fontSize: 13, color: 'var(--coral)' }}>{errorMessage}</div>
          )}

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
            <button
              type="button"
              onClick={onClose}
              disabled={busy}
              className="btn-ghost"
              style={{
                padding: '10px 18px',
                borderRadius: 9999,
                border: '1px solid var(--border)',
                background: 'transparent',
                color: 'var(--text)',
                fontSize: 14,
                fontWeight: 600,
                cursor: busy ? 'not-allowed' : 'pointer',
                fontFamily: 'inherit',
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!canSubmit}
              className="btn-solid"
              style={{
                padding: '10px 20px',
                borderRadius: 9999,
                border: 'none',
                background: 'var(--text)',
                color: 'var(--bg)',
                fontSize: 14,
                fontWeight: 700,
                cursor: !canSubmit ? 'not-allowed' : 'pointer',
                opacity: !canSubmit ? 0.55 : 1,
                fontFamily: 'inherit',
                letterSpacing: '-0.01em',
              }}
            >
              {busy ? 'Sending…' : 'Send feedback'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

interface TypeOptionProps {
  kind: FeedbackType;
  active: boolean;
  disabled?: boolean;
  onClick: () => void;
}

function TypeOption({ kind, active, disabled, onClick }: TypeOptionProps) {
  const isBug = kind === 'bug';
  const Glyph = isBug ? Bug : Hammer;
  const accent = isBug ? 'var(--coral)' : 'var(--accent)';
  const activeBg = isBug
    ? 'color-mix(in oklab, var(--coral) 14%, transparent)'
    : 'color-mix(in oklab, var(--accent) 14%, transparent)';

  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      disabled={disabled}
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'flex-start',
        gap: 6,
        padding: '12px 14px',
        borderRadius: 12,
        border: `1px solid ${active ? `color-mix(in oklab, ${accent} 60%, transparent)` : 'var(--border)'}`,
        background: active ? activeBg : 'var(--bg-subtle)',
        color: 'var(--text)',
        cursor: disabled ? 'not-allowed' : 'pointer',
        fontFamily: 'inherit',
        textAlign: 'left',
        transition: 'background-color 0.12s ease, border-color 0.12s ease',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <Glyph size={16} strokeWidth={2.2} color={active ? accent : 'var(--text-secondary)'} />
        <span style={{ fontSize: 14, fontWeight: 700, letterSpacing: '-0.01em' }}>
          {isBug ? 'Bug report' : 'Feature request'}
        </span>
      </div>
      <span style={{ fontSize: 12, color: 'var(--text-tertiary)', lineHeight: 1.4 }}>
        {isBug
          ? 'Something is broken or behaving wrong.'
          : 'An idea or improvement for the product.'}
      </span>
    </button>
  );
}
