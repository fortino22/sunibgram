'use client';

import { useCallback, useRef, useState } from 'react';

export function useMeasuredSize<T extends HTMLElement>() {
  const observerRef = useRef<ResizeObserver | null>(null);
  const [size, setSize] = useState({ w: 0, h: 0 });

  const setRef = useCallback((node: T | null) => {
    if (observerRef.current) {
      observerRef.current.disconnect();
      observerRef.current = null;
    }
    if (!node) {
      return;
    }
    const ro = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (!entry) return;
      const { width, height } = entry.contentRect;
      setSize((prev) => {
        const w = Math.floor(width);
        const h = Math.floor(height);
        if (prev.w === w && prev.h === h) return prev;
        return { w, h };
      });
    });
    ro.observe(node);
    observerRef.current = ro;
    // seed sync so first paint already has a size (no chart flash)
    const rect = node.getBoundingClientRect();
    setSize((prev) => {
      const w = Math.floor(rect.width);
      const h = Math.floor(rect.height);
      if (prev.w === w && prev.h === h) return prev;
      return { w, h };
    });
  }, []);

  return [setRef, size] as const;
}
