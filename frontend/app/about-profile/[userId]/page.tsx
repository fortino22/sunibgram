'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft, Calendar, MapPin, BadgeCheck, AtSign, Globe, ChevronDown } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { VerifiedBadge } from '@/components/shared/VerifiedBadge';
import { useProfile } from '@/hooks/useProfile';
import { useUsernameHistory } from '@/hooks/useUsernameHistory';

function parseDate(value?: string | null): Date | null {
  if (!value) return null;
  const d = new Date(value.replace(' ', 'T'));
  return isNaN(d.getTime()) ? null : d;
}

function monthYear(value?: string | null): string {
  const d = parseDate(value);
  return d ? d.toLocaleDateString('en-US', { month: 'long', year: 'numeric' }) : '—';
}

function fullDate(value?: string | null): string {
  const d = parseDate(value);
  return d ? d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) : '—';
}

const rowStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'flex-start',
  gap: 16,
  padding: '16px 20px',
  width: '100%',
  background: 'transparent',
  border: 'none',
  textAlign: 'left',
  fontFamily: 'inherit',
};

const labelStyle: React.CSSProperties = { fontSize: 15, fontWeight: 600, color: 'var(--text)', lineHeight: 1.3 };
const valueStyle: React.CSSProperties = { fontSize: 14, color: 'var(--text-tertiary)', marginTop: 2 };

function Row({ icon, label, value, children }: { icon: React.ReactNode; label: string; value?: string; children?: React.ReactNode }) {
  return (
    <div style={rowStyle}>
      <span style={{ color: 'var(--text-tertiary)', flexShrink: 0, marginTop: 1 }}>{icon}</span>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={labelStyle}>{label}</div>
        {value && <div style={valueStyle}>{value}</div>}
        {children}
      </div>
    </div>
  );
}

export default function AboutAccountPage() {
  const params = useParams<{ userId: string }>();
  const router = useRouter();
  const routeUserId = Number(params.userId);
  const validId = Number.isInteger(routeUserId) && routeUserId > 0;
  const userId = validId ? routeUserId : undefined;

  const { data: profile, isLoading } = useProfile(userId);
  const { data: history } = useUsernameHistory(userId);
  const [historyOpen, setHistoryOpen] = useState(false);

  const count = history?.count ?? 0;
  const lastChange = history?.changes?.[0]?.changed_at;

  return (
    <div style={{ minHeight: '100vh' }}>
      {/* Header */}
      <div style={{ position: 'sticky', top: 0, zIndex: 10, display: 'flex', alignItems: 'center', gap: 18, padding: '12px 18px', background: 'color-mix(in oklab, var(--bg) 85%, transparent)', backdropFilter: 'blur(8px)' }}>
        <button type="button" onClick={() => router.back()} aria-label="Back" className="btn-ghost" style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text)', padding: 6, borderRadius: '50%', display: 'flex' }}>
          <ArrowLeft size={20} strokeWidth={2} />
        </button>
        <span style={{ fontWeight: 800, fontSize: 19, color: 'var(--text)', letterSpacing: '-0.02em' }}>About this account</span>
      </div>

      {isLoading || !profile ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)', fontSize: 14 }}>Loading…</div>
      ) : (
        <>
          {/* Identity */}
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 8, padding: '28px 20px 24px' }}>
            <Avatar username={profile.name ?? profile.username} avatarUrl={profile.avatar} size="lg" />
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 4 }}>
              <span style={{ fontWeight: 800, fontSize: 18, color: 'var(--text)', letterSpacing: '-0.02em' }}>{profile.name ?? profile.username}</span>
              {profile.is_premium && <VerifiedBadge size={18} />}
            </div>
            <span style={{ fontSize: 14, color: 'var(--text-tertiary)' }}>@{profile.username}</span>
          </div>

          {/* Rows */}
          <div>
            <Row icon={<Calendar size={20} strokeWidth={1.8} />} label="Date joined" value={monthYear(profile.created_at)} />
            <Row icon={<MapPin size={20} strokeWidth={1.8} />} label="Account based in" value={profile.faculty_location?.name ?? 'Indonesia'} />
            {profile.is_premium && (
              <Row icon={<BadgeCheck size={20} strokeWidth={1.8} />} label="Verified" value="This account is verified" />
            )}

            {/* Username changes — expandable */}
            <button
              type="button"
              onClick={() => count > 0 && setHistoryOpen((v) => !v)}
              className={count > 0 ? 'hover-dim' : undefined}
              style={{ ...rowStyle, cursor: count > 0 ? 'pointer' : 'default', alignItems: 'center' }}
            >
              <span style={{ color: 'var(--text-tertiary)', flexShrink: 0 }}><AtSign size={20} strokeWidth={1.8} /></span>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={labelStyle}>{count} username {count === 1 ? 'change' : 'changes'}</div>
                <div style={valueStyle}>{count > 0 ? `Last on ${monthYear(lastChange)}` : 'Never changed'}</div>
              </div>
              {count > 0 && (
                <ChevronDown size={18} strokeWidth={2} style={{ color: 'var(--text-tertiary)', flexShrink: 0, transform: historyOpen ? 'rotate(180deg)' : 'none', transition: 'transform 0.15s ease' }} />
              )}
            </button>

            {historyOpen && history && (
              <div style={{ padding: '0 20px 14px 56px', display: 'flex', flexDirection: 'column', gap: 12 }}>
                {history.changes.map((c, i) => (
                  <div key={i} style={{ fontSize: 14, color: 'var(--text)' }}>
                    {c.old_username || c.new_username ? (
                      <span>
                        <span style={{ color: 'var(--text-tertiary)' }}>@{c.old_username || '—'}</span>
                        {' → '}
                        <span style={{ fontWeight: 600 }}>@{c.new_username || '—'}</span>
                      </span>
                    ) : (
                      <span style={{ color: 'var(--text-tertiary)' }}>Username changed</span>
                    )}
                    <div style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 2 }}>{fullDate(c.changed_at)}</div>
                  </div>
                ))}
              </div>
            )}

            <Row icon={<Globe size={20} strokeWidth={1.8} />} label="Connected via" value="web" />
          </div>
        </>
      )}
    </div>
  );
}
