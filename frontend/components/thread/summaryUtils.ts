/**
 * Splits raw summary text into clean bullet lines, stripping any leading list
 * markers ("- ", "* ", "• ", "1. ") the model emitted. Always returns at least
 * one entry so callers can map safely.
 */
export function toBullets(text: string): string[] {
  const lines = text
    .split('\n')
    .map((l) => l.replace(/^\s*(?:[-*•]|\d+[.)])\s*/, '').trim())
    .filter((l) => l.length > 0);
  return lines.length > 0 ? lines : [''];
}
