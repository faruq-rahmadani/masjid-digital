import type { ReactNode } from 'react'
import { Link } from 'react-router'

type MenuTileProps = {
  to: string
  icon: ReactNode
  label: string
}

// One cell of the Home "Kelola" grid. Show only the tiles the user's role may open.
export function MenuTile({ to, icon, label }: MenuTileProps) {
  return (
    <Link
      to={to}
      className="bg-card text-2xs text-ink flex flex-col items-center gap-1.5 rounded-xl px-1 py-3 text-center font-semibold active:opacity-80 [&_svg]:size-5.5 [&_svg]:stroke-[1.8]"
    >
      {icon}
      {label}
    </Link>
  )
}
