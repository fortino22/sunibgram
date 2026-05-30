import type { Metadata } from 'next';
import { AboutUseCases } from '@/components/about/AboutUseCases';
import { AboutStatMoment } from '@/components/about/AboutStatMoment';

export const metadata: Metadata = {
  title: 'Use Cases — Sunibgram',
  description:
    'See how Binus students use Sunibgram every day — from study groups to event planning.',
};

export default function UseCasesPage() {
  return (
    <>
      <AboutUseCases />
      <AboutStatMoment />
    </>
  );
}
