import type { ReactNode } from 'react'
import { Link } from 'react-router'

type MenuTileProps = {
  to: string
  icon: ReactNode
  label: string
}

// One cell of the Beranda "Kelola" grid. Show only the tiles the user's role may open.
export function MenuTile({ to, icon, label }: MenuTileProps) {
  return (
    <Link
      to={to}
      className="flex flex-col items-center gap-1.5 rounded-xl bg-card px-1 py-3 text-center text-2xs font-semibold text-ink active:opacity-80 [&_svg]:size-5.5 [&_svg]:stroke-[1.8]"
    >
      {icon}
      {label}
    </Link>
  )
}
