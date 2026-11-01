import { type FormEvent, useState } from 'react'
import { Button, Field, Sheet, Textarea, useToast } from '@/components'
import { useUpdateRunningText } from '@/hooks/useRunningTexts'
import type { RunningText } from '@/types/runningText'
import { MAX_TEXT_LENGTH, TV_DELAY } from '../utils'

type EditTextDrawerProps = {
  text: RunningText
  onClose: () => void
  onDelete: (text: RunningText) => void
}

export function EditTextDrawer({ text, onClose, onDelete }: EditTextDrawerProps) {
  const [value, setValue] = useState(text.text)
  const update = useUpdateRunningText()
  const toast = useToast()
  const trimmed = value.trim()
  const unchanged = trimmed === text.text

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (trimmed === '' || unchanged) return
    update.mutate(
      { id: text.id, text: trimmed, active: text.active },
      {
        onSuccess: () => {
          toast({ message: 'Running text disimpan', description: TV_DELAY })
          onClose()
        },
        onError: (error) => toast({ tone: 'danger', message: error.message }),
      },
    )
  }

  return (
    <Sheet
      open
      onClose={onClose}
      title="Edit running text"
      footer={
        <div className="grid grid-cols-2 gap-2.5">
          <Button variant="danger" size="xl" onClick={() => onDelete(text)}>
            Hapus
          </Button>
          <Button
            type="submit"
            form="edit-running-text"
            size="xl"
            loading={update.isPending}
            disabled={trimmed === '' || unchanged}
          >
            Simpan
          </Button>
        </div>
      }
    >
      <form id="edit-running-text" onSubmit={submit} noValidate>
        <Field label="Teks">
          <Textarea
            rows={4}
            maxLength={MAX_TEXT_LENGTH}
            showCount
            value={value}
            onChange={(event) => setValue(event.currentTarget.value)}
          />
        </Field>
      </form>
    </Sheet>
  )
}
