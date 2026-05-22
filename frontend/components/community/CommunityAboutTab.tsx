'use client';

import Link from 'next/link';
import { Avatar } from '@/components/shared/Avatar';
import { useCommunityAbout } from '@/hooks/useCommunities';
import type { CommunityRole } from '@/lib/services/communities';

interface CommunityAboutTabProps {
  communityId: number;
  memberCount: number;
  myRole: CommunityRole | null;
  onManageMembers: () => void;
}

const SECTION: React.CSSProperties = {
  margin: '12px 16px 0',
  padding: '14px 16px',
  borderRadius: 12,
  border: '1px solid var(--border)',
  background: 'var(--bg-subtle)',
};

const LABEL: React.CSSProperties = {
  fontSize: 11,
  fontWeight: 700,
  color: 'var(--text-tertiary)',
  letterSpacing: '0.06em',
  textTransform: 'uppercase',
  marginBottom: 10,
};

function SkeletonBlock({ height = 80 }: { height?: number }) {
  return (
    <div
      style={{
        ...SECTION,
        height,
        background: 'var(--bg-subtle)',
        borderRadius: 12,
        animation: 'pulse 1.5s ease-in-out infinite',
      }}
    />
  );
}

export function CommunityAboutTab({ communityId, memberCount, myRole, onManageMembers }: CommunityAboutTabProps) {
  const { data: about, isLoading } = useCommunityAbout(communityId);
  const isModOrAbove = myRole === 'MODERATOR' || myRole === 'ADMIN' || myRole === 'OWNER';

  if (isLoading) {
    return (
      <div style={{ paddingBottom: 24 }}>
        <SkeletonBlock height={90} />
        <SkeletonBlock height={110} />
        <SkeletonBlock height={80} />
      </div>
    );
  }

  if (!about) return null;

  return (
    <div style={{ paddingBottom: 32 }}>
      {/* Created by */}
      <div style={SECTION}>
        <div style={LABEL}>Created by</div>
        <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
          <Avatar username={about.creator.username} avatarUrl={about.creator.avatar_url || null} size="md" />
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
              <Link
                href={`/profile/${about.creator.user_id}`}
                style={{ fontWeight: 700, fontSize: 14, color: 'var(--text)', textDecoration: 'none' }}
              >
                {about.creator.name || about.creator.username}
              </Link>
              {about.creator.is_premium && (
                <span
                  style={{
                    fontSize: 10,
                    background: 'var(--accent-soft)',
                    color: 'var(--accent)',
                    padding: '1px 5px',
                    borderRadius: 4,
                    fontWeight: 600,
                  }}
                >
                  PRO
                </span>
              )}
              <span style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@{about.creator.username}</span>
            </div>
            {about.creator.bio && (
              <p style={{ fontSize: 13, color: 'var(--text-secondary)', margin: '4px 0 0', lineHeight: 1.5 }}>
                {about.creator.bio}
              </p>
            )}
            <p style={{ fontSize: 12, color: 'var(--text-tertiary)', margin: '4px 0 0' }}>
              Community since {about.creator.community_created_at}
            </p>
          </div>
        </div>
      </div>

      {/* About info */}
      <div style={SECTION}>
        <div style={LABEL}>About</div>
        {about.info.description && (
          <p style={{ fontSize: 14, color: 'var(--text)', lineHeight: 1.6, margin: '0 0 12px' }}>
            {about.info.description}
          </p>
        )}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <PolicyRow icon="👥" text={about.info.posting_policy} />
          <PolicyRow icon="🌐" text={about.info.visibility_policy} />
          <PolicyRow icon="🚪" text={about.info.join_policy} />
        </div>
      </div>

      {/* Rules */}
      {about.rules.length > 0 && (
        <div style={SECTION}>
          <div style={{ ...LABEL, display: 'flex', justifyContent: 'space-between' }}>
            <span>Rules</span>
            <span style={{ fontWeight: 400, textTransform: 'none', letterSpacing: 0 }}>
              {about.rules.length} {about.rules.length === 1 ? 'rule' : 'rules'}
            </span>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {about.rules.map((rule, i) => (
              <div key={rule.id} style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                <span
                  style={{
                    flexShrink: 0,
                    width: 22,
                    height: 22,
                    borderRadius: '50%',
                    background: 'var(--border)',
                    color: 'var(--text-secondary)',
                    fontSize: 11,
                    fontWeight: 700,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  {i + 1}
                </span>
                <span style={{ fontSize: 13, color: 'var(--text)', lineHeight: 1.5, paddingTop: 3 }}>{rule.text}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Admins */}
      {about.admins.length > 0 && (
        <div style={SECTION}>
          <div style={LABEL}>Admins</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {about.admins.map((admin) => (
              <MemberRow key={admin.user_id} member={admin} />
            ))}
          </div>
        </div>
      )}

      {/* Members */}
      <div style={SECTION}>
        <div style={{ ...LABEL, display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
          <span>Members</span>
          <span style={{ fontWeight: 400, textTransform: 'none', letterSpacing: 0, fontSize: 12 }}>
            {memberCount.toLocaleString()} total
          </span>
        </div>
        {about.top_members.length > 0 && (
          <div style={{ display: 'flex', gap: -4, marginBottom: 12 }}>
            {about.top_members.slice(0, 8).map((m, i) => (
              <div key={m.user_id} style={{ marginLeft: i === 0 ? 0 : -8, zIndex: about.top_members.length - i }}>
                <Avatar username={m.username} avatarUrl={m.avatar_url || null} size="sm" />
              </div>
            ))}
            {memberCount > 8 && (
              <div
                style={{
                  marginLeft: -8,
                  width: 32,
                  height: 32,
                  borderRadius: '50%',
                  background: 'var(--border)',
                  border: '2px solid var(--bg-card)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 10,
                  fontWeight: 700,
                  color: 'var(--text-secondary)',
                }}
              >
                +{memberCount - 8 > 999 ? '999' : memberCount - 8}
              </div>
            )}
          </div>
        )}
        {isModOrAbove && (
          <button
            onClick={onManageMembers}
            style={{
              background: 'none',
              border: '1px solid var(--border)',
              borderRadius: 8,
              padding: '7px 14px',
              fontSize: 13,
              fontWeight: 600,
              color: 'var(--text)',
              cursor: 'pointer',
              width: '100%',
            }}
          >
            Manage members
          </button>
        )}
      </div>
    </div>
  );
}

function PolicyRow({ icon, text }: { icon: string; text: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'flex-start', gap: 8 }}>
      <span style={{ fontSize: 14, lineHeight: 1.5 }}>{icon}</span>
      <span style={{ fontSize: 13, color: 'var(--text-secondary)', lineHeight: 1.5 }}>{text}</span>
    </div>
  );
}

function MemberRow({ member }: { member: { user_id: number; username: string; name: string; avatar_url: string; is_premium: boolean } }) {
  return (
    <Link
      href={`/profile/${member.user_id}`}
      style={{ display: 'flex', gap: 10, alignItems: 'center', textDecoration: 'none' }}
    >
      <Avatar username={member.username} avatarUrl={member.avatar_url || null} size="sm" />
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
          <span style={{ fontSize: 14, fontWeight: 700, color: 'var(--text)' }}>
            {member.name || member.username}
          </span>
          {member.is_premium && (
            <span
              style={{
                fontSize: 10,
                background: 'var(--accent-soft)',
                color: 'var(--accent)',
                padding: '1px 5px',
                borderRadius: 4,
                fontWeight: 600,
              }}
            >
              PRO
            </span>
          )}
        </div>
        <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>@{member.username}</span>
      </div>
    </Link>
  );
}
