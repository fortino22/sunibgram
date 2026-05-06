import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

const SID_COOKIE_OPTIONS = {
  httpOnly: true,
  secure: process.env.NODE_ENV === 'production',
  sameSite: 'strict' as const,
  path: '/',
  maxAge: 60 * 60 * 24 * 7,
};

interface RefreshResponse {
  access_token: string;
  sid: string;
  access_token_expires_at: string;
}

function isRefreshResponse(value: unknown): value is RefreshResponse {
  if (typeof value !== 'object' || value === null) return false;

  const response = value as Record<string, unknown>;
  return (
    typeof response.access_token === 'string' &&
    typeof response.sid === 'string' &&
    typeof response.access_token_expires_at === 'string'
  );
}

export async function GET() {
  const cookieStore = await cookies();
  const sid = cookieStore.get('sid')?.value;

  if (!sid) {
    return NextResponse.json({ message: 'Unauthorized' }, { status: 401 });
  }

  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sid }),
  });

  if (!response.ok) {
    return NextResponse.json({ message: 'Unauthorized' }, { status: 401 });
  }

  const data: unknown = await response.json();

  if (!isRefreshResponse(data)) {
    return NextResponse.json({ message: 'Unauthorized' }, { status: 401 });
  }

  if (data.sid !== sid) {
    cookieStore.set('sid', data.sid, SID_COOKIE_OPTIONS);
  }

  return NextResponse.json({
    access_token: data.access_token,
    sid: data.sid,
  });
}
