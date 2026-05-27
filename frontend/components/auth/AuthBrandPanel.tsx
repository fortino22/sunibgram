'use client';

import { useRef, useState } from 'react';

const VBW = 384;
const VBH = 512;

// Solid lightning bolt forming an "S" — sharp top/bottom spikes, a thick
// stepped middle, one concave notch on the right (Feather "zap" silhouette,
// elongated). Filled, no stroke, so the points stay crisp.
//   apex → long left edge → foot → bottom apex → long right edge → notch → apex
const BOLT = 'M236 0 L40 300 L190 300 L150 512 L344 224 L190 224 Z';

/**
 * Right-side brand panel for the auth pages. Renders a large lightning bolt "S"
 * as a hollow double outline (two parallel white edges, like the X on x.com):
 * a thick white stroke with a thinner bg stroke carved through its center. A
 * white radial gradient tracking the cursor sweeps light along the edges.
 */
export function AuthBrandPanel() {
  const svgRef = useRef<SVGSVGElement>(null);
  const [pos, setPos] = useState({ x: VBW / 2, y: VBH / 2 });

  const handleMove = (e: React.MouseEvent) => {
    const svg = svgRef.current;
    if (!svg) return;
    const r = svg.getBoundingClientRect();
    setPos({
      x: ((e.clientX - r.left) / r.width) * VBW,
      y: ((e.clientY - r.top) / r.height) * VBH,
    });
  };

  return (
    <div
      onMouseMove={handleMove}
      className="hidden lg:flex flex-1 items-center justify-center relative overflow-hidden"
      style={{ background: 'var(--bg)', paddingLeft: '1%' }}
    >
      <svg
        ref={svgRef}
        viewBox={`0 0 ${VBW} ${VBH}`}
        width="100%"
        height="auto"
        aria-hidden="true"
        style={{ maxWidth: 520, maxHeight: '84vh', transform: 'translateX(-12%)' }}
      >
        <defs>
          <radialGradient
            id="auth-light"
            gradientUnits="userSpaceOnUse"
            cx={pos.x}
            cy={pos.y}
            r={360}
          >
            <stop offset="0%" stopColor="#ffffff" />
            <stop offset="40%" stopColor="var(--border-strong)" />
            <stop offset="100%" stopColor="var(--border)" />
          </radialGradient>
        </defs>

        {/* Tilt clockwise so the zigzag reads as an "S" */}
        <g transform={`rotate(25 ${VBW / 2} ${VBH / 2})`}>
          {/* Thick white outline of the bolt — lit by the cursor gradient */}
          <path
            d={BOLT}
            fill="none"
            stroke="url(#auth-light)"
            strokeWidth={26}
            strokeLinejoin="round"
          />
          {/* Carve the center out, leaving two parallel white edges */}
          <path
            d={BOLT}
            fill="none"
            stroke="var(--bg)"
            strokeWidth={12}
            strokeLinejoin="round"
          />
        </g>
      </svg>
    </div>
  );
}
