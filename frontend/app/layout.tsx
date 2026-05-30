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

// Runs before first paint: restores the cached theme onto <html data-theme>
// so the white theme doesn't flash to black on load. Kept in sync by ThemeSync.
const themeInitScript = `(function(){try{var t=localStorage.getItem('sunibgram-theme');if(t==='white'||t==='black'){document.documentElement.dataset.theme=t;}}catch(e){}})();`;

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className={orbitron.variable} suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body>
        <Providers>
          <MainWrapper>{children}</MainWrapper>
        </Providers>
      </body>
    </html>
  );
}
