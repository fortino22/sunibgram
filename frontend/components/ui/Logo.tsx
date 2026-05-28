'use client';

import { useId } from 'react';
import type { CSSProperties } from 'react';

const VBW = 384;
const VBH = 512;

// Lightning-bolt "S" — identical silhouette to AuthBrandPanel. Rendered as a
// hollow double outline (a thick stroke with a thinner bg stroke carved through
// its center, like the X on x.com): two parallel edges with a transparent gap
// down the middle. A white sheen sweeps across on a loop, automatically.
const BOLT = 'M236 0 L40 300 L190 300 L150 512 L344 224 L190 224 Z';

interface LogoProps {
  size?: number;
  className?: string;
  style?: CSSProperties;
  /** Color carved through the center to form the hole — match the bg behind the logo. */
  bg?: string;
}

export function Logo({ size = 32, className, style, bg = 'var(--bg)' }: LogoProps) {
  const uid = useId();
  const sheenId = `logo-sheen-${uid}`;
  const cx = VBW / 2;
  const cy = VBH / 2;

  return (
    <svg
      width={size}
      height={size}
      viewBox={`0 0 ${VBW} ${VBH}`}
      role="img"
      aria-label="Sunibgram"
      className={className}
      style={style}
    >
      <defs>
        {/* A narrow bright band that sweeps left→right, then rests, on a loop */}
        <linearGradient id={sheenId} x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#fff" stopOpacity="0" />
          <stop offset="42%" stopColor="#fff" stopOpacity="0" />
          <stop offset="50%" stopColor="#fff" stopOpacity="0.95" />
          <stop offset="58%" stopColor="#fff" stopOpacity="0" />
          <stop offset="100%" stopColor="#fff" stopOpacity="0" />
          <animateTransform
            attributeName="gradientTransform"
            type="translate"
            values="-1 0; 1 0; 1 0"
            keyTimes="0; 0.45; 1"
            dur="3.5s"
            repeatCount="indefinite"
          />
        </linearGradient>
      </defs>

      {/* Tilt clockwise so the zigzag reads as an "S" */}
      <g transform={`rotate(25 ${cx} ${cy})`}>
        {/* bolt body */}
        <path d={BOLT} fill="none" stroke="currentColor" strokeWidth={26} strokeLinejoin="round" />
        {/* sweeping glint over the body */}
        <path
          d={BOLT}
          fill="none"
          stroke={`url(#${sheenId})`}
          strokeWidth={26}
          strokeLinejoin="round"
          style={{ mixBlendMode: 'lighten' }}
        />
        {/* carve the center → hollow double edge */}
        <path d={BOLT} fill="none" stroke={bg} strokeWidth={12} strokeLinejoin="round" />
      </g>
    </svg>
  );
}
