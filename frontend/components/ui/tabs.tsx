'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';

interface TabsContextValue {
  value: string;
  onValueChange: (value: string) => void;
}

const TabsContext = React.createContext<TabsContextValue | null>(null);

export function Tabs({
  value,
  onValueChange,
  className,
  children,
}: {
  value: string;
  onValueChange: (value: string) => void;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <TabsContext.Provider value={{ value, onValueChange }}>
      <div className={className}>{children}</div>
    </TabsContext.Provider>
  );
}

function useTabsContext() {
  const value = React.useContext(TabsContext);
  if (value === null) {
    throw new Error('Tabs components must be used within <Tabs />');
  }
  return value;
}

export function TabsList({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn('inline-flex rounded-2xl border p-1', className)}
      style={{ borderColor: 'var(--border)', background: 'var(--bg-subtle)' }}
      {...props}
    />
  );
}

export function TabsTrigger({
  value,
  className,
  children,
}: {
  value: string;
  className?: string;
  children: React.ReactNode;
}) {
  const tabs = useTabsContext();
  const active = tabs.value === value;
  const [hovered, setHovered] = React.useState(false);

  return (
    <button
      type="button"
      className={cn(
        'rounded-xl px-4 py-2 text-sm font-medium transition-colors',
        className
      )}
      style={{
        background: active ? 'var(--bg-card)' : hovered ? 'var(--bg-hover)' : 'transparent',
        color: active || hovered ? 'var(--text)' : 'var(--text-secondary)',
      }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onClick={() => tabs.onValueChange(value)}
    >
      {children}
    </button>
  );
}

export function TabsContent({
  value,
  className,
  children,
}: {
  value: string;
  className?: string;
  children: React.ReactNode;
}) {
  const tabs = useTabsContext();
  if (tabs.value !== value) {
    return null;
  }
  return <div className={className}>{children}</div>;
}
