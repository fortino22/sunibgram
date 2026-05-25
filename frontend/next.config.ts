import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  output: 'standalone',
  async rewrites() {
    return {
      beforeFiles: [
        { source: '/api/auth/set-session', destination: '/api/auth/set-session' },
        { source: '/api/auth/clear-session', destination: '/api/auth/clear-session' },
        { source: '/api/auth/refresh', destination: '/api/auth/refresh' },
      ],
      afterFiles: [
        {
          source: '/api/:path*',
          destination: `${process.env.API_URL || 'http://localhost:8080'}/api/:path*`,
        },
      ],
      fallback: [],
    };
  },
};

export default nextConfig;
