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
import { Button, Card, EmptyState, useToast } from '@/components'
import { useReorderRunningTexts, useUpdateRunningText } from '@/hooks/useRunningTexts'
import type { RunningText } from '@/types/runningText'
import { cn } from '@/utils/cn'
import { formatTimeAgo } from '@/utils/format/date'
import { matchesFilter, mergeOrder, type TextFilter } from '../utils'

type TextListCardProps = {
  texts: RunningText[]
  filter: TextFilter
  onEdit: (text: RunningText) => void
}

const EMPTY: Record<TextFilter, { title: string; description: string }> = {
  active: {
    title: 'Tidak ada teks aktif',
    description:
      'INFO bar di TV disembunyikan. Ketuk Tampilkan di filter Nonaktif untuk memunculkannya.',
  },
  inactive: { title: 'Tidak ada teks nonaktif', description: 'Semua teks sedang tampil di TV.' },
  all: {
    title: 'Belum ada running text',
    description:
      'Tulis teks di atas, lalu ketuk Tambah. Teks aktif berjalan di bagian bawah layar TV.',
  },
}

// Drag by the handle to reorder, Matikan/Tampilkan to take it off or put it on the TV, tap the
// text to edit or delete.
// Every change is saved right away. Mounted per filter (key), so `kept` starts empty each time.
export function TextListCard({ texts, filter, onEdit }: TextListCardProps) {
  // Texts toggled out of the filter stay until the filter changes, so a tap never makes a row vanish.
  const [kept, setKept] = useState<ReadonlySet<number>>(new Set())
  const reorder = useReorderRunningTexts()
  const toast = useToast()
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  )

  const shown = texts.filter((t) => matchesFilter(t, filter) || kept.has(t.id))
  if (shown.length === 0) {
    const empty = EMPTY[texts.length === 0 ? 'all' : filter]
    return <EmptyState title={empty.title} description={empty.description} />
  }

  const onDragEnd = ({ active, over }: DragEndEvent) => {
    if (!over || active.id === over.id) return
    const ids = shown.map((t) => t.id)
    const moved = arrayMove(ids, ids.indexOf(Number(active.id)), ids.indexOf(Number(over.id)))
    reorder.mutate(
      mergeOrder(
        texts.map((t) => t.id),
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
        accessibility={{ screenReaderInstructions: { draggable: DRAG_INSTRUCTIONS } }}
      >
        <SortableContext items={shown.map((t) => t.id)} strategy={verticalListSortingStrategy}>
          <ul className="divide-line-soft divide-y">
            {shown.map((t) => (
              <TextRow
                key={t.id}
                text={t}
                onEdit={onEdit}
                onToggled={() => setKept((ids) => new Set(ids).add(t.id))}
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

/** "Ahmad Fauzi · Sedang tampil" or "Ahmad Fauzi · Terakhir tampil 3 hari lalu". */
function meta(text: RunningText) {
  const shown = text.active
    ? 'Sedang tampil'
    : text.deactivated_at && `Terakhir tampil ${formatTimeAgo(text.deactivated_at)}`
  return [text.created_by_name, shown].filter(Boolean).join(' · ')
}

type TextRowProps = {
  text: RunningText
  onEdit: (text: RunningText) => void
  onToggled: () => void
}

function TextRow({ text, onEdit, onToggled }: TextRowProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: text.id })
  const update = useUpdateRunningText()
  const toast = useToast()

  const toggle = (active: boolean) => {
    onToggled()
    update.mutate(
      { id: text.id, text: text.text, active },
      { onError: (error) => toast({ tone: 'danger', message: error.message }) },
    )
  }

  return (
    <li
      ref={setNodeRef}
      style={{ transform: CSS.Translate.toString(transform), transition }}
      className={cn(
        'bg-card relative flex items-start gap-2 py-3',
        isDragging && 'shadow-overlay z-10 rounded-md',
      )}
    >
      <button
        ref={setActivatorNodeRef}
        type="button"
        aria-label={`Urutkan: ${text.text}`}
        className="text-ink-muted -ml-1 flex h-6 w-6 shrink-0 cursor-grab touch-none items-center justify-center"
        {...attributes}
        {...listeners}
      >
        <GripVertical aria-hidden className="size-4" />
      </button>
      <button type="button" onClick={() => onEdit(text)} className="min-w-0 flex-1 text-left">
        <span className={cn('block text-base leading-snug', !text.active && 'text-ink-muted')}>
          {text.text}
        </span>
        <span className="text-ink-muted mt-0.5 block text-xs">{meta(text)}</span>
      </button>
      {/* the action, not the state: the meta line already says whether it is on the TV */}
      <Button
        size="sm"
        variant={text.active ? 'secondary' : 'outline'}
        aria-label={`${text.active ? 'Matikan' : 'Tampilkan'}: ${text.text}`}
        onClick={() => toggle(!text.active)}
      >
        {text.active ? 'Matikan' : 'Tampilkan'}
      </Button>
    </li>
  )
}
