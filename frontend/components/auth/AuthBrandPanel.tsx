'use client';

import { useEffect, useState } from 'react';

const VBW = 384;
const VBH = 512;

// Solid lightning bolt forming an "S" — sharp top/bottom spikes, a thick
// stepped middle, one concave notch on the right (Feather "zap" silhouette,
// elongated). Stroked (not filled) to draw a hollow double outline.
//   apex → long left edge → foot → bottom apex → long right edge → notch → apex
const BOLT = 'M236 0 L40 300 L190 300 L150 512 L344 224 L190 224 Z';

/**
 * Right-side brand panel for the auth pages. Renders a large lightning bolt "S"
 * as a hollow double outline (two parallel white edges, like the X on x.com):
 * a thick white stroke with a thinner bg stroke carved through its center. Two
 * mirrored radial gradients track the cursor so the left and right edges light
 * symmetrically as the pointer moves anywhere on the page.
 */
export function AuthBrandPanel() {
  const [pos, setPos] = useState({ x: VBW / 2, y: VBH / 2 });

  // Track the cursor across the whole page (like x.com). Map the viewport
  // directly into the bolt's viewBox: pointer X → horizontal, Y → vertical.
  // No element measuring, so transforms/scaling can't throw it off and the
  // light center always lands on the bolt.
  useEffect(() => {
    const handleMove = (e: MouseEvent) => {
      setPos({
        x: (e.clientX / window.innerWidth) * VBW,
        y: (e.clientY / window.innerHeight) * VBH,
      });
    };
    window.addEventListener('mousemove', handleMove);
    return () => window.removeEventListener('mousemove', handleMove);
  }, []);

  return (
    <div
      className="hidden lg:flex flex-1 items-center justify-center relative overflow-hidden"
      style={{ background: 'var(--bg)', paddingLeft: '1%' }}
    >
      <svg
        viewBox={`0 0 ${VBW} ${VBH}`}
        width="100%"
        aria-hidden="true"
        style={{ height: 'auto', maxWidth: 520, maxHeight: '84vh', transform: 'translateX(-12%)' }}
      >
        <defs>
          {/* Light at the cursor */}
          <radialGradient
            id="auth-light"
            gradientUnits="userSpaceOnUse"
            cx={pos.x}
            cy={pos.y}
            r={560}
          >
            <stop offset="0%" stopColor="#ffffff" />
            <stop offset="10%" stopColor="var(--auth-bolt-mid)" />
            <stop offset="80%" stopColor="var(--auth-bolt-mid)" />
            <stop offset="100%" stopColor="var(--auth-bolt-edge)" />
          </radialGradient>
          {/* Mirror of that light across the vertical center axis */}
          <radialGradient
            id="auth-light-mirror"
            gradientUnits="userSpaceOnUse"
            cx={VBW - pos.x}
            cy={pos.y}
            r={560}
          >
            <stop offset="0%" stopColor="#ffffff" />
            <stop offset="10%" stopColor="var(--auth-bolt-mid)" />
            <stop offset="80%" stopColor="var(--auth-bolt-mid)" />
            <stop offset="100%" stopColor="var(--auth-bolt-edge)" />
          </radialGradient>
        </defs>

        {/* Tilt clockwise so the zigzag reads as an "S" */}
        <g transform={`rotate(25 ${VBW / 2} ${VBH / 2})`}>
          {/* White outline lit by the cursor */}
          <path
            d={BOLT}
            fill="none"
            stroke="url(#auth-light)"
            strokeWidth={26}
            strokeLinejoin="round"
          />
          {/* Same outline lit by the mirrored light — brighter of the two wins */}
          <path
            d={BOLT}
            fill="none"
            stroke="url(#auth-light-mirror)"
            strokeWidth={26}
            strokeLinejoin="round"
            style={{ mixBlendMode: 'lighten' }}
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
