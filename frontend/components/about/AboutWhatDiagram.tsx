'use client';

import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { Logo } from '@/components/ui/Logo';

/* ------------------------------------------------------------------ *
 *  One question, asked in five different ways across five channels    *
 *  because none of them are the canonical place for it. Then a rule   *
 *  with the Sunibgram mark breaks the page, and below it the same     *
 *  question reappears — same words, same weight — but answered.       *
 * ------------------------------------------------------------------ */

type Echo = {
  id: string;
  text: string;
  ts: string;
  x: number;            // left, % of canvas
  y: number;            // top, % of canvas
  size: number;         // px
  weight: 300 | 400 | 500 | 600;
  italic?: boolean;
  rotation: number;
  baseOpacity: number;
};

const HERO = 'Did the lab notes get posted?';

const ECHOES: Echo[] = [
  { id: 'e1', text: HERO,                          ts: '3 days ago', x: 4,  y: 5,  size: 22, weight: 600, rotation: -0.3, baseOpacity: 0.95, italic: false },
  { id: 'e2', text: 'ada yg punya lab notes ga?',  ts: '2 days ago', x: 32, y: 16, size: 14, weight: 400, rotation:  0.6, baseOpacity: 0.72, italic: true  },
  { id: 'e3', text: 'anyone got the lab notes?',   ts: 'yesterday',  x: 10, y: 24, size: 17, weight: 300, rotation: -0.5, baseOpacity: 0.6,  italic: false },
  { id: 'e4', text: 'laprak udah keluar blm?',     ts: '22h ago',    x: 40, y: 32, size: 13, weight: 400, rotation:  0.4, baseOpacity: 0.48, italic: true  },
  { id: 'e5', text: 'lab notes uploaded?',         ts: '47 min ago', x: 6,  y: 40, size: 16, weight: 300, rotation: -0.3, baseOpacity: 0.4,  italic: false },
];

const N = ECHOES.length;

export function AboutWhatDiagram() {
  /* step
   *   0           empty
   *   1..N        echo i appears
   *   N + 1       brand mark settles, rule halves sweep out
   *   N + 2       resolution appears, echoes dim
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
        }, 320 * i + 250));
      });

      const lastEcho = 320 * (N - 1) + 250;
      timeouts.push(setTimeout(() => { if (mounted) setStep(N + 1); }, lastEcho + 700));
      timeouts.push(setTimeout(() => { if (mounted) setStep(N + 2); }, lastEcho + 1300));

      timeouts.push(setTimeout(cycle, 8200));
    }

    cycle();
    return () => {
      mounted = false;
      timeouts.forEach(clearTimeout);
    };
  }, []);

  const ruleVisible    = step >= N + 1;
  const answerVisible  = step >= N + 2;
  const echoesDim      = step >= N + 2;

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
      {/* Echoes — cascading down the upper half */}
      <div style={{ position: 'absolute', inset: 0, padding: '18px 22px' }}>
        {ECHOES.map((e, i) => {
          const shown = step > i;
          const targetOpacity = shown
            ? echoesDim
              ? e.baseOpacity * 0.42
              : e.baseOpacity
            : 0;
          return (
            <motion.div
              key={e.id}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: targetOpacity, y: shown ? 0 : 6 }}
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
                gap: 10,
              }}
            >
              <span>{e.text}</span>
              <motion.span
                animate={{ opacity: echoesDim ? 0 : 1 }}
                transition={{ duration: 0.4 }}
                style={{
                  fontSize: Math.max(10, Math.round(e.size * 0.5)),
                  fontWeight: 400,
                  fontStyle: 'normal',
                  color: 'var(--text-tertiary)',
                  letterSpacing: 0,
                }}
              >
                {e.ts}
              </motion.span>
            </motion.div>
          );
        })}
      </div>

      {/* Rule with brand mark — halves sweep out from the centered logo */}
      <div
        style={{
          position: 'absolute',
          left: 22,
          right: 22,
          top: '54%',
          display: 'flex',
          alignItems: 'center',
          gap: 14,
        }}
      >
        <motion.div
          initial={{ scaleX: 0 }}
          animate={{ scaleX: ruleVisible ? 1 : 0 }}
          transition={{ duration: 0.6, delay: ruleVisible ? 0.18 : 0, ease: [0.22, 0.8, 0.36, 1] }}
          style={{
            flex: 1,
            height: 1,
            background: 'var(--text)',
            transformOrigin: 'right center',
            opacity: 0.9,
          }}
        />
        <motion.div
          initial={{ opacity: 0, scale: 0.7 }}
          animate={{ opacity: ruleVisible ? 1 : 0, scale: ruleVisible ? 1 : 0.7 }}
          transition={{ duration: 0.4, ease: [0.22, 0.8, 0.36, 1] }}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 7,
            color: 'var(--text)',
          }}
        >
          <Logo size={14} bg="var(--bg-card)" />
          <span
            style={{
              fontSize: 12,
              fontWeight: 700,
              letterSpacing: '0.02em',
            }}
          >
            Sunibgram
          </span>
        </motion.div>
        <motion.div
          initial={{ scaleX: 0 }}
          animate={{ scaleX: ruleVisible ? 1 : 0 }}
          transition={{ duration: 0.6, delay: ruleVisible ? 0.18 : 0, ease: [0.22, 0.8, 0.36, 1] }}
          style={{
            flex: 1,
            height: 1,
            background: 'var(--text)',
            transformOrigin: 'left center',
            opacity: 0.9,
          }}
        />
      </div>

      {/* Resolution — same words as the hero echo, same weight, answered */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: answerVisible ? 1 : 0, y: answerVisible ? 0 : 10 }}
        transition={{ duration: 0.5, ease: [0.22, 0.8, 0.36, 1] }}
        style={{
          position: 'absolute',
          left: 22,
          right: 22,
          top: '63%',
          display: 'flex',
          flexDirection: 'column',
          gap: 10,
        }}
      >
        <div
          style={{
            fontSize: 22,
            fontWeight: 600,
            color: 'var(--text)',
            letterSpacing: '-0.012em',
            lineHeight: 1.2,
          }}
        >
          {HERO}
        </div>

        <div
          style={{
            display: 'inline-flex',
            alignItems: 'baseline',
            gap: 10,
          }}
        >
          <span
            style={{
              fontSize: 14,
              fontWeight: 700,
              color: 'var(--accent)',
              letterSpacing: '-0.005em',
            }}
          >
            12 verified answers
          </span>
          <span style={{ color: 'var(--text-tertiary)', fontSize: 12 }}>
            on-topic · faculty feed
          </span>
        </div>
      </motion.div>
    </div>
  );
}
