// Mirrors NormalizePhone in internal/domain/user.go so the form rejects what the API would.

const SEPARATORS = /[\s\-.()]/g
const E164 = /^\+[1-9]\d{7,14}$/
const INDONESIAN_MOBILE = /^\+628\d{8,11}$/

/** "0812-3456-1001" → "+6281234561001", or null when invalid. */
export function normalizePhone(input: string): string | null {
  let phone = input.trim().replace(SEPARATORS, '')
  if (phone.startsWith('0')) phone = `+62${phone.slice(1)}`
  else if (phone.startsWith('62')) phone = `+${phone}`
  else if (!phone.startsWith('+')) return null

  if (!E164.test(phone)) return null
  if (phone.startsWith('+62') && !INDONESIAN_MOBILE.test(phone)) return null
  return phone
}

/** "+6281234561001" → "0812-3456-1001" for display. Other countries stay in E.164. */
export function formatPhone(e164: string): string {
  if (!e164.startsWith('+62')) return e164
  const local = `0${e164.slice(3)}`
  return [local.slice(0, 4), local.slice(4, 8), local.slice(8)].filter(Boolean).join('-')
}
