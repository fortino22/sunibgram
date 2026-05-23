'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Users, Globe, Calendar } from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';
import { useCommunityAbout } from '@/hooks/useCommunities';
import type { CommunityAboutMember } from '@/lib/services/communities';

interface CommunityAboutTabProps {
  communityId: number;
  memberCount: number;
}

function SkeletonBlock({ height = 80 }: { height?: number }) {
  return (
    <div
      style={{
        height,
        background: 'var(--bg-subtle)',
        margin: '0 0 1px',
        animation: 'pulse 1.5s ease-in-out infinite',
      }}
    />
  );
}

function Divider() {
  return <div style={{ height: 1, background: 'var(--border)' }} />;
}

function formatCommunityDate(dateStr: string): string {
  try {
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return dateStr;
    return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'long', year: 'numeric' });
  } catch {
    return dateStr;
  }
}

export function CommunityAboutTab({ communityId, memberCount }: CommunityAboutTabProps) {
  const router = useRouter();
  const { data: about, isLoading } = useCommunityAbout(communityId);

  if (isLoading) {
    return (
      <div style={{ paddingBottom: 24 }}>
        <SkeletonBlock height={140} />
        <SkeletonBlock height={180} />
        <SkeletonBlock height={160} />
      </div>
    );
  }

  if (!about) return null;

  const topMembersToShow = about.top_members.slice(0, 5);

  return (
    <div style={{ paddingBottom: 32 }}>
      {/* Community Info */}
      <div style={{ padding: '18px 16px 22px' }}>
        <div style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)', marginBottom: 14 }}>
          Community Info
        </div>

        {about.info.description && (
          <p style={{ fontSize: 15, color: 'var(--text)', lineHeight: 1.6, margin: '0 0 16px' }}>
            {about.info.description}
          </p>
        )}

        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <PolicyRow
            icon={<Users size={16} />}
            text={about.info.posting_policy}
          />
          <PolicyRow
            icon={<Globe size={16} />}
            text={
              [about.info.visibility_policy, about.info.join_policy]
                .filter(Boolean)
                .join(' ')
            }
          />
          <PolicyRow
            icon={<Calendar size={16} />}
            text={
              <span>
                {'Created '}
                {formatCommunityDate(about.creator.community_created_at)}
                {' by '}
                <Link
                  href={`/profile/${about.creator.user_id}`}
                  style={{ color: 'var(--text)', fontWeight: 600, textDecoration: 'none' }}
                >
                  @{about.creator.username}
                </Link>
              </span>
            }
          />
        </div>
      </div>

      <Divider />

      {/* Rules */}
      {about.rules.length > 0 && (
        <>
          <div style={{ padding: '18px 16px 22px' }}>
            <div style={{ fontSize: 17, fontWeight: 700, color: 'var(--text)', marginBottom: 4 }}>
              Rules
            </div>
            <p style={{ fontSize: 13, color: 'var(--text-tertiary)', margin: '0 0 16px', lineHeight: 1.5 }}>
              These are set and enforced by Community admins.
            </p>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
              {about.rules.map((rule, i) => (
                <div key={rule.id} style={{ display: 'flex', gap: 14, alignItems: 'flex-start' }}>
                  <span
                    style={{
                      flexShrink: 0,
                      fontSize: 15,
                      fontWeight: 700,
                      color: 'var(--text)',
                      minWidth: 16,
                      lineHeight: 1.4,
                    }}
                  >
                    {i + 1}
                  </span>
                  <span style={{ fontSize: 15, color: 'var(--text)', lineHeight: 1.4 }}>
                    {rule.text}
                  </span>
                </div>
              ))}
            </div>
          </div>
          <Divider />
        </>
      )}

      {/* Moderators */}
      {about.admins.length > 0 && (
        <>
          <div style={{ padding: '18px 0 6px' }}>
            <div
              style={{
                fontSize: 17,
                fontWeight: 700,
                color: 'var(--text)',
                padding: '0 16px',
                marginBottom: 6,
              }}
            >
              Admins
            </div>
            {about.admins.map((admin, i) => (
              <div key={admin.user_id}>
                <MemberRow member={admin} />
                {i < about.admins.length - 1 && (
                  <div style={{ height: 1, background: 'var(--border)', margin: '0 16px' }} />
                )}
              </div>
            ))}
          </div>
          <Divider />
        </>
      )}

      {/* Members */}
      <div style={{ padding: '18px 0 6px' }}>
        <div
          style={{
            fontSize: 17,
            fontWeight: 700,
            color: 'var(--text)',
            padding: '0 16px',
            marginBottom: 6,
          }}
        >
          Members
        </div>

        {topMembersToShow.map((member, i) => (
          <div key={member.user_id}>
            <MemberRow member={member} />
            {i < topMembersToShow.length - 1 && (
              <div style={{ height: 1, background: 'var(--border)', margin: '0 16px' }} />
            )}
          </div>
        ))}

        <div style={{ padding: '14px 16px 8px', display: 'flex', flexDirection: 'column', gap: 10 }}>
          <button
            onClick={() => router.push(`/communities/${communityId}/members`)}
            style={{
              width: '100%',
              padding: '10px 0',
              background: 'none',
              border: '1px solid var(--border)',
              borderRadius: 999,
              fontSize: 14,
              fontWeight: 600,
              color: 'var(--text)',
              cursor: 'pointer',
            }}
          >
            Show more
          </button>

        </div>
      </div>
    </div>
  );
}

function PolicyRow({ icon, text }: { icon: React.ReactNode; text: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
      <span style={{ flexShrink: 0, marginTop: 2, color: 'var(--text-tertiary)', display: 'flex' }}>
        {icon}
      </span>
      <span style={{ fontSize: 14, color: 'var(--text-secondary)', lineHeight: 1.5 }}>{text}</span>
    </div>
  );
}

function MemberRow({ member }: { member: CommunityAboutMember }) {
  return (
    <Link
      href={`/profile/${member.user_id}`}
      style={{
        display: 'flex',
        gap: 12,
        alignItems: 'center',
        textDecoration: 'none',
        padding: '12px 16px',
      }}
    >
      <Avatar username={member.username} avatarUrl={member.avatar_url || null} size={44} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 5, flexWrap: 'wrap' }}>
          <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--text)' }}>
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
        <div style={{ fontSize: 13, color: 'var(--text-tertiary)' }}>@{member.username}</div>
        {member.bio && (
          <div
            style={{
              fontSize: 13,
              color: 'var(--text-secondary)',
              marginTop: 2,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {member.bio}
          </div>
        )}
      </div>
    </Link>
  );
}
