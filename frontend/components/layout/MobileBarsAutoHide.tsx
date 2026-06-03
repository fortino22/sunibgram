'use client';

import { useEffect } from 'react';

/**
 * Toggles `mobile-bars-hidden` on <body> based on scroll direction.
 * Hides on scroll-down past a small threshold, reveals on scroll-up.
 * Idempotent: no-op on screens >480px since the bars don't render there.
 */
export function MobileBarsAutoHide() {
  useEffect(() => {
    let lastY = window.scrollY;
    let ticking = false;
    const REVEAL_DELTA = 6;   // px upward movement before reveal
    const HIDE_DELTA = 10;    // px downward movement before hide
    const HIDE_AFTER = 80;    // never hide while near the top

    const onScroll = () => {
      if (ticking) return;
      ticking = true;
      requestAnimationFrame(() => {
        const y = window.scrollY;
        const dy = y - lastY;
        const body = document.body;
        if (y < HIDE_AFTER) {
          body.classList.remove('mobile-bars-hidden');
        } else if (dy > HIDE_DELTA) {
          body.classList.add('mobile-bars-hidden');
        } else if (dy < -REVEAL_DELTA) {
          body.classList.remove('mobile-bars-hidden');
        }
        lastY = y;
        ticking = false;
      });
    };

    window.addEventListener('scroll', onScroll, { passive: true });
    return () => {
      window.removeEventListener('scroll', onScroll);
      document.body.classList.remove('mobile-bars-hidden');
    };
  }, []);

  return null;
}
