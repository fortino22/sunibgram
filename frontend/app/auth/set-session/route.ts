import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

const SID_COOKIE_OPTIONS = {
  httpOnly: true,
  secure: process.env.NODE_ENV === 'production',
  sameSite: 'lax' as const,
  path: '/',
  maxAge: 60 * 60 * 24 * 7,
};

interface SetSessionBody {
  sid?: unknown;
}

export async function POST(request: Request) {
  const body = (await request.json()) as SetSessionBody;

  if (typeof body.sid !== 'string' || body.sid.length === 0) {
    return NextResponse.json({ message: 'Missing sid' }, { status: 400 });
  }

  const cookieStore = await cookies();
  cookieStore.set('sid', body.sid, SID_COOKIE_OPTIONS);

  return NextResponse.json({ ok: true });
}
