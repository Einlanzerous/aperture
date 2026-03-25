/** Derive a 1–2 character uppercase initial string from a display name. */
export function getInitials(name: string): string {
  return name
    .split(/[\s\-_]+/)
    .map((w) => w[0] ?? '')
    .join('')
    .toUpperCase()
    .slice(0, 2)
}
