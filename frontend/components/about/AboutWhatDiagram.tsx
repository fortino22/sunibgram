'use client';

import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { Logo } from '@/components/ui/Logo';

/* ------------------------------------------------------------------ *
 *  Editorial echo of one question that the campus keeps asking and    *
 *  re-asking, in different words, across different channels, because  *
 *  no canonical place exists. Below a thin rule, a single Sunibgram   *
 *  surface gathers all of them into one answered post.                *
 * ------------------------------------------------------------------ */

type Echo = {
  id: string;
  text: string;
  x: number;          // left, % of canvas
  y: number;          // top, % of canvas
  size: number;       // px
  weight: 300 | 400 | 500 | 600;
  italic?: boolean;
  rotation: number;   // deg
  baseOpacity: number;
  ts: string;
};

const ECHOES: Echo[] = [
  { id: 'e1', text: 'Did the lab notes get posted?', x: 5,  y: 6,  size: 22, weight: 600, rotation: -0.6, baseOpacity: 0.92, ts: '3 days ago' },
  { id: 'e2', text: 'ada yg punya lab notes ga?',    x: 52, y: 14, size: 14, weight: 400, italic: true,  rotation:  1.1, baseOpacity: 0.72, ts: '2 days ago' },
  { id: 'e3', text: 'anyone got the lab notes?',     x: 12, y: 22, size: 17, weight: 300, rotation: -1.0, baseOpacity: 0.82, ts: 'yesterday' },
  { id: 'e4', text: 'laprak udah keluar blm?',       x: 55, y: 29, size: 15, weight: 400, italic: true,  rotation:  0.8, baseOpacity: 0.65, ts: '22h ago' },
  { id: 'e5', text: 'lab notes uploaded?',           x: 8,  y: 37, size: 18, weight: 400, rotation: -0.6, baseOpacity: 0.58, ts: '14h ago' },
  { id: 'e6', text: 'yo, lab notes?',                x: 60, y: 44, size: 13, weight: 300, italic: true,  rotation:  1.4, baseOpacity: 0.52, ts: '6h ago' },
  { id: 'e7', text: 'ada link lab notes?',           x: 14, y: 51, size: 16, weight: 400, rotation: -0.5, baseOpacity: 0.46, ts: '47 min' },
  { id: 'e8', text: 'tau di mana lab notes?',        x: 50, y: 58, size: 14, weight: 300, italic: true,  rotation:  0.7, baseOpacity: 0.42, ts: '12 min' },
];

const N = ECHOES.length;

export function AboutWhatDiagram() {
  /* step
   *   0           empty
   *   1..N        echo i appears
   *   N + 1       horizontal rule draws
   *   N + 2       resolution appears
   *   N + 3       echoes dim, resolution settles
   */
  const [step, setStep] = useState(0);

  useEffect(() => {
    let mounted = true;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function cycle() {
      if (!mounted) return;
      setStep(0);

      ECHOES.forEach((_, i) => {
        timeouts.push(setTimeout(() => {
          if (mounted) setStep(i + 1);
        }, 350 * i + 250));
      });

      const lastEcho = 350 * (N - 1) + 250;
      timeouts.push(setTimeout(() => { if (mounted) setStep(N + 1); }, lastEcho + 800));
      timeouts.push(setTimeout(() => { if (mounted) setStep(N + 2); }, lastEcho + 1200));
      timeouts.push(setTimeout(() => { if (mounted) setStep(N + 3); }, lastEcho + 1500));
      timeouts.push(setTimeout(cycle, 9000));
    }

    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  const ruleVisible    = step >= N + 1;
  const answerVisible  = step >= N + 2;
  const echoesDim      = step >= N + 3;

  return (
    <div
      style={{
        position: 'relative',
        width: '100%',
        aspectRatio: '1 / 1',
        maxWidth: 460,
        marginInline: 'auto',
        borderRadius: 20,
        border: '1px solid var(--border)',
        background: 'var(--bg-card)',
        overflow: 'hidden',
      }}
    >
      {/* Echoes — scattered across upper canvas */}
      <div style={{ position: 'absolute', inset: 0, padding: '14px 18px' }}>
        {ECHOES.map((e, i) => {
          const shown = step > i;
          const targetOpacity = shown
            ? echoesDim
              ? e.baseOpacity * 0.18
              : e.baseOpacity
            : 0;
          return (
            <motion.div
              key={e.id}
              initial={{ opacity: 0, y: 4 }}
              animate={{ opacity: targetOpacity, y: shown ? 0 : 4 }}
              transition={{ duration: 0.55, ease: [0.22, 0.8, 0.36, 1] }}
              style={{
                position: 'absolute',
                left: `${e.x}%`,
                top: `${e.y}%`,
                color: 'var(--text)',
                fontSize: e.size,
                fontWeight: e.weight,
                fontStyle: e.italic ? 'italic' : 'normal',
                letterSpacing: '-0.012em',
                lineHeight: 1.2,
                transform: `rotate(${e.rotation}deg)`,
                transformOrigin: 'left center',
                whiteSpace: 'nowrap',
                userSelect: 'none',
                display: 'inline-flex',
                alignItems: 'baseline',
                gap: 8,
              }}
            >
              <span>{e.text}</span>
              <span
                style={{
                  fontSize: Math.max(10, Math.round(e.size * 0.55)),
                  fontWeight: 400,
                  fontStyle: 'normal',
                  color: 'var(--text-tertiary)',
                  letterSpacing: 0,
                }}
              >
                {e.ts}
              </span>
            </motion.div>
          );
        })}
      </div>

      {/* Thin rule across the middle — draws left to right */}
      <motion.div
        initial={{ scaleX: 0 }}
        animate={{ scaleX: ruleVisible ? 1 : 0 }}
        transition={{ duration: 0.75, ease: [0.22, 0.8, 0.36, 1] }}
        style={{
          position: 'absolute',
          left: '18px',
          right: '18px',
          top: '68%',
          height: 1,
          background: 'var(--text)',
          transformOrigin: 'left center',
          opacity: 0.85,
        }}
      />

      {/* Resolution — one canonical place */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: answerVisible ? 1 : 0, y: answerVisible ? 0 : 10 }}
        transition={{ duration: 0.5, ease: [0.22, 0.8, 0.36, 1] }}
        style={{
          position: 'absolute',
          left: 22,
          right: 22,
          top: '73%',
          display: 'flex',
          flexDirection: 'column',
          gap: 6,
        }}
      >
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 7,
            fontSize: 11,
            fontWeight: 600,
            color: 'var(--text-tertiary)',
            textTransform: 'uppercase',
            letterSpacing: '0.1em',
          }}
        >
          <span style={{ color: 'var(--text)', display: 'inline-flex', alignItems: 'center' }}>
            <Logo size={13} bg="var(--bg-card)" />
          </span>
          <span style={{ color: 'var(--text)' }}>Sunibgram</span>
        </div>

        <div
          style={{
            fontSize: 17,
            fontWeight: 500,
            color: 'var(--text)',
            letterSpacing: '-0.012em',
            lineHeight: 1.3,
          }}
        >
          Did the lab notes get posted?
        </div>

        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 8,
            fontSize: 12,
            fontWeight: 500,
            color: 'var(--text-tertiary)',
            letterSpacing: '-0.005em',
          }}
        >
          <span style={{ color: 'var(--accent)', fontWeight: 600 }}>12 verified answers</span>
          <span aria-hidden style={{ opacity: 0.5 }}>·</span>
          <span>pinned · 2h ago</span>
        </div>
      </motion.div>
    </div>
  );
}
