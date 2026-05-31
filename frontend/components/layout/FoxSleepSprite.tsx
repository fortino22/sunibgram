'use client';

import { useEffect, useState } from 'react';

// Frames produced by scripts/slice-fox-sprite.mjs (public/assets/fox-sleep).
// White fox = dark theme, black fox = light theme. Loop only the curled
// sleeping-ball frames (skip the awake sit/curl lead-in) so the fox always
// looks asleep. Both ranges are 9 frames → one shared counter stays in sync.
const WHITE_SLEEP = { start: 2, count: 9 }; // white 02..10
const BLACK_SLEEP = { start: 1, count: 9 }; // black 01..09
const FRAME_MS = 420; // slow, restful breathing

const pad = (n: number) => String(n).padStart(2, '0');

const frameStyle: React.CSSProperties = {
  position: 'absolute',
  inset: 0,
  margin: 'auto',
  maxWidth: '100%',
  maxHeight: '100%',
  width: 'auto',
  height: 'auto',
  objectFit: 'contain',
  imageRendering: 'pixelated',
};

/**
 * Sleeping-fox animation for the "coming soon" popup. Cycles pre-cropped frame
 * PNGs (settle → curl → sleep). Renders both colour variants; CSS shows the
 * white fox on dark theme and the black fox on light theme.
 */
export function FoxSleepSprite() {
  const [frame, setFrame] = useState(0);

  useEffect(() => {
    const timer = setInterval(() => setFrame((f) => f + 1), FRAME_MS);
    return () => clearInterval(timer);
  }, []);

  return (
    <span style={{ position: 'relative', display: 'block', width: 132, height: 106 }} aria-hidden="true">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        className="fox-sleep-frame--white"
        src={`/assets/fox-sleep/white/${pad(WHITE_SLEEP.start + (frame % WHITE_SLEEP.count))}.png`}
        alt=""
        style={frameStyle}
      />
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        className="fox-sleep-frame--black"
        src={`/assets/fox-sleep/black/${pad(BLACK_SLEEP.start + (frame % BLACK_SLEEP.count))}.png`}
        alt=""
        style={frameStyle}
      />
    </span>
  );
}
