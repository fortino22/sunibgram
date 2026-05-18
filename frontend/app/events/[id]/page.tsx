'use client';

import { useParams } from 'next/navigation';
import { EventDetailView } from '@/components/events/EventDetailView';

export default function EventPage() {
  const { id } = useParams<{ id: string }>();
  return <EventDetailView eventId={id} />;
}
