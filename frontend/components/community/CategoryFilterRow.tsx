'use client';

import { useState, useRef, useEffect } from 'react';
import { ArrowLeft, ArrowRight } from 'lucide-react';
import type { Category } from '@/lib/services/communities';

interface CategoryFilterRowProps {
  categories: Category[];
  selectedId: string | undefined;
  onSelect: (id: string | undefined) => void;
  showBorderTop?: boolean;
}

export function CategoryFilterRow({
  categories,
  selectedId,
  onSelect,
  showBorderTop = true,
}: CategoryFilterRowProps) {
  const sliderRef = useRef<HTMLDivElement>(null);
  const drag = useRef({ active: false, startX: 0, scrollLeft: 0, moved: false });

  const [canLeft, setCanLeft] = useState(false);
  const [canRight, setCanRight] = useState(false);
  const [hover, setHover] = useState(false);

  const onMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
    const el = sliderRef.current;
    if (!el) return;
    drag.current = { active: true, startX: e.pageX - el.offsetLeft, scrollLeft: el.scrollLeft, moved: false };
    el.style.cursor = 'grabbing';
    el.style.userSelect = 'none';
  };

  const onMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!drag.current.active) return;
    const el = sliderRef.current;
    if (!el) return;
    const walk = e.pageX - el.offsetLeft - drag.current.startX;
    if (Math.abs(walk) > 4) drag.current.moved = true;
    el.scrollLeft = drag.current.scrollLeft - walk;
  };

  const onMouseUp = () => {
    drag.current.active = false;
    if (sliderRef.current) {
      sliderRef.current.style.cursor = 'grab';
      sliderRef.current.style.userSelect = '';
    }
  };

  const scrollStep = (dir: -1 | 1) => {
    const el = sliderRef.current;
    if (!el) return;
    const step = Math.max(160, el.clientWidth * 0.8);
    el.scrollBy({ left: dir * step, behavior: 'smooth' });
  };

  useEffect(() => {
    const el = sliderRef.current;
    if (!el) return;
    const update = () => {
      setCanLeft(el.scrollLeft > 0);
      setCanRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 1);
    };
    update();
    el.addEventListener('scroll', update, { passive: true });
    const ro = new ResizeObserver(update);
    ro.observe(el);
    return () => {
      el.removeEventListener('scroll', update);
      ro.disconnect();
    };
  }, [categories.length]);

  return (
    <div
      onMouseEnter={() => setHover(true)}
      onMouseLeave={() => setHover(false)}
      style={{
        position: 'relative',
        borderTop: showBorderTop ? '1px solid var(--border)' : undefined,
      }}
    >
      <div
        ref={sliderRef}
        onMouseDown={onMouseDown}
        onMouseMove={onMouseMove}
        onMouseUp={onMouseUp}
        onMouseLeave={onMouseUp}
        className="hide-scrollbar"
        style={{
          display: 'flex',
          gap: 8,
          padding: '12px 16px',
          overflowX: 'auto',
          cursor: 'grab',
        }}
      >
        <button
          onClick={() => { if (drag.current.moved) return; onSelect(undefined); }}
          className="hover-dim"
          style={{
            padding: '5px 14px',
            borderRadius: 999,
            border: `1px solid ${!selectedId ? 'var(--text)' : 'var(--border)'}`,
            background: !selectedId ? 'var(--text)' : 'transparent',
            color: !selectedId ? 'var(--bg)' : 'var(--text)',
            fontSize: 13,
            fontWeight: 600,
            cursor: 'pointer',
            whiteSpace: 'nowrap',
            flexShrink: 0,
          }}
        >
          All
        </button>
        {categories.map((c) => (
          <button
            key={c.id}
            onClick={() => { if (drag.current.moved) return; onSelect(c.id); }}
            className="hover-dim"
            style={{
              padding: '5px 14px',
              borderRadius: 999,
              border: `1px solid ${selectedId === c.id ? 'var(--text)' : 'var(--border)'}`,
              background: selectedId === c.id ? 'var(--text)' : 'transparent',
              color: selectedId === c.id ? 'var(--bg)' : 'var(--text)',
              fontSize: 13,
              fontWeight: 600,
              cursor: 'pointer',
              whiteSpace: 'nowrap',
              flexShrink: 0,
            }}
          >
            {c.name}
          </button>
        ))}
      </div>

      <button
        type="button"
        onClick={() => scrollStep(-1)}
        aria-label="Scroll categories left"
        className="hover-dim"
        style={{
          position: 'absolute',
          left: 8,
          top: '50%',
          transform: 'translateY(-50%)',
          zIndex: 2,
          width: 32,
          height: 32,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderRadius: 999,
          border: '1px solid var(--border)',
          background: 'var(--bg)',
          color: 'var(--text)',
          cursor: 'pointer',
          boxShadow: '6px 0 12px -4px var(--bg)',
          opacity: hover && canLeft ? 1 : 0,
          pointerEvents: hover && canLeft ? 'auto' : 'none',
          transition: 'opacity 0.15s ease',
        }}
      >
        <ArrowLeft size={18} />
      </button>

      <button
        type="button"
        onClick={() => scrollStep(1)}
        aria-label="Scroll categories right"
        className="hover-dim"
        style={{
          position: 'absolute',
          right: 8,
          top: '50%',
          transform: 'translateY(-50%)',
          zIndex: 2,
          width: 32,
          height: 32,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderRadius: 999,
          border: '1px solid var(--border)',
          background: 'var(--bg)',
          color: 'var(--text)',
          cursor: 'pointer',
          boxShadow: '-6px 0 12px -4px var(--bg)',
          opacity: hover && canRight ? 1 : 0,
          pointerEvents: hover && canRight ? 'auto' : 'none',
          transition: 'opacity 0.15s ease',
        }}
      >
        <ArrowRight size={18} />
      </button>
    </div>
  );
}
