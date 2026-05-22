import type { Metadata } from 'next';
import './globals.css';
import { MainWrapper } from '@/components/layout/MainWrapper';
import { Providers } from '@/components/Providers';

export const metadata: Metadata = {
  title: 'Binusgram — Campus Discourse',
  description: 'University social media platform for Binus students',
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>
        <Providers>
          <MainWrapper>{children}</MainWrapper>
        </Providers>
      </body>
    </html>
  );
}
