'use client';

import { useState, useRef, useCallback, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Logo } from '@/components/ui/Logo';

/* ------------------------------------------------------------------ */
/*  Data                                                               */
/* ------------------------------------------------------------------ */

type DropdownItem = {
  title: string;
  description: string;
  href: string;
};

type DropdownColumn = {
  heading: string;
  items: DropdownItem[];
};

type NavDropdown = {
  label: string;
  columns: DropdownColumn[];
};

type NavLink = {
  label: string;
  href: string;
};

const DROPDOWNS: NavDropdown[] = [
  {
    label: 'Features',
    columns: [
      {
        heading: 'Social',
        items: [
          {
            title: 'Threads',
            description: 'Share thoughts and updates with the campus community.',
            href: '/about/what',
          },
          {
            title: 'Communities',
            description: 'Join groups around shared interests and courses.',
            href: '/about/what',
          },
        ],
      },
      {
        heading: 'Tools',
        items: [
          {
            title: 'Events',
            description: 'Discover and organize campus events and meetups.',
            href: '/about/what',
          },
          {
            title: 'Messages',
            description: 'Real-time private messaging with fellow students.',
            href: '/about/what',
          },
        ],
      },
    ],
  },
  {
    label: 'About',
    columns: [
      {
        heading: 'Learn more',
        items: [
          {
            title: 'What it is',
            description: 'Learn about Sunibgram and our mission for Binus students.',
            href: '/about/what',
          },
          {
            title: 'The team',
            description: 'Meet the people building Sunibgram.',
            href: '/about/team',
          },
        ],
      },
      {
        heading: 'Community',
        items: [
          {
            title: 'Use cases',
            description: 'How students use Sunibgram every day.',
            href: '/about/use-cases',
          },
        ],
      },
    ],
  },
];

const PLAIN_LINKS: NavLink[] = [
  { label: 'Use cases', href: '/about/use-cases' },
  { label: 'Leaderboard', href: '/leaderboard' },
  { label: 'Sign in', href: '/login' },
];

/* ------------------------------------------------------------------ */
/*  Component                                                          */
/* ------------------------------------------------------------------ */

export function AboutNav() {
  const [openDropdown, setOpenDropdown] = useState<number | null>(null);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [mobileExpanded, setMobileExpanded] = useState<number | null>(null);
  const pathname = usePathname();
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  /* Dropdown hover handlers (desktop) */
  const handleMouseEnter = useCallback((index: number) => {
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    setOpenDropdown(index);
  }, []);

  const handleMouseLeave = useCallback(() => {
    timeoutRef.current = setTimeout(() => setOpenDropdown(null), 180);
  }, []);

  /* Close mobile menu on navigation */
  useEffect(() => {
    setMobileOpen(false);
    setMobileExpanded(null);
  }, [pathname]);

  /* Close on Escape */
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setOpenDropdown(null);
        setMobileOpen(false);
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, []);

  /* Lock body scroll when mobile menu is open */
  useEffect(() => {
    document.body.style.overflow = mobileOpen ? 'hidden' : '';
    return () => { document.body.style.overflow = ''; };
  }, [mobileOpen]);

  return (
    <nav className="about-nav">
      <div className="about-container about-nav-inner">
        {/* ---- Logo ---- */}
        <div className="about-nav-left">
          <Link href="/about" className="about-nav-logo">
            <Logo size={35} />
            <span className="about-nav-logo-text">Sunibgram</span>
          </Link>
        </div>

        {/* ---- Desktop links ---- */}
        <div className="about-nav-center">
          {DROPDOWNS.map((dropdown, i) => (
            <div
              key={dropdown.label}
              className="about-nav-trigger-wrap"
              onMouseEnter={() => handleMouseEnter(i)}
              onMouseLeave={handleMouseLeave}
            >
              <button
                className="about-nav-link about-nav-trigger"
                aria-expanded={openDropdown === i}
                onClick={() => setOpenDropdown(openDropdown === i ? null : i)}
              >
                {dropdown.label}
                <svg
                  width="10"
                  height="10"
                  viewBox="0 0 10 10"
                  fill="none"
                  className={`about-nav-chevron${openDropdown === i ? ' about-nav-chevron-open' : ''}`}
                >
                  <path
                    d="M2.5 3.75L5 6.25L7.5 3.75"
                    stroke="currentColor"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              </button>

              {/* Dropdown panel */}
              <div
                className={`about-nav-dropdown${openDropdown === i ? ' about-nav-dropdown-visible' : ''}`}
              >
                {dropdown.columns.map((col) => (
                  <div key={col.heading} className="about-nav-dropdown-col">
                    <span className="about-nav-dropdown-heading">
                      {col.heading}
                    </span>
                    {col.items.map((item) => (
                      <Link
                        key={item.title}
                        href={item.href}
                        className="about-nav-dropdown-item"
                        onClick={() => setOpenDropdown(null)}
                      >
                        <span className="about-nav-dropdown-title">
                          {item.title}
                        </span>
                        <span className="about-nav-dropdown-desc">
                          {item.description}
                        </span>
                      </Link>
                    ))}
                  </div>
                ))}
              </div>
            </div>
          ))}

          {PLAIN_LINKS.map((link) => (
            <Link key={link.label} href={link.href} className="about-nav-link">
              {link.label}
            </Link>
          ))}
        </div>

        {/* ---- CTA + Hamburger ---- */}
        <div className="about-nav-right">
          <Link href="/signup" className="about-btn-dark">
            Join Sunibgram
          </Link>
          <button
            className="about-nav-hamburger"
            onClick={() => setMobileOpen(!mobileOpen)}
            aria-label="Toggle menu"
            aria-expanded={mobileOpen}
          >
            <span className={`about-hamburger-bar about-hamburger-bar-1${mobileOpen ? ' about-hamburger-open' : ''}`} />
            <span className={`about-hamburger-bar about-hamburger-bar-2${mobileOpen ? ' about-hamburger-open' : ''}`} />
            <span className={`about-hamburger-bar about-hamburger-bar-3${mobileOpen ? ' about-hamburger-open' : ''}`} />
          </button>
        </div>
      </div>

      {/* ---- Mobile menu ---- */}
      <div className={`about-nav-mobile${mobileOpen ? ' about-nav-mobile-open' : ''}`}>
        <div className="about-nav-mobile-scroll">
          {DROPDOWNS.map((dropdown, i) => (
            <div key={dropdown.label} className="about-nav-mobile-group">
              <button
                className="about-nav-mobile-trigger"
                onClick={() => setMobileExpanded(mobileExpanded === i ? null : i)}
                aria-expanded={mobileExpanded === i}
              >
                {dropdown.label}
                <svg
                  width="12"
                  height="12"
                  viewBox="0 0 10 10"
                  fill="none"
                  className={`about-nav-chevron${mobileExpanded === i ? ' about-nav-chevron-open' : ''}`}
                >
                  <path
                    d="M2.5 3.75L5 6.25L7.5 3.75"
                    stroke="currentColor"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              </button>
              <div
                className={`about-nav-mobile-items${mobileExpanded === i ? ' about-nav-mobile-items-open' : ''}`}
              >
                {dropdown.columns.flatMap((col) => col.items).map((item) => (
                  <Link
                    key={item.title}
                    href={item.href}
                    className="about-nav-mobile-item"
                  >
                    <span className="about-nav-dropdown-title">{item.title}</span>
                    <span className="about-nav-dropdown-desc">{item.description}</span>
                  </Link>
                ))}
              </div>
            </div>
          ))}

          {PLAIN_LINKS.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className="about-nav-mobile-link"
            >
              {link.label}
            </Link>
          ))}

          <Link href="/signup" className="about-btn-dark about-nav-mobile-cta">
            Join Sunibgram
          </Link>
        </div>
      </div>
    </nav>
  );
}
