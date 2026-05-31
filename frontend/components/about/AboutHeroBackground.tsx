'use client';

import { useEffect, useRef } from 'react';

export function AboutHeroBackground() {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let mouseX = window.innerWidth / 2;
    let mouseY = window.innerHeight / 2;
    let currentX = mouseX;
    let currentY = mouseY;

    const onPointerMove = (e: PointerEvent) => {
      const rect = container.getBoundingClientRect();
      mouseX = e.clientX - rect.left;
      mouseY = e.clientY - rect.top;
    };

    window.addEventListener('pointermove', onPointerMove);

    let frame: number;
    const animate = (time: number) => {
      // Ease the cursor blob position
      currentX += (mouseX - currentX) * 0.04;
      currentY += (mouseY - currentY) * 0.04;

      container.style.setProperty('--cursor-x', `${currentX}px`);
      container.style.setProperty('--cursor-y', `${currentY}px`);

      // Time-based animations for the orbiting blobs
      const t = time * 0.0004;
      container.style.setProperty('--blob1-x', `${50 + Math.sin(t) * 30}%`);
      container.style.setProperty('--blob1-y', `${50 + Math.cos(t) * 30}%`);

      container.style.setProperty('--blob2-x', `${50 + Math.cos(t * 0.8 + 2) * 35}%`);
      container.style.setProperty('--blob2-y', `${50 + Math.sin(t * 0.9 + 2) * 35}%`);

      container.style.setProperty('--blob3-x', `${50 + Math.sin(t * 0.7 + 4) * 20}%`);
      container.style.setProperty('--blob3-y', `${50 + Math.cos(t * 0.6 + 4) * 40}%`);

      frame = requestAnimationFrame(animate);
    };

    frame = requestAnimationFrame(animate);

    return () => {
      window.removeEventListener('pointermove', onPointerMove);
      cancelAnimationFrame(frame);
    };
  }, []);

  return (
    <div
      ref={containerRef}
      style={{
        position: 'absolute',
        inset: 0,
        overflow: 'hidden',
        pointerEvents: 'none',
        zIndex: 0,
        background: 'var(--bg)',
      }}
    >
      <div
        style={{
          position: 'absolute',
          inset: '-20%',
          width: '140%',
          height: '140%',
          background: `
            radial-gradient(circle 600px at var(--cursor-x, 50%) var(--cursor-y, 50%), rgba(0, 242, 254, 0.85), transparent 70%),
            radial-gradient(circle 800px at var(--blob1-x, 25%) var(--blob1-y, 40%), rgba(79, 172, 254, 0.9), transparent 70%),
            radial-gradient(circle 900px at var(--blob2-x, 75%) var(--blob2-y, 60%), rgba(0, 204, 255, 0.85), transparent 70%),
            radial-gradient(circle 700px at var(--blob3-x, 50%) var(--blob3-y, 20%), rgba(102, 126, 234, 0.8), transparent 70%)
          `,
          transition: 'background 0s',
          filter: 'blur(40px)',
          transform: 'translateZ(0)',
          opacity: 1,
        }}
      />
      {/* Grid overlay */}
      <div className="about-hero-grid" style={{ zIndex: 1 }} />
    </div>
  );
}
