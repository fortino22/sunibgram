export interface RefreshOk {
  status: 'ok';
  accessToken: string;
  sid: string;
}

export interface RefreshRevoked {
  status: 'revoked';
}

export interface RefreshTransient {
  status: 'transient';
}

export type RefreshResult = RefreshOk | RefreshRevoked | RefreshTransient;

interface RefreshPayload {
  authenticated?: boolean;
  access_token?: string;
  sid?: string;
  revoked?: boolean;
  transient?: boolean;
}

function isRefreshPayload(value: unknown): value is RefreshPayload {
  return typeof value === 'object' && value !== null;
}

let inflight: Promise<RefreshResult> | null = null;

async function doRefresh(): Promise<RefreshResult> {
  let response: Response;
  try {
    response = await fetch('/auth/refresh', { cache: 'no-store' });
  } catch {
    return { status: 'transient' };
  }

  let data: unknown = null;
  try {
    data = await response.json();
  } catch {
    return { status: 'transient' };
  }

  if (!isRefreshPayload(data)) {
    return { status: 'transient' };
  }

  if (data.authenticated && typeof data.access_token === 'string' && typeof data.sid === 'string') {
    return { status: 'ok', accessToken: data.access_token, sid: data.sid };
  }

  if (data.revoked === true) {
    return { status: 'revoked' };
  }

  if (data.transient === true || response.status >= 500) {
    return { status: 'transient' };
  }

  if (response.status === 401 || response.status === 403) {
    return { status: 'revoked' };
  }

  return { status: 'transient' };
}

export function refreshOnce(): Promise<RefreshResult> {
  if (inflight) return inflight;
  inflight = doRefresh().finally(() => {
    inflight = null;
  });
  return inflight;
}
