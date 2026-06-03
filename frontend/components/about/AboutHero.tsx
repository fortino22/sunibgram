'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { AnimatePresence, motion } from 'framer-motion';
import { AboutHeroBackground } from '@/components/about/AboutHeroBackground';

const BINUS_LOGOS = [
  'binus university.png',
  'binus international.png',
  'binus bandung.png',
  'binus bekasi.png',
  'binus malang.png',
  'binus semarang.png',
  'binus greater jakarta.png',
  'binus online.png',
  'binus bussiness.png',
];

const CYCLE_MS = 2400;
const LOGO_HEIGHT = '1.6em';

export function AboutHero() {
  const [index, setIndex] = useState(0);

  /* Preload every logo so the layout animation has real widths from frame 1. */
  useEffect(() => {
    BINUS_LOGOS.forEach((filename) => {
      const img = new window.Image();
      img.src = `/assets/binus/${encodeURIComponent(filename)}`;
    });
  }, []);

  useEffect(() => {
    const id = setInterval(() => {
      setIndex((i) => (i + 1) % BINUS_LOGOS.length);
    }, CYCLE_MS);
    return () => clearInterval(id);
  }, []);

  const currentSrc = `/assets/binus/${encodeURIComponent(BINUS_LOGOS[index])}`;

  return (
    <section className="relative" id="what">
      <AboutHeroBackground />
      <div className="about-hero-grain" aria-hidden />
      <div
        className="about-container-narrow relative z-10 flex flex-col items-center text-center"
        style={{ paddingTop: '40px', paddingBottom: '40px' }}
      >
        <h1
          style={{
            color: 'var(--text)',
            fontWeight: 500,
            lineHeight: 1.08,
            fontSize: 'clamp(2.5rem, 6vw, 5rem)',
            letterSpacing: '-0.025em',
            maxWidth: '52rem',
          }}
        >
          The{' '}
          <motion.span
            layout
            className="about-inline-chip"
            transition={{ layout: { duration: 0.7, ease: [0.22, 0.8, 0.36, 1] } }}
            style={{ position: 'relative', overflow: 'hidden', padding: '4px 6px' }}
          >
            <AnimatePresence initial={false} mode="popLayout">
              <motion.img
                key={currentSrc}
                src={currentSrc}
                alt="Binus University"
                initial={{ y: '120%', opacity: 0 }}
                animate={{ y: 0, opacity: 1 }}
                exit={{ y: '-120%', opacity: 0 }}
                transition={{ duration: 0.75, ease: [0.22, 0.8, 0.36, 1] }}
                style={{
                  height: LOGO_HEIGHT,
                  width: 'auto',
                  objectFit: 'contain',
                  display: 'block',
                }}
              />
            </AnimatePresence>
          </motion.span>{' '}
          social network.
        </h1>

        <p
          style={{
            color: 'var(--text)',
            fontSize: '1.0625rem',
            lineHeight: 1.6,
            maxWidth: '38rem',
            marginTop: '28px',
          }}
        >
          The campus app for Binus students. Threads, communities, and events —
          one verified identity, no anonymous noise.
        </p>

        <div className="flex items-center justify-center" style={{ marginTop: '36px' }}>
          <Link
            href="/signup"
            className="about-btn-primary about-btn-primary-inverted"
            style={{ color: 'var(--text)', borderColor: 'rgba(0, 0, 0, 0.18)' }}
          >
            Join with your Binus email
          </Link>
        </div>
      </div>
    </section>
  );
}
