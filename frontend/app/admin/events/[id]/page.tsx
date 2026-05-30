'use client';

import { use, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { CalendarDays, Check, MapPin } from 'lucide-react';
import { toast } from 'sonner';
import { useAuthStore } from '@/stores/authStore';
import { AdminPage } from '@/components/admin/AdminPage';
import { StatusBadge } from '@/components/admin/StatusBadge';
import { EventDocViewer } from '@/components/admin/DocViewer';
import { ConfirmModal } from '@/components/admin/ConfirmModal';
import {
  getEvent,
  approveEvent,
  declineEvent,
  requestRevision,
  deleteEvent,
} from '@/lib/services/admin/events';

export default function AdminEventDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const token = useAuthStore((s) => s.token);
  const qc = useQueryClient();
  const router = useRouter();
  const [declineReason, setDeclineReason] = useState('');
  const [revisionNote, setRevisionNote] = useState('');
  const [confirmDelete, setConfirmDelete] = useState(false);

  const q = useQuery({
    queryKey: ['admin', 'event', id],
    queryFn: () => getEvent(token!, id),
    enabled: !!token,
  });

  const invalidate = () => {
    qc.invalidateQueries({ queryKey: ['admin', 'event', id] });
    qc.invalidateQueries({ queryKey: ['admin', 'events'] });
  };

  const approveMut = useMutation({
    mutationFn: () => approveEvent(token!, id),
    onSuccess: () => { invalidate(); toast.success('Event approved'); },
    onError: (err: Error) => toast.error(err.message),
  });
  const declineMut = useMutation({
    mutationFn: () => declineEvent(token!, id, declineReason),
    onSuccess: () => { invalidate(); setDeclineReason(''); toast.success('Event declined'); },
    onError: (err: Error) => toast.error(err.message),
  });
  const revisionMut = useMutation({
    mutationFn: () => requestRevision(token!, id, revisionNote),
    onSuccess: () => { invalidate(); setRevisionNote(''); toast.success('Revision request sent'); },
    onError: (err: Error) => toast.error(err.message),
  });
  const deleteMut = useMutation({
    mutationFn: () => deleteEvent(token!, id),
    onSuccess: () => { toast.success('Event deleted'); router.replace('/admin/events'); },
    onError: (err: Error) => toast.error(err.message),
  });

  if (q.isLoading) {
    return <AdminPage title="Loading…"><div /></AdminPage>;
  }
  if (q.isError || !q.data) {
    return <AdminPage title="Event"><div style={{ color: 'var(--coral)' }}>Failed to load.</div></AdminPage>;
  }

  const e = q.data;
  const docUrls = e.doc_urls ?? (e.attachment_url ? [e.attachment_url] : []);
  const isPending = e.status === 'pending';

  return (
    <AdminPage
      icon={CalendarDays}
      title={e.title}
      description={`${e.type.toUpperCase()} · submitted by user #${e.user_id}`}
      actions={
        <Link
          href="/admin/events"
          className="admin-btn admin-btn-outline admin-btn-sm"
          style={{ textDecoration: 'none' }}
        >
          ← Back
        </Link>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'center' }}>
          <StatusBadge status={e.status} />
          <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>
            {new Date(e.start_time).toLocaleString()} – {new Date(e.end_time).toLocaleString()}
          </span>
          {e.location && (
            <span style={{ fontSize: 13, color: 'var(--text-tertiary)', display: 'flex', alignItems: 'center', gap: 4 }}>
              <MapPin size={13} />
              {e.location}
            </span>
          )}
        </div>

        {e.banner_url && (
          <img
            src={e.banner_url}
            alt="Banner"
            style={{ width: '100%', maxHeight: 320, objectFit: 'cover', borderRadius: 16, border: '1px solid var(--border)' }}
          />
        )}

        <div style={{ padding: 18, borderRadius: 16, border: '1px solid var(--border)', background: 'var(--bg-card)' }}>
          <h3 style={{ fontSize: 14, fontWeight: 600, marginBottom: 8, color: 'var(--text-secondary)' }}>Description</h3>
          <p style={{ fontSize: 14, lineHeight: 1.7, whiteSpace: 'pre-wrap', color: 'var(--text)' }}>
            {e.description || '—'}
          </p>
        </div>

        <div>
          <h3 style={{ fontSize: 16, fontWeight: 600, marginBottom: 12, color: 'var(--text)' }}>Documents</h3>
          <EventDocViewer urls={docUrls} />
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 16 }}>
          <ActionCard title="Approve">
            {e.status === 'approved' ? (
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  padding: '10px 14px',
                  borderRadius: 10,
                  background: 'var(--green-soft)',
                  color: 'var(--green)',
                  fontSize: 13,
                  fontWeight: 600,
                }}
              >
                <Check size={16} strokeWidth={2.4} />
                Event approved
              </div>
            ) : !isPending ? (
              <p style={{ fontSize: 13, color: 'var(--text-tertiary)', margin: 0 }}>
                Only pending events can be approved.
              </p>
            ) : (
              <>
                <p style={{ fontSize: 13, color: 'var(--text-secondary)', marginBottom: 12 }}>
                  Mark this event as approved. It will appear on the public events page.
                </p>
                <button
                  type="button"
                  onClick={() => approveMut.mutate()}
                  disabled={approveMut.isPending}
                  className="admin-btn admin-btn-success"
                >
                  {approveMut.isPending ? 'Working…' : 'Approve Event'}
                </button>
              </>
            )}
          </ActionCard>

          <ActionCard title="Request Revision">
            <textarea
              value={revisionNote}
              onChange={(ev) => setRevisionNote(ev.target.value)}
              placeholder="Tell the organizer what to revise…"
              className="admin-textarea"
              style={{ minHeight: 90, marginBottom: 10 }}
            />
            <button
              type="button"
              onClick={() => revisionMut.mutate()}
              disabled={revisionMut.isPending || !revisionNote.trim()}
              className="admin-btn admin-btn-primary"
            >
              {revisionMut.isPending ? 'Sending…' : 'Send Revision Request'}
            </button>
          </ActionCard>

          <ActionCard title="Decline">
            {e.status === 'declined' ? (
              <p style={{ fontSize: 13, color: 'var(--text-tertiary)', margin: 0 }}>
                This event has been declined.
              </p>
            ) : !isPending ? (
              <p style={{ fontSize: 13, color: 'var(--text-tertiary)', margin: 0 }}>
                Only pending events can be declined.
              </p>
            ) : (
              <>
                <textarea
                  value={declineReason}
                  onChange={(ev) => setDeclineReason(ev.target.value)}
                  placeholder="Reason for declining…"
                  className="admin-textarea"
                  style={{ minHeight: 90, marginBottom: 10 }}
                />
                <button
                  type="button"
                  onClick={() => declineMut.mutate()}
                  disabled={declineMut.isPending || !declineReason.trim()}
                  className="admin-btn admin-btn-outline-danger"
                >
                  {declineMut.isPending ? 'Declining…' : 'Decline Event'}
                </button>
              </>
            )}
          </ActionCard>
        </div>

        <div>
          <h3 style={{ fontSize: 16, fontWeight: 600, marginBottom: 12, color: 'var(--text)' }}>
            Review History
          </h3>
          {(!e.notes || e.notes.length === 0) ? (
            <div style={{ padding: 20, color: 'var(--text-tertiary)', textAlign: 'center', border: '1px dashed var(--border)', borderRadius: 12, fontSize: 14 }}>
              No review history.
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {e.notes.map((n) => (
                <div
                  key={n.revision_number}
                  style={{ padding: 14, borderRadius: 12, border: '1px solid var(--border)', background: 'var(--bg-card)' }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}>
                    <span style={{ fontWeight: 600, fontSize: 13 }}>
                      Review #{n.revision_number}{n.admin_name ? ` · ${n.admin_name}` : ''}
                    </span>
                    <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>
                      {new Date(n.created_at).toLocaleString()}
                    </span>
                  </div>
                  <p style={{ fontSize: 14, color: 'var(--text)', whiteSpace: 'pre-wrap', lineHeight: 1.6, margin: 0 }}>
                    {n.notes}
                  </p>
                </div>
              ))}
            </div>
          )}
        </div>

        <div style={{ paddingTop: 12, borderTop: '1px solid var(--border)', display: 'flex', justifyContent: 'flex-end' }}>
          <button
            type="button"
            onClick={() => setConfirmDelete(true)}
            className="admin-btn admin-btn-outline-danger"
          >
            Delete Event
          </button>
        </div>
      </div>

      <ConfirmModal
        open={confirmDelete}
        title="Delete this event?"
        description="The event and all its review history will be removed."
        destructive
        confirmLabel="Delete"
        busy={deleteMut.isPending}
        onConfirm={() => deleteMut.mutate()}
        onClose={() => setConfirmDelete(false)}
      />
    </AdminPage>
  );
}

function ActionCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div style={{ padding: 18, borderRadius: 16, border: '1px solid var(--border)', background: 'var(--bg-card)' }}>
      <h4 style={{ fontSize: 14, fontWeight: 700, marginBottom: 10, color: 'var(--text)' }}>{title}</h4>
      {children}
    </div>
  );
}
