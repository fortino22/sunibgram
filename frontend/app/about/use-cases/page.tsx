import type { Metadata } from 'next';
import { AboutPageHero } from '@/components/about/AboutPageHero';
import { AboutUseCases } from '@/components/about/AboutUseCases';
import { AboutStatMoment } from '@/components/about/AboutStatMoment';
import { AboutFinalCTA } from '@/components/about/AboutFinalCTA';

export const metadata: Metadata = {
  title: 'Use Cases — Sunibgram',
  description:
    'See how Binus students use Sunibgram every day — from finding study spots to running clubs and organizing events.',
};

export default function UseCasesPage() {
  return (
    <>
      <AboutPageHero
        eyebrow="Use cases"
        title="However you show up on campus."
        description="Sunibgram is one app for first-years finding their faculty, clubs running their rooms, study groups coordinating, and organizers running real events."
        primaryCta={{ label: 'Join with your Binus email', href: '/signup' }}
        secondaryCta={{ label: 'See the features', href: '/about/threads' }}
      />

      <AboutUseCases />

      <AboutStatMoment />
      <AboutFinalCTA />
    </>
  );
}
