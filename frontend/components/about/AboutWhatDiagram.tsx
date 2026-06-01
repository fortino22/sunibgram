'use client';

import { motion } from 'framer-motion';
import {
  MessageCircle,
  Users,
  CalendarDays,
  Send,
  Trophy,
  AtSign,
} from 'lucide-react';
import { Logo } from '@/components/ui/Logo';

/* Positions in percent of container (0–100). The hub sits dead-center.
   Six surrounding nodes arranged in a balanced radial layout. */
const HUB = { x: 50, y: 50 };
const NODES = [
  { Icon: MessageCircle, label: 'Threads',     x: 18, y: 18 },
  { Icon: Users,         label: 'Communities', x: 82, y: 18 },
  { Icon: AtSign,        label: '@binus.ac.id', x: 6,  y: 50 },
  { Icon: Send,          label: 'Messages',    x: 94, y: 50 },
  { Icon: Trophy,        label: 'Leaderboard', x: 18, y: 82 },
  { Icon: CalendarDays,  label: 'Events',      x: 82, y: 82 },
] as const;

export function AboutWhatDiagram() {
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
      {/* Grid background — subtle, fades toward center */}
      <div
        aria-hidden
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage:
            'linear-gradient(to right, color-mix(in oklab, var(--border) 60%, transparent) 1px, transparent 1px), linear-gradient(to bottom, color-mix(in oklab, var(--border) 60%, transparent) 1px, transparent 1px)',
          backgroundSize: '32px 32px',
          maskImage: 'radial-gradient(ellipse 60% 60% at 50% 50%, black 40%, transparent 80%)',
          WebkitMaskImage: 'radial-gradient(ellipse 60% 60% at 50% 50%, black 40%, transparent 80%)',
        }}
      />

      {/* Soft accent glow behind the hub */}
      <motion.div
        aria-hidden
        animate={{ opacity: [0.45, 0.7, 0.45] }}
        transition={{ duration: 5, repeat: Infinity, ease: 'easeInOut' }}
        style={{
          position: 'absolute',
          left: '50%',
          top: '50%',
          width: 280,
          height: 280,
          transform: 'translate(-50%, -50%)',
          background:
            'radial-gradient(circle, color-mix(in oklab, var(--accent) 22%, transparent) 0%, transparent 60%)',
          pointerEvents: 'none',
        }}
      />

      {/* Connection lines — SVG, drawn from the hub to each node */}
      <svg
        viewBox="0 0 100 100"
        preserveAspectRatio="none"
        style={{ position: 'absolute', inset: 0, width: '100%', height: '100%', pointerEvents: 'none' }}
      >
        <defs>
          <linearGradient id="what-line" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%"  stopColor="var(--accent)"  stopOpacity="0.7" />
            <stop offset="100%" stopColor="var(--accent)" stopOpacity="0.15" />
          </linearGradient>
        </defs>

        {NODES.map((n, i) => (
          <motion.line
            key={`line-${n.label}`}
            x1={HUB.x}
            y1={HUB.y}
            x2={n.x}
            y2={n.y}
            stroke="url(#what-line)"
            strokeWidth="0.4"
            vectorEffect="non-scaling-stroke"
            initial={{ pathLength: 0, opacity: 0 }}
            animate={{ pathLength: 1, opacity: 1 }}
            transition={{ delay: 0.4 + i * 0.08, duration: 0.8, ease: 'easeOut' }}
          />
        ))}

        {/* Travelling dots — one per line, looping along the path */}
        {NODES.map((n, i) => (
          <motion.circle
            key={`dot-${n.label}`}
            r="0.6"
            fill="var(--accent)"
            initial={{ opacity: 0 }}
            animate={{
              opacity: [0, 1, 1, 0],
              cx: [HUB.x, n.x],
              cy: [HUB.y, n.y],
            }}
            transition={{
              duration: 2.2,
              repeat: Infinity,
              repeatDelay: 1.8,
              delay: 1.6 + i * 0.35,
              ease: 'easeInOut',
              times: [0, 0.1, 0.9, 1],
            }}
          />
        ))}
      </svg>

      {/* Hub — Sunibgram logo */}
      <motion.div
        initial={{ opacity: 0, scale: 0.6 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ delay: 0.1, duration: 0.5, ease: [0.22, 0.8, 0.36, 1] }}
        style={{
          position: 'absolute',
          left: `${HUB.x}%`,
          top: `${HUB.y}%`,
          transform: 'translate(-50%, -50%)',
          zIndex: 2,
        }}
      >
        <motion.div
          animate={{ scale: [1, 1.04, 1] }}
          transition={{ duration: 4, repeat: Infinity, ease: 'easeInOut' }}
          style={{
            width: 88,
            height: 88,
            borderRadius: 24,
            background: 'var(--bg-card)',
            border: '1px solid var(--border)',
            boxShadow: '0 12px 32px -8px color-mix(in oklab, var(--accent) 35%, transparent)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <Logo size={44} bg="var(--bg-card)" />
        </motion.div>
        <div
          style={{
            position: 'absolute',
            left: '50%',
            top: 'calc(100% + 10px)',
            transform: 'translateX(-50%)',
            fontSize: 12,
            fontWeight: 700,
            color: 'var(--text)',
            letterSpacing: '0.04em',
            whiteSpace: 'nowrap',
          }}
        >
          Sunibgram
        </div>
      </motion.div>

      {/* Surrounding nodes — chip with icon + label */}
      {NODES.map((n, i) => (
        <motion.div
          key={n.label}
          initial={{ opacity: 0, scale: 0.7 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.55 + i * 0.08, duration: 0.45, ease: [0.22, 0.8, 0.36, 1] }}
          style={{
            position: 'absolute',
            left: `${n.x}%`,
            top: `${n.y}%`,
            transform: 'translate(-50%, -50%)',
            zIndex: 2,
          }}
        >
          <motion.div
            animate={{ y: [0, -3, 0] }}
            transition={{
              duration: 4 + (i % 3) * 0.6,
              repeat: Infinity,
              ease: 'easeInOut',
              delay: i * 0.2,
            }}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 8,
              padding: '8px 12px',
              borderRadius: 999,
              border: '1px solid var(--border)',
              background: 'var(--bg-card)',
              boxShadow: '0 4px 12px -6px color-mix(in oklab, var(--text) 18%, transparent)',
              whiteSpace: 'nowrap',
            }}
          >
            <span
              style={{
                display: 'inline-flex',
                width: 24,
                height: 24,
                borderRadius: 8,
                background: 'var(--accent-soft)',
                color: 'var(--accent)',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <n.Icon size={13} strokeWidth={2.4} />
            </span>
            <span style={{ fontSize: 12, fontWeight: 700, color: 'var(--text)', letterSpacing: '-0.005em' }}>
              {n.label}
            </span>
          </motion.div>
        </motion.div>
      ))}
    </div>
  );
}
