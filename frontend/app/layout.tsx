import type { Metadata } from 'next';
import { Orbitron } from 'next/font/google';
import './globals.css';
import { MainWrapper } from '@/components/layout/MainWrapper';
import { Providers } from '@/components/Providers';

const orbitron = Orbitron({
  subsets: ['latin'],
  weight: ['600', '700', '800', '900'],
  variable: '--font-orbitron',
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'Sunibgram',
  description: 'University social media platform for Binus students',
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className={orbitron.variable}>
      <body>
        <Providers>
          <MainWrapper>{children}</MainWrapper>
        </Providers>
      </body>
    </html>
  );
}
