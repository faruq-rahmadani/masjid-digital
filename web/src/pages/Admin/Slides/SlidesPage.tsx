import { useState } from 'react'
import {
  Button,
  Callout,
  Card,
  ConfirmDialog,
  PageBody,
  PageHeader,
  SegmentedControl,
  Skeleton,
  useToast,
} from '@/components'
import { useDeleteSlide, useSlides } from '@/hooks/useSlides'
import type { Slide } from '@/types/slide'
import { type ActiveFilter, matchesFilter } from '@/utils/listOrder'
import { SlideListCard } from './components/SlideListCard'
import { UploadCard } from './components/UploadCard'
import { EditSlideDrawer } from './drawers/EditSlideDrawer'
import { TV_DELAY } from './utils'

// Slide (E6-04). Super admin and takmir; the route is guarded by RequireAccess. Every TV shows
// the same slides.
export function SlidesPage() {
  const slides = useSlides()
  const active = slides.data?.filter((s) => s.active).length

  return (
    <>
      <PageHeader
        title="Slide"
        meta={slides.data && `${active} aktif dari ${slides.data.length}`}
      />
      <SlidesContent />
    </>
  )
}

function SlidesContent() {
  const slides = useSlides()
  const [editing, setEditing] = useState<Slide | null>(null)
  const [deleting, setDeleting] = useState<Slide | null>(null)
  // Old, switched-off posters pile up; the list opens on what is on the TV now.
  const [filter, setFilter] = useState<ActiveFilter>('active')
  const remove = useDeleteSlide()
  const toast = useToast()

  if (slides.error) {
    return (
      <PageBody>
        <Callout
          tone="danger"
          title="Slide gagal dimuat"
          action={
            <Button size="md" variant="secondary" onClick={() => slides.refetch()}>
              Coba lagi
            </Button>
          }
        >
          {slides.error.message}
        </Callout>
      </PageBody>
    )
  }

  if (slides.data === undefined) {
    return (
      <PageBody>
        <Card aria-busy aria-label="Memuat slide" className="flex flex-col gap-3">
          <Skeleton className="h-10" />
          <Skeleton className="h-10" />
          <Skeleton className="h-10" />
        </Card>
      </PageBody>
    )
  }

  const confirmDelete = (slide: Slide) =>
    remove.mutate(slide.id, {
      onSuccess: () => toast({ message: 'Slide dihapus', description: TV_DELAY }),
      onError: (error) => toast({ tone: 'danger', message: error.message }),
    })

  return (
    <>
      <PageBody>
        {/* new slides are active, so show them */}
        <UploadCard onUploaded={() => filter === 'inactive' && setFilter('active')} />
        <div className="flex flex-col gap-2">
          {slides.data.length > 0 && (
            <SegmentedControl
              label="Tampilkan"
              variant="pill"
              options={filterOptions(slides.data)}
              value={filter}
              onChange={setFilter}
            />
          )}
          <SlideListCard key={filter} slides={slides.data} filter={filter} onEdit={setEditing} />
          {slides.data.length > 0 && (
            <p className="text-ink-muted text-xs">
              Geser ikon titik untuk urutkan, ketuk slide untuk edit atau hapus. Slide yang
              dimatikan tetap tersimpan.
            </p>
          )}
        </div>
      </PageBody>

      {editing && (
        <EditSlideDrawer
          slide={editing}
          onClose={() => setEditing(null)}
          onDelete={(slide) => {
            setEditing(null)
            setDeleting(slide)
          }}
        />
      )}
      <ConfirmDialog
        open={deleting !== null}
        onClose={() => setDeleting(null)}
        title="Hapus slide?"
        description={deleting && `"${deleting.title}" dan gambarnya dihapus dari semua layar TV.`}
        confirmLabel="Hapus"
        danger
        onConfirm={() => deleting && confirmDelete(deleting)}
      />
    </>
  )
}

function filterOptions(slides: Slide[]) {
  const count = (filter: ActiveFilter) => slides.filter((s) => matchesFilter(s, filter)).length
  return [
    { value: 'active', label: `Aktif (${count('active')})` },
    { value: 'inactive', label: `Nonaktif (${count('inactive')})` },
    { value: 'all', label: `Semua (${slides.length})` },
  ] as const
}
