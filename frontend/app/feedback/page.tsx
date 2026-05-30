'use client';

import { useEffect, useMemo, useRef, useState, type CSSProperties } from 'react';
import { Plus, Filter, LayoutGrid, List, Check } from 'lucide-react';
import { EmptyState } from '@/components/shared/EmptyState';
import { FeedbackCard } from '@/components/feedback/FeedbackCard';
import { FeedbackCardSkeleton } from '@/components/feedback/FeedbackCardSkeleton';
import { FeedbackComposerModal } from '@/components/feedback/FeedbackComposerModal';
import { useCreateFeedback, useFeedbackList } from '@/hooks/useFeedback';
import {
  FEEDBACK_STATUSES,
  FEEDBACK_STATUS_LABEL,
  type CreateFeedbackInput,
  type FeedbackStatus,
  type FeedbackType,
} from '@/lib/services/feedback';

type TypeTab = 'all' | FeedbackType;

const TYPE_TABS: { key: TypeTab; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'bug', label: 'Bugs' },
  { key: 'feature_request', label: 'Features' },
];

type StatusFilter = 'all' | FeedbackStatus;

const STATUS_FILTERS: { key: StatusFilter; label: string }[] = [
  { key: 'all', label: 'Any status' },
  ...FEEDBACK_STATUSES.map((s) => ({ key: s, label: FEEDBACK_STATUS_LABEL[s] })),
];

type ViewMode = 'grid' | 'list';

export default function FeedbackPage() {
  const [typeTab, setTypeTab] = useState<TypeTab>('all');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [view, setView] = useState<ViewMode>('grid');
  const [filterOpen, setFilterOpen] = useState(false);
  const filterRef = useRef<HTMLDivElement>(null);
  const [page, setPage] = useState(1);
  const [composerOpen, setComposerOpen] = useState(false);
  const [composerError, setComposerError] = useState<string | null>(null);

  useEffect(() => {
    if (!filterOpen) return;
    const handle = (e: MouseEvent) => {
      if (filterRef.current && !filterRef.current.contains(e.target as Node)) {
        setFilterOpen(false);
      }
    };
    document.addEventListener('mousedown', handle);
    return () => document.removeEventListener('mousedown', handle);
  }, [filterOpen]);

  const params = useMemo(
    () => ({
      type: typeTab === 'all' ? null : typeTab,
      status: statusFilter === 'all' ? null : statusFilter,
      page,
      limit: 20,
    }),
    [typeTab, statusFilter, page],
  );

  const query = useFeedbackList(params);
  const createMutation = useCreateFeedback();

  const items = query.data?.data ?? [];
  const total = query.data?.meta.total ?? 0;
  const hasMore = items.length > 0 && page * 20 < total;
  const hasPrev = page > 1;

  function changeType(next: TypeTab) {
    setTypeTab(next);
    setPage(1);
  }

  function changeStatus(next: StatusFilter) {
    setStatusFilter(next);
    setPage(1);
  }

  function submitComposer(input: CreateFeedbackInput) {
    setComposerError(null);
    createMutation.mutate(input, {
      onSuccess: () => {
        setComposerOpen(false);
        setPage(1);
      },
      onError: (err) =>
        setComposerError(err instanceof Error ? err.message : 'Could not send feedback.'),
    });
  }

  return (
    <div>
      <div
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 10,
          background: 'var(--bg-blur)',
          backdropFilter: 'blur(12px)',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '20px 20px 14px',
          }}
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            <div style={{ fontWeight: 500, fontSize: 20, color: 'var(--text)', letterSpacing: '-0.02em' }}>
              Feedback
            </div>
            <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>
              Report bugs, request features, vote on what others want.
            </div>
          </div>
          <button
            type="button"
            onClick={() => {
              setComposerError(null);
              setComposerOpen(true);
            }}
            className="btn-solid"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '8px 16px',
              borderRadius: 9999,
              border: 'none',
              background: 'var(--text)',
              color: 'var(--bg)',
              fontWeight: 700,
              fontSize: 13,
              fontFamily: 'inherit',
              cursor: 'pointer',
              letterSpacing: '-0.01em',
            }}
          >
            <Plus size={14} strokeWidth={2.5} />
            New
          </button>
        </div>

        <div role="tablist" aria-label="Feedback type" style={{ display: 'flex', gap: 0 }}>
          {TYPE_TABS.map((tab) => (
            <TabButton
              key={tab.key}
              label={tab.label}
              active={typeTab === tab.key}
              onClick={() => changeType(tab.key)}
            />
          ))}
        </div>
      </div>

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'flex-end',
          gap: 8,
          padding: '14px 20px 0',
        }}
      >
        <div
          role="group"
          aria-label="View mode"
          style={{
            display: 'flex',
            gap: 2,
            padding: 3,
            borderRadius: 9999,
            background: 'var(--bg-subtle)',
            border: '1px solid var(--border)',
          }}
        >
          <ViewToggle
            kind="grid"
            active={view === 'grid'}
            onClick={() => setView('grid')}
          />
          <ViewToggle
            kind="list"
            active={view === 'list'}
            onClick={() => setView('list')}
          />
        </div>

        <div ref={filterRef} style={{ position: 'relative' }}>
          <button
            type="button"
            onClick={() => setFilterOpen((v) => !v)}
            aria-haspopup="menu"
            aria-expanded={filterOpen}
            className="btn-ghost"
            title="Filter by status"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '8px 12px',
              borderRadius: 9999,
              border: `1px solid ${statusFilter === 'all' ? 'var(--border)' : 'color-mix(in oklab, var(--accent) 55%, transparent)'}`,
              background: statusFilter === 'all'
                ? 'transparent'
                : 'color-mix(in oklab, var(--accent) 14%, transparent)',
              color: 'var(--text)',
              fontFamily: 'inherit',
              fontSize: 13,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            <Filter size={14} strokeWidth={2} />
            {statusFilter !== 'all' && (
              <span style={{ fontSize: 12 }}>
                {FEEDBACK_STATUS_LABEL[statusFilter]}
              </span>
            )}
          </button>

          {filterOpen && (
            <div
              role="menu"
              style={{
                position: 'absolute',
                top: 'calc(100% + 6px)',
                right: 0,
                minWidth: 200,
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 12,
                padding: 6,
                zIndex: 30,
                boxShadow: '0 8px 24px rgba(0,0,0,0.45)',
              }}
            >
              <div
                style={{
                  padding: '6px 10px 8px',
                  fontSize: 10.5,
                  fontWeight: 700,
                  letterSpacing: '0.08em',
                  textTransform: 'uppercase',
                  color: 'var(--text-tertiary)',
                }}
              >
                Status
              </div>
              {STATUS_FILTERS.map((s) => {
                const active = statusFilter === s.key;
                return (
                  <button
                    key={s.key}
                    type="button"
                    role="menuitemradio"
                    aria-checked={active}
                    onClick={() => {
                      changeStatus(s.key);
                      setFilterOpen(false);
                    }}
                    className="btn-ghost"
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      width: '100%',
                      padding: '8px 10px',
                      borderRadius: 8,
                      border: 'none',
                      background: active
                        ? 'color-mix(in oklab, var(--accent) 14%, transparent)'
                        : 'transparent',
                      color: 'var(--text)',
                      fontFamily: 'inherit',
                      fontSize: 13,
                      fontWeight: active ? 700 : 500,
                      cursor: 'pointer',
                      textAlign: 'left',
                    }}
                  >
                    <span>{s.label}</span>
                    {active && <Check size={14} strokeWidth={2.5} color="var(--accent)" />}
                  </button>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {query.isLoading ? (
        <div style={listContainerStyle(view)}>
          <FeedbackCardSkeleton view={view} />
          <FeedbackCardSkeleton view={view} />
          <FeedbackCardSkeleton view={view} />
          <FeedbackCardSkeleton view={view} />
        </div>
      ) : query.isError ? (
        <EmptyState
          title="Feedback unavailable"
          subtitle="We couldn't load feedback right now. Try again in a moment."
        />
      ) : items.length === 0 ? (
        <EmptyState
          title="Nothing here yet"
          subtitle={
            typeTab === 'all' && statusFilter === 'all'
              ? 'Be the first — send a bug report or feature request.'
              : 'Try clearing the filters above.'
          }
        />
      ) : (
        <div style={listContainerStyle(view)}>
          {items.map((feedback) => (
            <FeedbackCard key={feedback.id} feedback={feedback} view={view} />
          ))}
        </div>
      )}

      {!query.isLoading && !query.isError && items.length > 0 && (hasPrev || hasMore) && (
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            gap: 8,
            padding: '16px 20px 32px',
          }}
        >
          <button
            type="button"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={!hasPrev}
            className="btn-ghost"
            style={{
              padding: '8px 16px',
              borderRadius: 9999,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
              fontSize: 13,
              fontWeight: 600,
              cursor: hasPrev ? 'pointer' : 'not-allowed',
              opacity: hasPrev ? 1 : 0.4,
              fontFamily: 'inherit',
            }}
          >
            Previous
          </button>
          <span
            style={{
              alignSelf: 'center',
              fontSize: 12,
              color: 'var(--text-tertiary)',
              padding: '0 6px',
            }}
          >
            Page {page}
          </span>
          <button
            type="button"
            onClick={() => setPage((p) => p + 1)}
            disabled={!hasMore}
            className="btn-ghost"
            style={{
              padding: '8px 16px',
              borderRadius: 9999,
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text)',
              fontSize: 13,
              fontWeight: 600,
              cursor: hasMore ? 'pointer' : 'not-allowed',
              opacity: hasMore ? 1 : 0.4,
              fontFamily: 'inherit',
            }}
          >
            Next
          </button>
        </div>
      )}

      {composerOpen && (
        <FeedbackComposerModal
          busy={createMutation.isPending}
          errorMessage={composerError}
          onSubmit={submitComposer}
          onClose={() => {
            if (!createMutation.isPending) {
              setComposerOpen(false);
              setComposerError(null);
            }
          }}
        />
      )}
    </div>
  );
}

function listContainerStyle(view: ViewMode): CSSProperties {
  if (view === 'grid') {
    return {
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
      gap: 16,
      padding: '20px 20px 0',
    };
  }
  return {
    display: 'flex',
    flexDirection: 'column',
    gap: 10,
    padding: '16px 20px 0',
  };
}

interface ViewToggleProps {
  kind: ViewMode;
  active: boolean;
  onClick: () => void;
}

function ViewToggle({ kind, active, onClick }: ViewToggleProps) {
  const Glyph = kind === 'grid' ? LayoutGrid : List;
  return (
    <button
      type="button"
      aria-label={kind === 'grid' ? 'Grid view' : 'List view'}
      aria-pressed={active}
      onClick={onClick}
      style={{
        width: 30,
        height: 28,
        borderRadius: 9999,
        border: 'none',
        background: active ? 'var(--text)' : 'transparent',
        color: active ? 'var(--bg)' : 'var(--text-secondary)',
        cursor: 'pointer',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        transition: 'background 0.16s ease, color 0.16s ease',
      }}
    >
      <Glyph size={13} strokeWidth={2.2} />
    </button>
  );
}

interface TabButtonProps {
  label: string;
  active: boolean;
  onClick: () => void;
}

function TabButton({ label, active, onClick }: TabButtonProps) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      style={{
        flex: 1,
        padding: '14px 0',
        background: 'transparent',
        border: 'none',
        cursor: 'pointer',
        color: active ? 'var(--text)' : 'var(--text-tertiary)',
        fontFamily: 'inherit',
        fontWeight: active ? 700 : 500,
        fontSize: 14,
        position: 'relative',
        transition: 'color 0.12s ease',
      }}
    >
      {label}
      {active && (
        <span
          aria-hidden
          style={{
            position: 'absolute',
            bottom: 0,
            left: '50%',
            transform: 'translateX(-50%)',
            width: 48,
            height: 3,
            borderRadius: 2,
            background: 'var(--accent)',
          }}
        />
      )}
    </button>
  );
}

