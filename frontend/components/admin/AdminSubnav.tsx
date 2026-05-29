'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';
import {
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Flag,
  GraduationCap,
  LayoutDashboard,
  MessageSquareText,
  Receipt,
  ShieldAlert,
  Tags,
  TrendingUp,
  Users,
  type LucideIcon,
} from 'lucide-react';

interface NavLink {
  href: string;
  label: string;
  icon: LucideIcon;
}

const GROUPS: NavLink[][] = [
  [{ href: '/admin', label: 'Overview', icon: LayoutDashboard }],
  [
    { href: '/admin/users', label: 'Users', icon: Users },
    { href: '/admin/feedback', label: 'Feedback', icon: MessageSquareText },
  ],
  [
    { href: '/admin/reports', label: 'Reports', icon: Flag },
    { href: '/admin/trending-reports', label: 'Trending', icon: TrendingUp },
  ],
  [
    { href: '/admin/events', label: 'Events', icon: CalendarDays },
    { href: '/admin/faculties', label: 'Faculties', icon: GraduationCap },
    { href: '/admin/categories', label: 'Categories', icon: Tags },
    { href: '/admin/violations', label: 'Report Types', icon: ShieldAlert },
  ],
  [{ href: '/admin/orders', label: 'Orders', icon: Receipt }],
];

const ALL_LINKS: NavLink[] = GROUPS.flat();

const useIsoLayoutEffect = typeof window === 'undefined' ? useEffect : useLayoutEffect;

export function AdminSubnav() {
  const pathname = usePathname();
  const wrapperRef = useRef<HTMLDivElement | null>(null);
  const navRef = useRef<HTMLDivElement | null>(null);
  const linkRefs = useRef<Map<string, HTMLAnchorElement>>(new Map());
  const [pill, setPill] = useState<{ left: number; width: number; visible: boolean }>(
    { left: 0, width: 0, visible: false }
  );

  const isActive = (href: string): boolean => {
    if (href === '/admin') return pathname === '/admin';
    return pathname.startsWith(href);
  };

  const activeHref = ALL_LINKS.find((l) => isActive(l.href))?.href;

  const measurePill = useCallback(() => {
    if (!activeHref || !navRef.current) {
      setPill((p) => ({ ...p, visible: false }));
      return;
    }
    const el = linkRefs.current.get(activeHref);
    if (!el) return;
    const navRect = navRef.current.getBoundingClientRect();
    const elRect = el.getBoundingClientRect();
    setPill({
      left: elRect.left - navRect.left + navRef.current.scrollLeft,
      width: elRect.width,
      visible: true,
    });
  }, [activeHref]);

  const updateOverflow = useCallback(() => {
    const nav = navRef.current;
    const wrapper = wrapperRef.current;
    if (!nav || !wrapper) return;
    const { scrollLeft, scrollWidth, clientWidth } = nav;
    const hasLeft = scrollLeft > 1;
    const hasRight = scrollLeft + clientWidth < scrollWidth - 1;
    wrapper.dataset.overflowLeft = hasLeft ? 'true' : 'false';
    wrapper.dataset.overflowRight = hasRight ? 'true' : 'false';
  }, []);

  useIsoLayoutEffect(() => {
    measurePill();
    updateOverflow();
  }, [measurePill, updateOverflow, pathname]);

  useEffect(() => {
    const handle = () => {
      measurePill();
      updateOverflow();
    };
    window.addEventListener('resize', handle);
    const nav = navRef.current;
    nav?.addEventListener('scroll', updateOverflow, { passive: true });
    const ro = typeof ResizeObserver !== 'undefined' && nav ? new ResizeObserver(handle) : null;
    if (ro && nav) ro.observe(nav);
    return () => {
      window.removeEventListener('resize', handle);
      nav?.removeEventListener('scroll', updateOverflow);
      ro?.disconnect();
    };
  }, [measurePill, updateOverflow]);

  return (
    <div
      ref={wrapperRef}
      data-overflow-left="false"
      data-overflow-right="false"
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 10,
        background: 'var(--bg)',
        borderBottom: '1px solid var(--border)',
      }}
    >
      <button
        type="button"
        aria-label="Scroll tabs left"
        onClick={() => navRef.current?.scrollBy({ left: -240, behavior: 'smooth' })}
        className="admin-subnav-arrow admin-subnav-arrow-left"
      >
        <ChevronLeft size={16} strokeWidth={2.2} />
      </button>
      <button
        type="button"
        aria-label="Scroll tabs right"
        onClick={() => navRef.current?.scrollBy({ left: 240, behavior: 'smooth' })}
        className="admin-subnav-arrow admin-subnav-arrow-right"
      >
        <ChevronRight size={16} strokeWidth={2.2} />
      </button>
      <nav
        ref={navRef}
        className="hide-scrollbar"
        style={{
          position: 'relative',
          display: 'flex',
          alignItems: 'center',
          gap: 2,
          padding: '10px 24px',
          overflowX: 'auto',
        }}
      >
        <div
          className="admin-subnav-pill"
          style={{
            left: pill.left,
            width: pill.width,
            opacity: pill.visible ? 1 : 0,
          }}
          aria-hidden
        />
        {GROUPS.map((group, gi) => (
          <div
            key={gi}
            style={{ display: 'flex', alignItems: 'center', gap: 2 }}
          >
            {gi > 0 && <span className="admin-subnav-divider" aria-hidden />}
            {group.map((l) => {
              const active = isActive(l.href);
              const Icon = l.icon;
              return (
                <Link
                  key={l.href}
                  ref={(node) => {
                    if (node) linkRefs.current.set(l.href, node);
                    else linkRefs.current.delete(l.href);
                  }}
                  href={l.href}
                  data-active={active}
                  className="admin-subnav-link"
                >
                  <Icon size={16} strokeWidth={2} aria-hidden />
                  <span>{l.label}</span>
                </Link>
              );
            })}
          </div>
        ))}
      </nav>
    </div>
  );
}
