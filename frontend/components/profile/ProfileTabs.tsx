'use client';

export type ProfileTab = 'threads' | 'replies' | 'media' | 'likes';

const TABS: { id: ProfileTab; label: string }[] = [
  { id: 'threads', label: 'Threads' },
  { id: 'replies', label: 'Replies' },
  { id: 'media', label: 'Media' },
  { id: 'likes', label: 'Likes' },
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
        borderBottom: '1px solid var(--border)',
        position: 'sticky',
        top: 77,
        zIndex: 10,
        background: 'oklch(10.5% 0.005 260 / 0.92)',
        backdropFilter: 'blur(12px)',
      }}
    >
      {TABS.map((tab) => {
        const isActive = activeTab === tab.id;
        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={isActive}
            onClick={() => onTabChange(tab.id)}
            className="btn-ghost"
            style={{
              flex: 1,
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'stretch',
              padding: 0,
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              fontSize: 14,
              fontWeight: isActive ? 700 : 400,
              color: isActive ? 'var(--text)' : 'var(--text-tertiary)',
              fontFamily: 'inherit',
              letterSpacing: '-0.01em',
              whiteSpace: 'nowrap',
            }}
          >
            <span
              style={{
                padding: '16px 0',
                borderBottom: isActive ? '2px solid var(--accent)' : '2px solid transparent',
                marginBottom: -1,
              }}
            >
              {tab.label}
            </span>
          </button>
        );
      })}
    </div>
  );
}
