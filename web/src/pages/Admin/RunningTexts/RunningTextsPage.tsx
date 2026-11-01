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
import {
  useDeleteRunningText,
  useRunningTextSetting,
  useRunningTexts,
} from '@/hooks/useRunningTexts'
import type { RunningText } from '@/types/runningText'
import { AddTextCard } from './components/AddTextCard'
import { SpeedCard } from './components/SpeedCard'
import { TextListCard } from './components/TextListCard'
import { EditTextDrawer } from './drawers/EditTextDrawer'
import { matchesFilter, type TextFilter, TV_DELAY } from './utils'

// Running text (E6-05). Super admin and takmir; the route is guarded by RequireAccess.
export function RunningTextsPage() {
  return (
    <>
      <PageHeader title="Running Text" />
      <RunningTextsContent />
    </>
  )
}

function RunningTextsContent() {
  const texts = useRunningTexts()
  const setting = useRunningTextSetting()
  const [editing, setEditing] = useState<RunningText | null>(null)
  const [deleting, setDeleting] = useState<RunningText | null>(null)
  // Old, switched-off texts pile up; the list opens on what is on the TV now.
  const [filter, setFilter] = useState<TextFilter>('active')
  const remove = useDeleteRunningText()
  const toast = useToast()
  const failed = texts.error ?? setting.error

  if (failed) {
    return (
      <PageBody>
        <Callout
          tone="danger"
          title="Running text gagal dimuat"
          action={
            <Button
              size="md"
              variant="secondary"
              onClick={() => {
                texts.refetch()
                setting.refetch()
              }}
            >
              Coba lagi
            </Button>
          }
        >
          {failed.message}
        </Callout>
      </PageBody>
    )
  }

  if (texts.data === undefined || setting.data === undefined) {
    return (
      <PageBody>
        <Card aria-busy aria-label="Memuat running text" className="flex flex-col gap-3">
          <Skeleton className="h-8" />
          <Skeleton className="h-8" />
          <Skeleton className="h-8" />
        </Card>
      </PageBody>
    )
  }

  const confirmDelete = (text: RunningText) =>
    remove.mutate(text.id, {
      onSuccess: () => toast({ message: 'Running text dihapus', description: TV_DELAY }),
      onError: (error) => toast({ tone: 'danger', message: error.message }),
    })

  return (
    <>
      <PageBody>
        {/* a new text is active, so show it */}
        <AddTextCard onAdded={() => filter === 'inactive' && setFilter('active')} />
        <div className="flex flex-col gap-2">
          {texts.data.length > 0 && (
            <SegmentedControl
              label="Tampilkan"
              variant="pill"
              options={filterOptions(texts.data)}
              value={filter}
              onChange={setFilter}
            />
          )}
          <TextListCard key={filter} texts={texts.data} filter={filter} onEdit={setEditing} />
          {texts.data.length > 0 && (
            <p className="text-ink-muted text-xs">
              Geser ikon titik untuk urutkan, ketuk teks untuk edit atau hapus. Teks yang dimatikan
              tetap tersimpan.
            </p>
          )}
        </div>
        <SpeedCard speed={setting.data.speed} texts={texts.data} />
      </PageBody>

      {editing && (
        <EditTextDrawer
          text={editing}
          onClose={() => setEditing(null)}
          onDelete={(text) => {
            setEditing(null)
            setDeleting(text)
          }}
        />
      )}
      <ConfirmDialog
        open={deleting !== null}
        onClose={() => setDeleting(null)}
        title="Hapus running text?"
        description={deleting && `"${deleting.text}" tidak lagi tampil di layar TV.`}
        confirmLabel="Hapus"
        danger
        onConfirm={() => deleting && confirmDelete(deleting)}
      />
    </>
  )
}

function filterOptions(texts: RunningText[]) {
  const count = (filter: TextFilter) => texts.filter((t) => matchesFilter(t, filter)).length
  return [
    { value: 'active', label: `Aktif (${count('active')})` },
    { value: 'inactive', label: `Nonaktif (${count('inactive')})` },
    { value: 'all', label: `Semua (${texts.length})` },
  ] as const
}
