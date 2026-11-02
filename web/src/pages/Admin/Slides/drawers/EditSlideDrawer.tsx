import { type FormEvent, useState } from 'react'
import { Button, Sheet, SlideFrame, useToast } from '@/components'
import { useUpdateSlide } from '@/hooks/useSlides'
import { slideImageUrl } from '@/services/slideService'
import type { Slide } from '@/types/slide'
import { SlideFields, trimmed } from '../components/SlideFields'
import { TV_DELAY } from '../utils'

type EditSlideDrawerProps = {
  slide: Slide
  onClose: () => void
  onDelete: (slide: Slide) => void
}

// The image can't be replaced: a new picture is a new slide (upload it, then delete this one).
export function EditSlideDrawer({ slide, onClose, onDelete }: EditSlideDrawerProps) {
  const [value, setValue] = useState({
    title: slide.title,
    caption: slide.caption,
    duration_seconds: slide.duration_seconds,
    show_caption: slide.show_caption,
  })
  const update = useUpdateSlide()
  const toast = useToast()
  const body = trimmed(value)
  const unchanged =
    body.title === slide.title &&
    body.caption === slide.caption &&
    body.duration_seconds === slide.duration_seconds &&
    body.show_caption === slide.show_caption

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (body.title === '' || unchanged) return
    update.mutate(
      { id: slide.id, ...body, active: slide.active },
      {
        onSuccess: () => {
          toast({ message: 'Slide disimpan', description: TV_DELAY })
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
      title="Edit slide"
      footer={
        <div className="grid grid-cols-2 gap-2.5">
          <Button variant="danger" size="xl" onClick={() => onDelete(slide)}>
            Hapus
          </Button>
          <Button
            type="submit"
            form="edit-slide"
            size="xl"
            loading={update.isPending}
            disabled={body.title === '' || unchanged}
          >
            Simpan
          </Button>
        </div>
      }
    >
      <form id="edit-slide" onSubmit={submit} noValidate>
        <SlideFields
          preview={
            <SlideFrame
              size="preview"
              src={slideImageUrl(slide.id)}
              title={body.title}
              caption={body.caption}
              showCaption={body.show_caption}
            />
          }
          previewHint="Pratinjau di layar TV. Posisi gambar tidak bisa diubah; upload ulang untuk memotong lain."
          value={value}
          onChange={setValue}
        />
      </form>
    </Sheet>
  )
}
