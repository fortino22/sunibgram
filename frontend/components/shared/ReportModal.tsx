'use client';

import { useState } from 'react';
import { useReportPost, useReportUser, useViolations } from '@/hooks/useReports';

type ReportTarget =
  | { type: 'post'; postId: number }
  | { type: 'user'; userId: number };

interface ReportModalProps {
  target: ReportTarget;
  onClose: () => void;
}

const COPY = {
  post: {
    title: 'Report post',
    question: 'Why are you reporting this post?',
  },
  user: {
    title: 'Report user',
    question: 'Why are you reporting this account?',
  },
} as const;

export function ReportModal({ target, onClose }: ReportModalProps) {
  const { data: violationsData, isLoading: violationsLoading } = useViolations();
  const violations = violationsData?.data ?? [];

  const postMutation = useReportPost();
  const userMutation = useReportUser();
  const { isPending, error } =
    target.type === 'post' ? postMutation : userMutation;

  const [selectedViolationId, setSelectedViolationId] = useState('');
  const [explanation, setExplanation] = useState('');

  const copy = COPY[target.type];

  const canSubmit =
    selectedViolationId.length > 0 &&
    explanation.trim().length > 0 &&
    explanation.length <= 2000 &&
    !isPending;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    try {
      if (target.type === 'post') {
        await postMutation.mutateAsync({
          violation_id: selectedViolationId,
          reported_post_id: target.postId,
          explanation: explanation.trim(),
        });
      } else {
        await userMutation.mutateAsync({
          violation_id: selectedViolationId,
          reported_user_id: target.userId,
          explanation: explanation.trim(),
        });
      }
      onClose();
    } catch {
      // surfaced via mutation error state
    }
  }

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.6)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 100,
        padding: 16,
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        style={{
          background: 'var(--bg-card)',
          borderRadius: 16,
          width: '100%',
          maxWidth: 460,
          maxHeight: '90vh',
          overflowY: 'auto',
          border: '1px solid var(--border)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '16px 20px',
            borderBottom: '1px solid var(--border)',
            position: 'sticky',
            top: 0,
            background: 'var(--bg-card)',
            zIndex: 1,
          }}
        >
          <span style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)' }}>{copy.title}</span>
          <button
            type="button"
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--text-tertiary)',
              fontSize: 20,
              cursor: 'pointer',
            }}
          >
            ✕
          </button>
        </div>

        <form
          onSubmit={handleSubmit}
          style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 16 }}
        >
          <div>
            <label
              style={{
                fontSize: 12,
                fontWeight: 600,
                color: 'var(--text-secondary)',
                display: 'block',
                marginBottom: 8,
              }}
            >
              {copy.question}
            </label>

            {violationsLoading ? (
              <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>Loading…</div>
            ) : violations.length === 0 ? (
              <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>
                No violation types available.
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                {violations.map((v) => {
                  const selected = selectedViolationId === v.id;
                  return (
                    <label
                      key={v.id}
                      style={{
                        display: 'flex',
                        gap: 10,
                        padding: '10px 12px',
                        borderRadius: 8,
                        border: `1px solid ${selected ? 'var(--text)' : 'var(--border)'}`,
                        background: selected ? 'var(--bg-subtle)' : 'transparent',
                        cursor: 'pointer',
                      }}
                    >
                      <input
                        type="radio"
                        name="violation"
                        value={v.id}
                        checked={selected}
                        onChange={() => setSelectedViolationId(v.id)}
                        style={{ marginTop: 3, accentColor: 'var(--text)' }}
                      />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)' }}>
                          {v.name}
                        </div>
                        {v.description && (
                          <div
                            style={{
                              fontSize: 12,
                              color: 'var(--text-tertiary)',
                              marginTop: 2,
                              lineHeight: 1.4,
                            }}
                          >
                            {v.description}
                          </div>
                        )}
                      </div>
                    </label>
                  );
                })}
              </div>
            )}
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
              <label
                style={{ fontSize: 12, fontWeight: 600, color: 'var(--text-secondary)' }}
              >
                Explanation
              </label>
              <span
                style={{
                  fontSize: 11,
                  color:
                    explanation.length > 2000
                      ? 'var(--error, #e53e3e)'
                      : 'var(--text-tertiary)',
                }}
              >
                {explanation.length}/2000
              </span>
            </div>
            <textarea
              value={explanation}
              onChange={(e) => setExplanation(e.target.value)}
              placeholder="Tell us more about the issue."
              maxLength={2000}
              style={{
                width: '100%',
                padding: '10px 12px',
                borderRadius: 8,
                border: '1px solid var(--border)',
                background: 'var(--bg-subtle)',
                color: 'var(--text)',
                fontSize: 14,
                boxSizing: 'border-box',
                minHeight: 96,
                resize: 'vertical',
                fontFamily: 'inherit',
              }}
            />
          </div>

          {error && (
            <div style={{ fontSize: 13, color: 'var(--error, #e53e3e)' }}>
              {error instanceof Error ? error.message : 'Something went wrong'}
            </div>
          )}

          <button
            type="submit"
            className="btn-solid"
            disabled={!canSubmit}
            style={{
              width: '100%',
              padding: '12px',
              borderRadius: 999,
              border: 'none',
              background: 'var(--text)',
              color: 'var(--bg)',
              fontSize: 15,
              fontWeight: 700,
              cursor: !canSubmit ? 'default' : 'pointer',
              opacity: !canSubmit ? 0.5 : 1,
            }}
          >
            {isPending ? 'Submitting…' : 'Submit report'}
          </button>
        </form>
      </div>
    </div>
  );
}
