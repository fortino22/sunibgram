'use client';

export type LeaderboardTab = 'trending' | 'stats';

const TABS: { id: LeaderboardTab; label: string }[] = [
  { id: 'trending', label: 'Trending' },
  { id: 'stats', label: 'Your Stats' },
];

interface LeaderboardTabsProps {
  activeTab: LeaderboardTab;
  onTabChange: (tab: LeaderboardTab) => void;
}

export function LeaderboardTabs({ activeTab, onTabChange }: LeaderboardTabsProps) {
  return (
    <div
      role="tablist"
      style={{
        display: 'flex',
        borderBottom: '1px solid var(--border)',
        margin: '0 -24px',
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
              border: 'none',
              background: 'none',
              cursor: 'pointer',
              fontSize: 14,
              fontWeight: isActive ? 700 : 400,
              color: isActive ? 'var(--text)' : 'var(--text-tertiary)',
              fontFamily: 'inherit',
              letterSpacing: '-0.01em',
            }}
          >
            <span
              style={{
                padding: '14px 0',
                borderBottom: isActive ? '2px solid var(--text)' : '2px solid transparent',
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
