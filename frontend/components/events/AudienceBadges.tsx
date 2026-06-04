'use client';

import { Globe, MapPin } from 'lucide-react';
import type { EventAudience } from '@/lib/services/events';

interface Props {
  audience?: EventAudience[];
  max?: number;
}

// AudienceBadges renders an event's region audience as compact chips. An
// "All Regions" audience collapses to a single globe chip.
export function AudienceBadges({ audience, max = 3 }: Props) {
  const list = audience ?? [];
  if (list.length === 0) return null;

  const allRegions = list.find((a) => a.isAllRegions);
  const chip = (key: string, icon: React.ReactNode, label: string) => (
    <span
      key={key}
      style={{
        display: 'inline-flex', alignItems: 'center', gap: 4,
        fontSize: 11, fontWeight: 600, color: 'var(--text-secondary)',
        background: 'var(--bg-hover)', border: '1px solid var(--border)',
        padding: '2px 8px', borderRadius: 999, whiteSpace: 'nowrap',
      }}
    >
      {icon}
      {label}
    </span>
  );

  if (allRegions) {
    return (
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
        {chip('all', <Globe size={11} strokeWidth={1.8} />, 'All regions')}
      </div>
    );
  }

  const shown = list.slice(0, max);
  const extra = list.length - shown.length;
  return (
    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
      {shown.map((a) => chip(String(a.id), <MapPin size={11} strokeWidth={1.8} />, a.name))}
      {extra > 0 && chip('more', null, `+${extra}`)}
    </div>
  );
}
