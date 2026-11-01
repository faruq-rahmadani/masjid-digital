import { useState } from 'react'
import {
  Button,
  Callout,
  Card,
  PageBody,
  PageHeader,
  SectionHeader,
  Skeleton,
  useToast,
} from '@/components'
import { useDeleteFridayDuty, useFridayDuties } from '@/hooks/useFridayDuties'
import { useMosqueProfile } from '@/hooks/useMosqueProfile'
import { useNow } from '@/hooks/useNow'
import type { FridayDuty } from '@/types/fridayDuty'
import { localDate } from '@/utils/prayer/calendar'
import { DutyList } from './components/DutyList'
import { NextFridayCard } from './components/NextFridayCard'
import { FridayDutyDrawer } from './drawers/FridayDutyDrawer'
import { DeleteDutyModal } from './modals/DeleteDutyModal'
import { nextFridays } from './utils'

/** The coming Friday plus the following weeks shown below it. */
const FRIDAYS = 9

// Petugas Jumat (E5-05). Super admin and takmir; the route is guarded by RequireAccess. Every
// change is saved on its own right away, after a confirmation, so there is no Simpan bar.
export function FridayDutiesPage() {
  const now = useNow()
  const mosque = useMosqueProfile()
  const today = localDate(now, mosque.data?.timezone ?? 'Asia/Jakarta')
  const fridays = nextFridays(today, FRIDAYS)

  return (
    <>
      <PageHeader title="Petugas Jumat" meta="Langsung tersimpan dan tampil di layar Jumat" />
      <FridayDutiesContent fridays={fridays} />
    </>
  )
}

function FridayDutiesContent({ fridays }: { fridays: string[] }) {
  const [editingDate, setEditingDate] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<FridayDuty | null>(null)
  const duties = useFridayDuties(fridays[0] ?? '', fridays.at(-1) ?? '')
  const remove = useDeleteFridayDuty()
  const toast = useToast()
  const [next, ...later] = fridays

  if (duties.error) {
    return (
      <PageBody>
        <Callout
          tone="danger"
          title="Petugas Jumat gagal dimuat"
          action={
            <Button size="md" variant="secondary" onClick={() => duties.refetch()}>
              Coba lagi
            </Button>
          }
        >
          {duties.error.message}
        </Callout>
      </PageBody>
    )
  }

  if (!duties.data || !next) {
    return (
      <PageBody>
        <Card aria-busy aria-label="Memuat petugas Jumat" className="flex flex-col gap-3">
          <Skeleton className="h-8" />
          <Skeleton className="h-8" />
          <Skeleton className="h-8" />
        </Card>
      </PageBody>
    )
  }

  const deleteDuty = (duty: FridayDuty) =>
    remove.mutate(duty.date, {
      onSuccess: () => toast({ message: 'Petugas Jumat dihapus' }),
      onError: (error) => toast({ tone: 'danger', message: error.message }),
    })

  const empty = later.filter((date) => !duties.data.some((d) => d.date === date)).length

  return (
    <>
      <PageBody>
        <NextFridayCard
          date={next}
          duty={duties.data.find((d) => d.date === next)}
          onEdit={() => setEditingDate(next)}
        />
        <SectionHeader
          title="Jumat berikutnya"
          description={empty > 0 ? `${empty} Jumat belum diisi` : 'Semua sudah diisi'}
          className="mt-2"
        />
        <DutyList
          fridays={later}
          duties={duties.data}
          onEdit={setEditingDate}
          onDelete={setDeleting}
        />
      </PageBody>

      {editingDate && (
        <FridayDutyDrawer
          date={editingDate}
          duty={duties.data.find((d) => d.date === editingDate)}
          onClose={() => setEditingDate(null)}
        />
      )}
      <DeleteDutyModal duty={deleting} onClose={() => setDeleting(null)} onConfirm={deleteDuty} />
    </>
  )
}
