import type { Metadata } from 'next';
import { AboutNav } from '@/components/about/AboutNav';
import { AboutHero } from '@/components/about/AboutHero';
import { AboutSplitValue } from '@/components/about/AboutSplitValue';
import { AboutSplitConnection } from '@/components/about/AboutSplitConnection';
import { AboutUseCases } from '@/components/about/AboutUseCases';
import { AboutStatMoment } from '@/components/about/AboutStatMoment';
import { AboutSplitRoles } from '@/components/about/AboutSplitRoles';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';
import { AboutFooter } from '@/components/about/AboutFooter';

export const metadata: Metadata = {
  title: 'Sunibgram — The social network for Binus',
  description:
    'Sunibgram is the verified social network for Binus University students. Threads, communities, events, and chat — one campus, one app.',
};

export default function AboutPage() {
  return (
    <div className="about-root" data-theme="white">
      <AboutNav />
      <AboutHero />
      <AboutSplitValue />
      <AboutSplitConnection />
      <AboutUseCases />
      <AboutStatMoment />
      <AboutSplitRoles />
      <AboutFinalCTA />
      <AboutFooter />
    </div>
  );
}
