import type { Metadata } from 'next';
import { AboutHero } from '@/components/about/AboutHero';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';

export const metadata: Metadata = {
  title: 'Sunibgram — The social network for Binus',
  description:
    'Sunibgram is the verified social network for Binus University students. Threads, communities, events, and chat — one campus, one app.',
};

export default function AboutPage() {
  return (
    <>
      <AboutHero />
      <AboutFinalCTA />
    </>
  );
}
