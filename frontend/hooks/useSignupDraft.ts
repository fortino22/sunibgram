'use client';

import type { SignupInput, SignupStep } from '@/components/auth/schemas';

// Persists the signup form across page refreshes. sessionStorage (per decision):
// survives reload, auto-clears when the tab/browser closes, never written to
// disk long-term. Passwords ARE included here; avatar/banner are File objects
// and cannot be serialized, so they are re-picked after a refresh.
const DRAFT_KEY = 'signup-draft';

export interface SignupDraft {
  values: Partial<SignupInput>;
  step: SignupStep;
}

const EMPTY_DRAFT: SignupDraft = { values: {}, step: 1 };

export function loadDraft(): SignupDraft {
  if (typeof window === 'undefined') return EMPTY_DRAFT;
  try {
    const raw = window.sessionStorage.getItem(DRAFT_KEY);
    if (!raw) return EMPTY_DRAFT;
    const parsed = JSON.parse(raw) as Partial<SignupDraft>;
    const step = parsed.step === 2 || parsed.step === 3 ? parsed.step : 1;
    return { values: parsed.values ?? {}, step };
  } catch {
    return EMPTY_DRAFT;
  }
}

export function saveDraft(values: Partial<SignupInput>, step: SignupStep): void {
  if (typeof window === 'undefined') return;
  try {
    window.sessionStorage.setItem(DRAFT_KEY, JSON.stringify({ values, step }));
  } catch {
    // Quota / private-mode failures are non-fatal — the form still works in-memory.
  }
}

export function clearDraft(): void {
  if (typeof window === 'undefined') return;
  try {
    window.sessionStorage.removeItem(DRAFT_KEY);
  } catch {
    // ignore
  }
}
