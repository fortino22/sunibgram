'use client';

import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import { CommunityEventDetail } from '@/components/community-events/CommunityEventDetail';
import { EmptyState } from '@/components/shared/EmptyState';

export default function CommunityEventDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const eventId = Number(params.id);
  const isValidId = Number.isInteger(eventId) && eventId > 0;

  if (!isValidId) {
    return <EmptyState title="Event not found" subtitle="Invalid event id." />;
  }

  return (
    <div style={{ maxWidth: 720, margin: '0 auto', padding: '16px' }}>
      <button
        onClick={() => router.back()}
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: 6,
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          color: 'var(--text-secondary)',
          fontSize: 14,
          marginBottom: 16,
          padding: 0,
        }}
      >
        <ArrowLeft size={18} />
        Back
      </button>
      <CommunityEventDetail eventId={eventId} />
    </div>
  );
}
