import { useId, useState, type ReactNode } from 'react'
import { Button } from '@/components/actions/Button'
import { Textarea } from '@/components/forms/Textarea'
import { Sheet } from './Sheet'

type ConfirmDialogProps = {
  open: boolean
  onClose: () => void
  title: string
  description?: ReactNode
  confirmLabel: string
  /** Rust confirm button for destructive actions (hapus, nonaktifkan, tutup buku). */
  danger?: boolean
  /** PRD: editing or deleting a transaction needs a reason. The confirm stays disabled until one is typed. */
  requireReason?: boolean
  onConfirm: (reason: string) => void
}

// Confirmation as a bottom sheet so the buttons stay in thumb reach.
export function ConfirmDialog({
  open,
  onClose,
  title,
  description,
  confirmLabel,
  danger = false,
  requireReason = false,
  onConfirm,
}: ConfirmDialogProps) {
  const [reason, setReason] = useState('')
  const reasonId = useId()
  const blocked = requireReason && reason.trim() === ''

  const close = () => {
    setReason('')
    onClose()
  }

  return (
    <Sheet
      open={open}
      onClose={close}
      title={title}
      footer={
        <div className="grid grid-cols-2 gap-2.5">
          <Button variant="secondary" onClick={close}>
            Batal
          </Button>
          <Button
            variant={danger ? 'danger' : 'primary'}
            disabled={blocked}
            onClick={() => {
              onConfirm(reason.trim())
              close()
            }}
          >
            {confirmLabel}
          </Button>
        </div>
      }
    >
      {description && <div className="text-base text-ink-muted">{description}</div>}
      {requireReason && (
        <div className="mt-4 flex flex-col gap-1.5">
          <label htmlFor={reasonId} className="text-xs font-semibold text-ink-soft">
            Alasan (wajib)
          </label>
          <Textarea
            id={reasonId}
            value={reason}
            onChange={(event) => setReason(event.currentTarget.value)}
            maxLength={200}
            showCount
          />
        </div>
      )}
    </Sheet>
  )
}
