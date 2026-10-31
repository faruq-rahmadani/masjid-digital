import { EllipsisVertical } from 'lucide-react'
import { useState } from 'react'
import { IconButton } from '@/components/actions/IconButton'
import { Sheet } from '@/components/overlays/Sheet'
import { cn } from '@/utils/cn'

export type Action = {
  label: string
  onSelect: () => void
  /** Destructive actions (nonaktifkan, hapus) are shown in rust. */
  danger?: boolean
  disabled?: boolean
}

type ActionMenuProps = {
  /** Names the row, e.g. the user's name; the button reads "Aksi untuk {label}". */
  label: string
  actions: Action[]
}

// The ⋯ button at the end of a list row. Opens its actions in a bottom sheet (thumb-reachable).
export function ActionMenu({ label, actions }: ActionMenuProps) {
  const [open, setOpen] = useState(false)

  return (
    <>
      <IconButton
        variant="ghost"
        label={`Aksi untuk ${label}`}
        icon={<EllipsisVertical />}
        aria-haspopup="dialog"
        onClick={() => setOpen(true)}
      />
      <Sheet open={open} onClose={() => setOpen(false)} title={label}>
        <ul className="divide-line-soft bg-card divide-y rounded-xl px-4">
          {actions.map((action) => (
            <li key={action.label}>
              <button
                type="button"
                disabled={action.disabled}
                onClick={() => {
                  setOpen(false)
                  action.onSelect()
                }}
                className={cn(
                  'min-h-touch w-full py-3 text-left text-base font-semibold disabled:opacity-50',
                  action.danger && 'text-expense',
                )}
              >
                {action.label}
              </button>
            </li>
          ))}
        </ul>
      </Sheet>
    </>
  )
}
