'use client';

import { useRef, type ClipboardEvent, type KeyboardEvent } from 'react';

interface OtpInputProps {
  value: string[];
  onChange: (value: string[]) => void;
}

export function OtpInput({ value, onChange }: OtpInputProps) {
  const refs = useRef<(HTMLInputElement | null)[]>([]);

  const handleChange = (index: number, char: string) => {
    const digit = char.replace(/\D/g, '').slice(-1);
    const next = [...value];
    next[index] = digit;
    onChange(next);
    if (digit && index < 5) refs.current[index + 1]?.focus();
  };

  const handleKeyDown = (index: number, event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Backspace' && !value[index] && index > 0) {
      refs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (event: ClipboardEvent<HTMLInputElement>) => {
    event.preventDefault();
    const digits = event.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6);
    const next = [...value];
    digits.split('').forEach((digit, index) => {
      next[index] = digit;
    });
    onChange(next);
    refs.current[Math.min(digits.length, 5)]?.focus();
  };

  return (
    <div style={{ display: 'flex', gap: 10, justifyContent: 'center' }}>
      {value.map((digit, index) => (
        <input
          key={index}
          ref={(element) => {
            refs.current[index] = element;
          }}
          type="text"
          inputMode="numeric"
          maxLength={1}
          value={digit}
          onChange={(event) => handleChange(index, event.target.value)}
          onKeyDown={(event) => handleKeyDown(index, event)}
          onPaste={handlePaste}
          style={{
            width: 48,
            height: 56,
            textAlign: 'center',
            fontSize: 22,
            fontWeight: 600,
            background: 'var(--bg-subtle)',
            border: `1.5px solid ${digit ? 'var(--accent)' : 'var(--border)'}`,
            borderRadius: 12,
            color: 'var(--text)',
            outline: 'none',
            fontFamily: 'inherit',
            caretColor: 'var(--accent)',
          }}
        />
      ))}
    </div>
  );
}
