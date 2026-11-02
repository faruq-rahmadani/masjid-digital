import { type FormEvent, useEffect, useState } from 'react'
import Cropper, { type Area, type Point } from 'react-easy-crop'
import { Button, Sheet, SLIDE_ASPECT, SlideCaptionBar, Slider, useToast } from '@/components'
import { useCreateSlide } from '@/hooks/useSlides'
import { SlideFields, trimmed } from '../components/SlideFields'
import { cropImage } from '../cropImage'
import { TV_DELAY, titleFromFileName } from '../utils'

type NewSlideDrawerProps = {
  file: File
  onClose: () => void
  onCreated: () => void
}

const MAX_ZOOM = 3

// The admin frames the picked image to the TV's slide panel (drag, pinch or the zoom slider), with
// the caption bar laid over it as the TV will. Nothing is uploaded until Simpan, which uploads
// only the framed part, scaled down (cropImage).
export function NewSlideDrawer({ file, onClose, onCreated }: NewSlideDrawerProps) {
  const [value, setValue] = useState(() => ({
    title: titleFromFileName(file.name),
    caption: '',
    duration_seconds: 10,
    show_caption: true,
  }))
  const [preview, setPreview] = useState('')
  const [crop, setCrop] = useState<Point>({ x: 0, y: 0 })
  const [zoom, setZoom] = useState(1)
  // in the image's pixels; null until the cropper has measured the image
  const [area, setArea] = useState<Area | null>(null)
  const [cropping, setCropping] = useState(false)
  const create = useCreateSlide()
  const toast = useToast()
  const body = trimmed(value)

  useEffect(() => {
    const url = URL.createObjectURL(file)
    setPreview(url)
    return () => URL.revokeObjectURL(url)
  }, [file])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    if (body.title === '' || !area) return

    let cropped: File
    setCropping(true)
    try {
      cropped = await cropImage(file, area)
    } catch (error) {
      toast({ tone: 'danger', message: (error as Error).message })
      return
    } finally {
      setCropping(false)
    }

    create.mutate(
      { file: cropped, ...body },
      {
        onSuccess: () => {
          toast({ message: 'Slide ditambahkan', description: TV_DELAY })
          onCreated()
        },
        onError: (error) => toast({ tone: 'danger', message: error.message }),
      },
    )
  }

  return (
    <Sheet
      open
      onClose={onClose}
      title="Slide baru"
      footer={
        <div className="grid grid-cols-2 gap-2.5">
          <Button variant="outline" size="xl" onClick={onClose}>
            Batal
          </Button>
          <Button
            type="submit"
            form="new-slide"
            size="xl"
            loading={cropping || create.isPending}
            disabled={body.title === '' || !area}
          >
            Simpan
          </Button>
        </div>
      }
    >
      <form id="new-slide" onSubmit={submit} noValidate>
        <SlideFields
          preview={
            <div className="flex flex-col gap-1">
              <div className="bg-tv-media @container relative aspect-[1312/582] overflow-hidden rounded-md">
                {preview && (
                  <Cropper
                    image={preview}
                    crop={crop}
                    zoom={zoom}
                    maxZoom={MAX_ZOOM}
                    aspect={SLIDE_ASPECT}
                    objectFit="cover"
                    showGrid={false}
                    onCropChange={setCrop}
                    onZoomChange={setZoom}
                    onCropComplete={(_, pixels) => setArea(pixels)}
                    // the frame is the whole box: no outline or dimmed margin
                    style={{ cropAreaStyle: { border: 'none', boxShadow: 'none' } }}
                  />
                )}
                {body.show_caption && (
                  <SlideCaptionBar
                    size="preview"
                    title={body.title}
                    caption={body.caption}
                    className="pointer-events-none z-10"
                  />
                )}
              </div>
              <Slider
                aria-label="Zoom"
                min={1}
                max={MAX_ZOOM}
                step={0.01}
                value={zoom}
                onChange={(event) => setZoom(Number(event.currentTarget.value))}
                minLabel="Kecil"
                maxLabel="Besar"
              />
            </div>
          }
          previewHint="Geser gambar dan atur zoom sampai pas. Bagian di luar kotak tidak ikut diunggah."
          value={value}
          onChange={setValue}
        />
      </form>
    </Sheet>
  )
}
