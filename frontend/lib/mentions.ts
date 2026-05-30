// Mentions are stored in thread content as `@[username|id]` tokens. The picker
// inserts clean `@username` while composing and tracks the id separately; the
// token is produced on submit (encode) and parsed for display (decode). These
// helpers give the same round-trip to the inline edit editor so the raw token
// never leaks into the textarea.

// Matches a stored mention token globally: @[username|id]
const TOKEN_RE = /@\[([a-zA-Z0-9_.]+)\|(\d+)\]/g;

// Matches a plain `@username` in editor text.
const PLAIN_MENTION_RE = /@([a-zA-Z0-9_.]+)/g;

/** Stored content -> editor display. Strips `|id`, returns a username->id map. */
export function decodeMentionsForEditing(content: string): {
  text: string;
  mentionMap: Map<string, number>;
} {
  const mentionMap = new Map<string, number>();
  const text = content.replace(TOKEN_RE, (_full, name: string, id: string) => {
    mentionMap.set(name, Number(id));
    return `@${name}`;
  });
  return { text, mentionMap };
}

/** Editor display -> stored content. Re-tokenizes mentions present in the map. */
export function encodeMentions(text: string, mentionMap: Map<string, string | number>): string {
  return text.replace(PLAIN_MENTION_RE, (full, name: string) => {
    const id = mentionMap.get(name);
    return id ? `@[${name}|${id}]` : full;
  });
}
