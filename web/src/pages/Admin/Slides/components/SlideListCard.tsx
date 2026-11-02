import {
  closestCenter,
  DndContext,
  type DragEndEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical } from 'lucide-react'
import { useState } from 'react'
import { Button, Card, EmptyState, Thumbnail, useToast } from '@/components'
import { useReorderSlides, useUpdateSlide } from '@/hooks/useSlides'
import { slideImageUrl } from '@/services/slideService'
import type { Slide } from '@/types/slide'
import { cn } from '@/utils/cn'
import { type ActiveFilter, matchesFilter, mergeOrder } from '@/utils/listOrder'

type SlideListCardProps = {
  slides: Slide[]
  filter: ActiveFilter
  onEdit: (slide: Slide) => void
}

const EMPTY: Record<ActiveFilter, { title: string; description: string }> = {
  active: {
    title: 'Tidak ada slide aktif',
    description:
      'Panel slide di TV kosong. Ketuk Tampilkan di filter Nonaktif untuk memunculkannya.',
  },
  inactive: {
    title: 'Tidak ada slide nonaktif',
    description: 'Semua slide sedang tampil di TV.',
  },
  all: {
    title: 'Belum ada slide',
    description: 'Upload gambar di atas. Slide aktif bergantian di panel tengah layar TV.',
  },
}

// Same pattern as the running text list: drag by the handle to reorder, Matikan/Tampilkan to take
// it off or put it on the TV, tap the slide to edit or delete. Every change is saved right away.
// Mounted per filter (key), so `kept` starts empty each time.
export function SlideListCard({ slides, filter, onEdit }: SlideListCardProps) {
  // Slides toggled out of the filter stay until the filter changes, so a tap never makes a row vanish.
  const [kept, setKept] = useState<ReadonlySet<number>>(new Set())
  const reorder = useReorderSlides()
  const toast = useToast()
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  )

  const shown = slides.filter((s) => matchesFilter(s, filter) || kept.has(s.id))
  if (shown.length === 0) {
    const empty = EMPTY[slides.length === 0 ? 'all' : filter]
    return <EmptyState title={empty.title} description={empty.description} />
  }

  const onDragEnd = ({ active, over }: DragEndEvent) => {
    if (!over || active.id === over.id) return
    const ids = shown.map((s) => s.id)
    const moved = arrayMove(ids, ids.indexOf(Number(active.id)), ids.indexOf(Number(over.id)))
    reorder.mutate(
      mergeOrder(
        slides.map((s) => s.id),
        moved,
      ),
      { onError: (error) => toast({ tone: 'danger', message: error.message }) },
    )
  }

  return (
    <Card padding="list">
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={onDragEnd}
        accessibility={{
          screenReaderInstructions: { draggable: DRAG_INSTRUCTIONS },
        }}
      >
        <SortableContext items={shown.map((s) => s.id)} strategy={verticalListSortingStrategy}>
          <ul className="divide-line-soft divide-y">
            {shown.map((s) => (
              <SlideRow
                key={s.id}
                slide={s}
                onEdit={onEdit}
                onToggled={() => setKept((ids) => new Set(ids).add(s.id))}
              />
            ))}
          </ul>
        </SortableContext>
      </DndContext>
    </Card>
  )
}

const DRAG_INSTRUCTIONS =
  'Tekan spasi atau enter untuk mengangkat, panah atas atau bawah untuk memindahkan, spasi atau enter lagi untuk meletakkan, escape untuk batal.'

/** "10 dtk · Sedang tampil" or "10 dtk · Nonaktif". */
const meta = (slide: Slide) =>
  `${slide.duration_seconds} dtk · ${slide.active ? 'Sedang tampil' : 'Nonaktif'}`

type SlideRowProps = {
  slide: Slide
  onEdit: (slide: Slide) => void
  onToggled: () => void
}

function SlideRow({ slide, onEdit, onToggled }: SlideRowProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: slide.id })
  const update = useUpdateSlide()
  const toast = useToast()

  const toggle = (active: boolean) => {
    onToggled()
    update.mutate(
      {
        id: slide.id,
        title: slide.title,
        caption: slide.caption,
        duration_seconds: slide.duration_seconds,
        show_caption: slide.show_caption,
        active,
      },
      { onError: (error) => toast({ tone: 'danger', message: error.message }) },
    )
  }

  return (
    <li
      ref={setNodeRef}
      style={{ transform: CSS.Translate.toString(transform), transition }}
      className={cn(
        'bg-card relative flex items-center gap-2 py-3',
        isDragging && 'shadow-overlay z-10 rounded-md',
      )}
    >
      <button
        ref={setActivatorNodeRef}
        type="button"
        aria-label={`Urutkan: ${slide.title}`}
        className="text-ink-muted -ml-1 flex h-6 w-6 shrink-0 cursor-grab touch-none items-center justify-center"
        {...attributes}
        {...listeners}
      >
        <GripVertical aria-hidden className="size-4" />
      </button>
      <button
        type="button"
        onClick={() => onEdit(slide)}
        className="flex min-w-0 flex-1 items-center gap-3 text-left"
      >
        <SlideThumbnail slide={slide} />
        <span className="min-w-0">
          <span
            className={cn(
              'block truncate text-base leading-snug font-semibold',
              !slide.active && 'text-ink-muted',
            )}
          >
            {slide.title}
          </span>
          <span className="text-ink-muted mt-0.5 block text-xs">{meta(slide)}</span>
        </span>
      </button>
      {/* the action, not the state: the meta line already says whether it is on the TV */}
      <Button
        size="sm"
        variant={slide.active ? 'secondary' : 'outline'}
        aria-label={`${slide.active ? 'Matikan' : 'Tampilkan'}: ${slide.title}`}
        onClick={() => toggle(!slide.active)}
      >
        {slide.active ? 'Matikan' : 'Tampilkan'}
      </Button>
    </li>
  )
}

/**
 * The slide's image with, when its title shows on the TV, a miniature of the caption bar: a dark
 * strip with a white line standing in for the text (real text would be unreadable at 72×40).
 */
function SlideThumbnail({ slide }: { slide: Slide }) {
  return (
    <span className="relative shrink-0" data-caption={slide.show_caption || undefined}>
      <Thumbnail src={slideImageUrl(slide.id)} muted={!slide.active} className="block" />
      {slide.show_caption && (
        <span
          aria-hidden
          className={cn(
            'bg-tv-caption/70 absolute inset-x-0 bottom-0 h-[32%] rounded-b-sm border-t border-white/20',
            !slide.active && 'opacity-50',
          )}
        >
          <span className="absolute top-1/2 left-1.5 h-[3px] w-1/2 -translate-y-1/2 rounded-full bg-white/85" />
        </span>
      )}
    </span>
  )
}
