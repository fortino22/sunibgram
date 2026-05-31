import type { Metadata } from 'next';
import { AboutHero } from '@/components/about/AboutHero';
import { AboutSplitConnection } from '@/components/about/AboutSplitConnection';

export const metadata: Metadata = {
  title: 'Sunibgram — The social network for Binus',
  description:
    'Sunibgram is the verified social network for Binus University students. Threads, communities, events, and chat — one campus, one app.',
};

export default function AboutPage() {
  return (
    <div className="about-fold">
      <AboutHero />
      <AboutSplitConnection />
    </div>
  );
}
