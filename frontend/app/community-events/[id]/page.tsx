'use client';

import { useParams } from 'next/navigation';
import { CommunityEventDetail } from '@/components/community-events/CommunityEventDetail';
import { EmptyState } from '@/components/shared/EmptyState';

export default function CommunityEventDetailPage() {
  const params = useParams<{ id: string }>();
  const eventId = Number(params.id);
  const isValidId = Number.isInteger(eventId) && eventId > 0;

  if (!isValidId) {
    return <EmptyState title="Event not found" subtitle="Invalid event id." />;
  }

  return <CommunityEventDetail eventId={eventId} />;
}
