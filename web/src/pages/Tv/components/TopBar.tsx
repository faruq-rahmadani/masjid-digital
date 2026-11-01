import type { ReactNode } from 'react'

// Title row of the full-screen modes (adzan, iqamah, jumat): mosque name left, time right.
export function TopBar({ name, right }: { name: string; right: ReactNode }) {
  return (
    <div className="text-tv-30 flex items-center justify-between font-bold">
      <div>{name}</div>
      <div>{right}</div>
    </div>
  )
}
