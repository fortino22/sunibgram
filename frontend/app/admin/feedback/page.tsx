'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { MessageSquareText } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { Pager } from '@/components/admin/Pager';
import { DetailModal, DetailRow } from '@/components/admin/DetailModal';
import { AdminFeedbackCard, formatRelativeTime } from '@/components/admin/AdminFeedbackCard';
import { AdminFeedbackTypeBadge, AdminFeedbackStatusBadge } from '@/components/admin/FeedbackBadges';
import { Avatar } from '@/components/shared/Avatar';
import { ImageZoomLightbox } from '@/components/shared/ImageZoomLightbox';
import { listFeedback, updateFeedbackStatus } from '@/lib/services/admin/feedback';
import type { Feedback } from '@/lib/services/admin/types';

const STATUS_OPTIONS = ['open', 'in_progress', 'resolved', 'dismissed'];

export default function AdminFeedbackPage() {
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const [statusFilter, setStatusFilter] = useState('');
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<Feedback | null>(null);
  const [lightbox, setLightbox] = useState<{ images: string[]; index: number; title: string } | null>(null);

  const q = useQuery({
    queryKey: ['admin', 'feedback', statusFilter, page],
    queryFn: () => listFeedback(token!, { status: statusFilter || undefined, page, limit: 20 }),
    enabled: !!token,
  });

  const statusMut = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) => updateFeedbackStatus(token!, id, status),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['admin', 'feedback'] });
      setSelected((cur) => (cur && cur.id === vars.id ? { ...cur, status: vars.status } : cur));
    },
  });

  const total = q.data?.meta?.total ?? 0;
  const limit = q.data?.meta?.limit ?? 20;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const rows = q.data?.data ?? [];

  function openImage(images: string[], index: number, title: string) {
    setLightbox({ images, index, title });
  }

  return (
    <AdminPage
      icon={MessageSquareText}
      title="Feedback"
      description="User-submitted feedback and bug reports."
      actions={
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          className="admin-select admin-select-pill"
        >
          <option value="">All statuses</option>
          {STATUS_OPTIONS.map((s) => <option key={s} value={s}>{s}</option>)}
        </select>
      }
    >
      {q.isLoading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>Loading…</div>
      ) : q.isError ? (
        <div style={{ padding: 40, color: 'var(--coral)' }}>{(q.error as Error).message}</div>
      ) : rows.length === 0 ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>No feedback</div>
      ) : (
        <>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(340px, 1fr))',
              gap: 14,
            }}
          >
            {rows.map((f) => (
              <AdminFeedbackCard
                key={f.id}
                feedback={f}
                onOpen={() => setSelected(f)}
                onImageClick={(images, index) => openImage(images, index, f.title)}
              />
            ))}
          </div>
          <Pager page={page} totalPages={totalPages} total={total} onChange={setPage} />
        </>
      )}

      <DetailModal
        open={!!selected}
        onClose={() => setSelected(null)}
        title={selected?.title ?? 'Feedback'}
        hero={
          selected && (selected.images?.length ?? 0) > 0 ? (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              {selected.images!.map((src, idx) => (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  key={src}
                  src={src}
                  alt="Feedback attachment"
                  loading="lazy"
                  onClick={() => openImage(selected.images!, idx, selected.title)}
                  style={{
                    width: 96,
                    height: 96,
                    objectFit: 'cover',
                    borderRadius: 10,
                    border: '1px solid var(--border)',
                    display: 'block',
                    cursor: 'zoom-in',
                  }}
                />
              ))}
            </div>
          ) : undefined
        }
      >
        {selected && (
          <>
            <DetailRow label="Type">
              <AdminFeedbackTypeBadge type={selected.type} />
            </DetailRow>
            <DetailRow label="Status">
              <select
                value={selected.status}
                onChange={(e) => statusMut.mutate({ id: selected.id, status: e.target.value })}
                className="admin-select admin-select-sm"
              >
                {STATUS_OPTIONS.map((s) => <option key={s} value={s}>{s}</option>)}
              </select>
            </DetailRow>
            <DetailRow label="Author">
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
                <Avatar
                  username={selected.author?.name ?? selected.author?.username ?? '?'}
                  avatarUrl={selected.author?.user_avatar}
                  size={28}
                />
                <span>
                  {selected.author?.name ?? selected.author?.username ?? 'Unknown'}
                  {selected.author?.username && (
                    <span style={{ color: 'var(--text-tertiary)' }}> @{selected.author.username}</span>
                  )}
                </span>
              </span>
            </DetailRow>
            <DetailRow label="Created">
              {selected.created_at
                ? `${new Date(selected.created_at).toLocaleString()} (${formatRelativeTime(selected.created_at)})`
                : '—'}
            </DetailRow>
            <DetailRow label="Votes">
              <span style={{ display: 'inline-flex', gap: 12 }}>
                <span>▲ {selected.likes ?? 0}</span>
                <span>▼ {selected.dislikes ?? 0}</span>
              </span>
            </DetailRow>
            <DetailRow label="Body">
              <span style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', lineHeight: 1.55 }}>
                {selected.body}
              </span>
            </DetailRow>
          </>
        )}
      </DetailModal>

      {lightbox && (
        <ImageZoomLightbox
          images={lightbox.images}
          index={lightbox.index}
          title={lightbox.title}
          onClose={() => setLightbox(null)}
          onNavigate={(index) => setLightbox((cur) => (cur ? { ...cur, index } : cur))}
        />
      )}
    </AdminPage>
  );
}
