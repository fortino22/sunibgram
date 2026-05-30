'use client';

import { useEffect } from 'react';
import { THEME_STORAGE_KEY, useUserSettings } from '@/hooks/useUserSettings';

/**
 * Syncs the resolved user theme onto <html data-theme> and caches it in
 * localStorage so the pre-paint inline script in app/layout.tsx can restore it
 * on the next load without a flash (FOUC). Renders nothing.
 */
export function ThemeSync() {
  const { data: settings } = useUserSettings();
  const theme = settings?.user_theme;

  useEffect(() => {
    if (!theme) return;
    document.documentElement.dataset.theme = theme;
    try {
      localStorage.setItem(THEME_STORAGE_KEY, theme);
    } catch {
      // ignore storage write failures (private mode, quota)
    }
  }, [theme]);

  return null;
}
