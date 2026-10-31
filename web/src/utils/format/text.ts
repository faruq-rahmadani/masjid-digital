// Small text helpers for display.

/** "[Ketua DKM]" → "KD", "Nama Bendahara" → "NB" */
export function initials(name: string) {
  return name
    .replace(/[^\p{L}\s]/gu, ' ')
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((word) => word.charAt(0).toUpperCase())
    .join('')
}
