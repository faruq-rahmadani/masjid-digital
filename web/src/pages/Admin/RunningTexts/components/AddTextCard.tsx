import { type FormEvent, useState } from 'react'
import { Button, Card, Field, Textarea, useToast } from '@/components'
import { useCreateRunningText } from '@/hooks/useRunningTexts'
import { MAX_TEXT_LENGTH, TV_DELAY } from '../utils'

export function AddTextCard({ onAdded }: { onAdded: () => void }) {
  const [text, setText] = useState('')
  const create = useCreateRunningText()
  const toast = useToast()
  const empty = text.trim() === ''

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (empty) return
    create.mutate(text.trim(), {
      onSuccess: () => {
        setText('')
        onAdded()
        toast({ message: 'Running text ditambahkan', description: TV_DELAY })
      },
      onError: (error) => toast({ tone: 'danger', message: error.message }),
    })
  }

  return (
    <Card>
      <form onSubmit={submit} className="flex flex-col gap-2.5" noValidate>
        <Field label="Teks baru">
          <Textarea
            rows={3}
            maxLength={MAX_TEXT_LENGTH}
            showCount
            placeholder="Contoh: Kerja bakti bersih masjid Ahad, kumpul pukul 07.00"
            value={text}
            onChange={(event) => setText(event.currentTarget.value)}
          />
        </Field>
        <Button type="submit" size="lg" block loading={create.isPending} disabled={empty}>
          Tambah
        </Button>
      </form>
    </Card>
  )
}
