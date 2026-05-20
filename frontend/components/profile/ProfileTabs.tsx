'use client';

export type ProfileTab = 'threads' | 'replies' | 'media' | 'likes' | 'followers' | 'following';

const TABS: { id: ProfileTab; label: string }[] = [
  { id: 'threads', label: 'Threads' },
  { id: 'replies', label: 'Replies' },
  { id: 'media', label: 'Media' },
  { id: 'likes', label: 'Likes' },
  { id: 'followers', label: 'Followers' },
  { id: 'following', label: 'Following' },
];

interface ProfileTabsProps {
  activeTab: ProfileTab;
  onTabChange: (tab: ProfileTab) => void;
}

export function ProfileTabs({ activeTab, onTabChange }: ProfileTabsProps) {
  return (
    <div
      role="tablist"
      style={{
        display: 'flex',
        gap: 4,
        overflowX: 'auto',
        borderBottom: '1px solid var(--border)',
        position: 'sticky',
        top: 77,
        zIndex: 10,
        background: 'oklch(10.5% 0.005 260 / 0.92)',
        backdropFilter: 'blur(12px)',
        padding: '0 12px',
      }}
    >
      {TABS.map((tab) => (
        <button
          key={tab.id}
          type="button"
          role="tab"
          aria-selected={activeTab === tab.id}
          onClick={() => onTabChange(tab.id)}
          style={{
            flexShrink: 0,
            padding: '14px 12px',
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            fontSize: 14,
            fontWeight: activeTab === tab.id ? 700 : 400,
            color: activeTab === tab.id ? 'var(--text)' : 'var(--text-tertiary)',
            borderBottom: activeTab === tab.id ? '2px solid var(--accent)' : '2px solid transparent',
            fontFamily: 'inherit',
            letterSpacing: '-0.01em',
            whiteSpace: 'nowrap',
            transition: 'all 0.15s ease',
            marginBottom: -1,
          }}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
