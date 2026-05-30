import { AboutNav } from '@/components/about/AboutNav';
import { AboutFooter } from '@/components/about/AboutFooter';

export default function AboutLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="about-root">
      <AboutNav />
      {children}
      <AboutFooter />
    </div>
  );
}
