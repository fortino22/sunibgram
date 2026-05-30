import type { Metadata } from 'next';
import { AboutSplitRoles } from '@/components/about/AboutSplitRoles';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';

export const metadata: Metadata = {
  title: 'The Team — Sunibgram',
  description:
    'Meet the students building Sunibgram — a campus social network made by Binus, for Binus.',
};

export default function TeamPage() {
  return (
    <>
      <AboutSplitRoles />
      <AboutFinalCTA />
    </>
  );
}
