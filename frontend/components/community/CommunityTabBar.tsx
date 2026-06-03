'use client';

import { useEffect, useRef, useState } from 'react';
import { ChevronDown, Check } from 'lucide-react';

export interface TabItem {
  id: string;
  label: string;
}

interface CommunityTabBarProps {
  tabs: TabItem[];
  activeId: string;
  onSelect: (id: string) => void;
  /** Max chips shown inline (including the "More" button). Default 5. */
  maxVisible?: number;
}

/**
 * Horizontal tab strip that never shows more than `maxVisible` chips. Extra tabs
 * collapse into a "More ▾" dropdown. If the active tab lives in the overflow it
 * is promoted into the visible row so it stays highlighted.
 */
export function CommunityTabBar({ tabs, activeId, onSelect, maxVisible = 5 }: CommunityTabBarProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    function onClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }
    window.addEventListener('mousedown', onClick);
    return () => window.removeEventListener('mousedown', onClick);
  }, [menuOpen]);

  let visible = tabs;
  let overflow: TabItem[] = [];
  if (tabs.length > maxVisible) {
    // Reserve the last slot for the "More" button → show first (maxVisible - 1).
    const inlineCount = maxVisible - 1;
    visible = tabs.slice(0, inlineCount);
    overflow = tabs.slice(inlineCount);
    // Promote the active overflow tab into the visible row.
    const activeInOverflow = overflow.find((t) => t.id === activeId);
    if (activeInOverflow) {
      const demoted = visible[visible.length - 1];
      visible = [...visible.slice(0, -1), activeInOverflow];
      overflow = [demoted, ...overflow.filter((t) => t.id !== activeId)];
    }
  }

  const overflowActive = overflow.some((t) => t.id === activeId);

  function tabStyle(active: boolean): React.CSSProperties {
    return {
      flex: 1,
      width: '100%',
      position: 'relative',
      padding: '12px 0',
      fontSize: 14,
      fontWeight: active ? 700 : 400,
      color: active ? 'var(--text)' : 'var(--text-tertiary)',
      background: 'none',
      border: 'none',
      cursor: 'pointer',
      whiteSpace: 'nowrap',
      textAlign: 'center',
      borderBottom: active ? '2px solid var(--text)' : '2px solid transparent',
    };
  }

  return (
    <div style={{ display: 'flex', borderBottom: '1px solid var(--border)', position: 'relative' }}>
      {visible.map((tab) => (
        <button key={tab.id} onClick={() => onSelect(tab.id)} className="hover-dim" style={tabStyle(tab.id === activeId)}>
          {tab.label}
        </button>
      ))}

      {overflow.length > 0 && (
        <div ref={menuRef} style={{ position: 'relative', flex: 1, display: 'flex' }}>
          <button
            onClick={() => setMenuOpen((v) => !v)}
            className="hover-dim"
            style={{
              ...tabStyle(overflowActive),
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 4,
            }}
          >
            More
            <ChevronDown size={15} />
          </button>
          {menuOpen && (
            <div
              role="menu"
              className="tab-overflow-menu"
              style={{
                position: 'absolute',
                top: 'calc(100% + 8px)',
                right: 0,
                minWidth: 200,
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 14,
                boxShadow: 'var(--shadow-dropdown)',
                padding: 8,
                zIndex: 30,
              }}
            >
              {overflow.map((tab) => {
                const active = tab.id === activeId;
                return (
                  <button
                    key={tab.id}
                    role="menuitem"
                    className="tab-overflow-item"
                    data-active={active}
                    onClick={() => {
                      onSelect(tab.id);
                      setMenuOpen(false);
                    }}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      gap: 10,
                      width: '100%',
                      textAlign: 'left',
                      padding: '11px 14px',
                      fontSize: 14,
                      fontWeight: active ? 700 : 500,
                      color: active ? 'var(--accent)' : 'var(--text-secondary)',
                      border: 'none',
                      borderRadius: 10,
                      cursor: 'pointer',
                    }}
                  >
                    <span>{tab.label}</span>
                    {active && <Check size={16} />}
                  </button>
                );
              })}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
