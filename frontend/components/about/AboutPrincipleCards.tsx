'use client';

import { motion } from 'framer-motion';
import {
  ShieldCheck,
  Users,
  GraduationCap,
  EyeOff,
  Compass,
  HeartHandshake,
  type LucideIcon,
} from 'lucide-react';

type Principle = {
  Icon: LucideIcon;
  title: string;
  body: string;
};

const PRINCIPLES: Principle[] = [
  {
    Icon: ShieldCheck,
    title: 'Verified, always',
    body:
      'Every account is tied to a real @binus.ac.id email. No bots, no anonymous trolls, no off-campus strangers.',
  },
  {
    Icon: Users,
    title: 'Built around campus',
    body:
      'Faculty, batch, and community are first-class. The feed knows the difference between a classmate and a stranger.',
  },
  {
    Icon: GraduationCap,
    title: 'Student-first',
    body:
      'Made by Binus students, for Binus students. We ship what we ourselves wished existed during lectures.',
  },
  {
    Icon: EyeOff,
    title: 'No engagement-bait',
    body:
      'No recommended outrage, no dark patterns, no infinite vanity metrics. The feed is for catching up, not doom-scrolling.',
  },
  {
    Icon: Compass,
    title: 'Easy to find your people',
    body:
      'Communities, faculty feeds, and search make it simple to land in the right room — even on your first week.',
  },
  {
    Icon: HeartHandshake,
    title: 'Honest about what we are',
    body:
      'Sunibgram is a student-built project, not an official Binus product. Independence keeps us honest.',
  },
];

export function AboutPrincipleCards() {
  return (
    <div className="about-card-grid">
      {PRINCIPLES.map(({ Icon, title, body }, i) => (
        <motion.div
          key={title}
          className="about-info-card"
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-80px' }}
          transition={{ duration: 0.45, delay: i * 0.07, ease: [0.22, 0.8, 0.36, 1] }}
        >
          <motion.div
            className="about-info-card-icon"
            whileHover={{ scale: 1.08, rotate: -4 }}
            animate={{ y: [0, -2, 0] }}
            transition={{
              y: { duration: 3.6 + (i % 3) * 0.5, repeat: Infinity, ease: 'easeInOut', delay: i * 0.18 },
            }}
          >
            <Icon size={20} strokeWidth={2} />
          </motion.div>
          <div className="about-info-card-title">{title}</div>
          <div className="about-info-card-desc">{body}</div>
        </motion.div>
      ))}
    </div>
  );
}
