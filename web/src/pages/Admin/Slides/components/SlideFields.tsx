import type { ReactNode } from 'react'
import { Checkbox, Field, Input, Select, Textarea } from '@/components'
import type { CreateSlideRequest } from '@/types/slide'
import { DURATION_OPTIONS, MAX_CAPTION_LENGTH, MAX_TITLE_LENGTH } from '../utils'

type SlideFieldsProps = {
  /** How the TV will show the slide: the cropper for a new one, a SlideFrame when editing. */
  preview: ReactNode
  /** Under the preview. */
  previewHint: string
  value: CreateSlideRequest
  onChange: (value: CreateSlideRequest) => void
}

// The details of a slide, shared by the new-slide and edit sheets, under a live preview of how
// the TV will show it.
export function SlideFields({ preview, previewHint, value, onChange }: SlideFieldsProps) {
  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex flex-col gap-1.5">
        {preview}
        <p className="text-ink-muted text-xs">{previewHint}</p>
      </div>
      {/* the hint never changes, so ticking the box doesn't resize the sheet */}
      <div>
        <Checkbox
          label="Tampilkan judul & keterangan di TV"
          checked={value.show_caption}
          onChange={(event) => onChange({ ...value, show_caption: event.currentTarget.checked })}
        />
        <p className="text-ink-muted -mt-1 text-xs">
          Matikan untuk poster yang sudah ada tulisannya. Judul tetap jadi nama di daftar slide.
        </p>
      </div>
      <Field label="Judul">
        <Input
          maxLength={MAX_TITLE_LENGTH}
          value={value.title}
          onChange={(event) => onChange({ ...value, title: event.currentTarget.value })}
        />
      </Field>
      <Field label="Keterangan (opsional)">
        <Textarea
          rows={2}
          maxLength={MAX_CAPTION_LENGTH}
          showCount
          placeholder="Contoh: Transfer ke BSI 7123456789 a.n. Masjid"
          disabled={!value.show_caption}
          value={value.caption}
          onChange={(event) => onChange({ ...value, caption: event.currentTarget.value })}
        />
      </Field>
      <Field label="Durasi">
        <Select
          options={DURATION_OPTIONS}
          value={String(value.duration_seconds)}
          onChange={(duration) => onChange({ ...value, duration_seconds: Number(duration) })}
        />
      </Field>
    </div>
  )
}

/** The fields as the API takes them: trimmed. */
export const trimmed = (value: CreateSlideRequest): CreateSlideRequest => ({
  ...value,
  title: value.title.trim(),
  caption: value.caption.trim(),
})
