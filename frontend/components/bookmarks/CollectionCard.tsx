'use client';

import Link from 'next/link';
import { useState } from 'react';
import { Bookmark, MoreHorizontal, Pencil, Trash2 } from 'lucide-react';
import type { BookmarkCollection } from '@/lib/services/bookmarkCollections';

interface CollectionCardProps {
  collection: BookmarkCollection;
  onEdit?: (collection: BookmarkCollection) => void;
  onDelete?: (collection: BookmarkCollection) => void;
}

export function CollectionCard({ collection, onEdit, onDelete }: CollectionCardProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [hovered, setHovered] = useState(false);

  const itemLabel = collection.item_count === 1 ? '1 post' : `${collection.item_count} posts`;

  return (
    <div
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => { setHovered(false); setMenuOpen(false); }}
      style={{
        position: 'relative',
        border: '1px solid var(--border)',
        borderRadius: 16,
        background: hovered ? 'var(--bg-hover)' : 'var(--bg-card)',
        transition: 'background 0.12s ease',
      }}
    >
      <Link
        href={`/bookmarks/collections/${collection.id}`}
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: 8,
          padding: '16px 18px',
          textDecoration: 'none',
          color: 'inherit',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <span
            aria-hidden
            style={{
              width: 36,
              height: 36,
              borderRadius: 10,
              background: 'var(--accent-soft)',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'var(--accent)',
              flexShrink: 0,
            }}
          >
            <Bookmark size={18} strokeWidth={2} />
          </span>
          <div style={{ minWidth: 0, flex: 1 }}>
            <div
              style={{
                fontSize: 15,
                fontWeight: 600,
                color: 'var(--text)',
                letterSpacing: '-0.01em',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {collection.name}
            </div>
            <div style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 2 }}>
              {itemLabel}
            </div>
          </div>
        </div>
        {collection.description && (
          <div
            style={{
              fontSize: 13,
              color: 'var(--text-secondary)',
              lineHeight: 1.5,
              display: '-webkit-box',
              WebkitLineClamp: 2,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
            }}
          >
            {collection.description}
          </div>
        )}
      </Link>

      {(onEdit || onDelete) && (
        <div style={{ position: 'absolute', top: 10, right: 10 }}>
          <button
            type="button"
            aria-label="Collection actions"
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              setMenuOpen((v) => !v);
            }}
            style={{
              width: 32,
              height: 32,
              borderRadius: 9999,
              background: menuOpen ? 'var(--bg-hover)' : 'transparent',
              border: 'none',
              color: 'var(--text-tertiary)',
              cursor: 'pointer',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <MoreHorizontal size={18} />
          </button>
          {menuOpen && (
            <div
              role="menu"
              style={{
                position: 'absolute',
                top: 36,
                right: 0,
                minWidth: 160,
                background: 'var(--bg-card)',
                border: '1px solid var(--border)',
                borderRadius: 12,
                padding: 6,
                boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
                zIndex: 20,
              }}
            >
              {onEdit && (
                <MenuButton
                  icon={<Pencil size={14} />}
                  label="Edit"
                  onClick={() => { setMenuOpen(false); onEdit(collection); }}
                />
              )}
              {onDelete && (
                <MenuButton
                  icon={<Trash2 size={14} />}
                  label="Delete"
                  destructive
                  onClick={() => { setMenuOpen(false); onDelete(collection); }}
                />
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

interface MenuButtonProps {
  icon: React.ReactNode;
  label: string;
  destructive?: boolean;
  onClick: () => void;
}

function MenuButton({ icon, label, destructive, onClick }: MenuButtonProps) {
  return (
    <button
      type="button"
      role="menuitem"
      onClick={(e) => { e.preventDefault(); e.stopPropagation(); onClick(); }}
      onMouseEnter={(e) => { e.currentTarget.style.background = destructive ? 'var(--coral-soft)' : 'var(--bg-hover)'; }}
      onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; }}
      style={{
        width: '100%',
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        padding: '9px 12px',
        background: 'transparent',
        border: 'none',
        borderRadius: 8,
        cursor: 'pointer',
        color: destructive ? 'var(--coral)' : 'var(--text)',
        fontSize: 13,
        fontFamily: 'inherit',
        fontWeight: 500,
        textAlign: 'left',
      }}
    >
      {icon}
      <span>{label}</span>
    </button>
  );
}
