import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

const SID_COOKIE_OPTIONS = {
  httpOnly: true,
  secure: process.env.NODE_ENV === 'production',
  sameSite: 'lax' as const,
  path: '/',
  maxAge: 60 * 60 * 24 * 7,
};

interface RefreshResponse {
  access_token: string;
  sid: string;
}

function isRefreshResponse(value: unknown): value is RefreshResponse {
  if (typeof value !== 'object' || value === null) return false;

  const response = value as Record<string, unknown>;
  return (
    typeof response.access_token === 'string' &&
    typeof response.sid === 'string'
  );
}

export async function GET() {
  try {
    const cookieStore = await cookies();
    const sid = cookieStore.get('sid')?.value;

    if (!sid) {
      return NextResponse.json({ authenticated: false });
    }

    const baseUrl = (process.env.API_URL ?? process.env.NEXT_PUBLIC_API_URL ?? '').replace(/\/+$/, '');
    if (!baseUrl) {
      return NextResponse.json({ message: 'Server misconfiguration' }, { status: 500 });
    }

    const response = await fetch(`${baseUrl}/api/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sid }),
    });

    if (!response.ok) {
      cookieStore.delete('sid');
      return NextResponse.json({ authenticated: false });
    }

    const data: unknown = await response.json();

    if (!isRefreshResponse(data)) {
      cookieStore.delete('sid');
      return NextResponse.json({ authenticated: false });
    }

    if (data.sid !== sid) {
      cookieStore.set('sid', data.sid, SID_COOKIE_OPTIONS);
    }

    return NextResponse.json({
      authenticated: true,
      access_token: data.access_token,
      sid: data.sid,
    });
  } catch (err) {
    console.error('[/api/auth/refresh]', err);
    return NextResponse.json({ message: 'Internal server error' }, { status: 500 });
  }
}
