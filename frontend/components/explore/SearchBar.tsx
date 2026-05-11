'use client';

import { useRef, useState } from 'react';

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function SearchBar({ value, onChange, placeholder = 'Search' }: SearchBarProps) {
  const [focused, setFocused] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        background: 'var(--bg-subtle)',
        border: focused ? '1px solid var(--accent)' : '1px solid transparent',
        borderRadius: 9999,
        padding: '8px 16px',
        transition: 'border-color 0.15s',
        flex: 1,
      }}
    >
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        style={{
          color: focused ? 'var(--accent)' : 'var(--text-tertiary)',
          flexShrink: 0,
          transition: 'color 0.15s',
        }}
      >
        <circle cx="11" cy="11" r="8" stroke="currentColor" strokeWidth="2" />
        <path d="M21 21l-4.35-4.35" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      </svg>
      <input
        ref={inputRef}
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        placeholder={placeholder}
        style={{
          flex: 1,
          background: 'transparent',
          border: 'none',
          outline: 'none',
          fontSize: 15,
          color: 'var(--text)',
          fontFamily: 'inherit',
        }}
      />
      {value && (
        <button
          onMouseDown={(e) => {
            e.preventDefault();
            onChange('');
            inputRef.current?.focus();
          }}
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: 18,
            height: 18,
            borderRadius: '50%',
            background: 'var(--accent)',
            border: 'none',
            cursor: 'pointer',
            flexShrink: 0,
            color: 'white',
            fontSize: 10,
            fontWeight: 700,
          }}
          aria-label="Clear search"
        >
          ✕
        </button>
      )}
    </div>
  );
}
