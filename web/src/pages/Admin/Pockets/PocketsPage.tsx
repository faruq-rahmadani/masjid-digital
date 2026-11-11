import { Lock, Monitor, Plus } from 'lucide-react'
import { useState } from 'react'
import {
  Badge,
  Button,
  Callout,
  Card,
  PageBody,
  PageHeader,
  ProgressBar,
  SectionHeader,
  Skeleton,
} from '@/components'
import { usePockets } from '@/hooks/useFinance'
import { useSession } from '@/stores/sessionStore'
import type { Pocket } from '@/types/finance'
import { accessTo } from '@/utils/auth/permissions'
import { isBuiltIn, isOpen, POCKET_KIND_LABEL, POCKET_KIND_TONE } from '@/utils/finance'
import { formatRupiah } from '@/utils/format/money'
import { FinishPocketDrawer } from './drawers/FinishPocketDrawer'
import { PocketDrawer } from './drawers/PocketDrawer'

// Kas (E7-18, E7-19): the four built-in kas (locked) and the terikat ones pengurus add.
export function PocketsPage() {
  const pockets = usePockets()
  const session = useSession()
  const canEdit = accessTo(session.role, 'finance') === 'edit'
  const [editing, setEditing] = useState<Pocket | 'new' | null>(null)
  const [finishing, setFinishing] = useState<Pocket | null>(null)

  const active = pockets.data?.filter(isOpen) ?? []
  const archived = pockets.data?.filter((p) => !isOpen(p)) ?? []

  return (
    <>
      <PageHeader title="Kas" meta={pockets.data && `${active.length} kas aktif`} />
      <PageBody>
        {pockets.error && <Callout tone="danger">{pockets.error.message}</Callout>}
        {!pockets.data && !pockets.error && <Skeleton className="h-40" />}
        {canEdit && (
          <Button size="lg" block onClick={() => setEditing('new')}>
            <Plus aria-hidden className="size-4" /> Tambah kas
          </Button>
        )}
        {active.map((p) => (
          <PocketCard key={p.id} pocket={p} onOpen={() => setEditing(p)} />
        ))}
        {archived.length > 0 && (
          <>
            <SectionHeader title={`Diarsipkan (${archived.length})`} />
            {archived.map((p) => (
              <PocketCard key={p.id} pocket={p} onOpen={() => setEditing(p)} />
            ))}
          </>
        )}
      </PageBody>

      {editing && (
        <PocketDrawer
          pocket={editing === 'new' ? null : editing}
          canEdit={canEdit}
          onFinish={(p) => {
            setEditing(null)
            setFinishing(p)
          }}
          onClose={() => setEditing(null)}
        />
      )}
      {finishing && <FinishPocketDrawer pocket={finishing} onClose={() => setFinishing(null)} />}
    </>
  )
}

function PocketCard({ pocket, onOpen }: { pocket: Pocket; onOpen: () => void }) {
  return (
    <Card padding="lg">
      <button type="button" onClick={onOpen} className="flex w-full flex-col gap-2 text-left">
        <span className="flex items-center gap-2">
          <span className="flex-1 font-bold">{pocket.name}</span>
          <Badge tone={POCKET_KIND_TONE[pocket.kind]} size="sm">
            {POCKET_KIND_LABEL[pocket.kind]}
          </Badge>
          {isBuiltIn(pocket.kind) && (
            <Lock aria-label="Kas bawaan" className="text-ink-muted size-3.5" />
          )}
        </span>
        <span className="text-xl font-extrabold">{formatRupiah(pocket.balance)}</span>
        {pocket.purpose && <span className="text-ink-soft text-xs">{pocket.purpose}</span>}
        {pocket.progress && pocket.target_amount !== null && (
          <span className="flex flex-col gap-1">
            <ProgressBar
              label={`Target ${pocket.name}`}
              value={pocket.progress.collected}
              max={pocket.target_amount}
              tone={pocket.progress.target_reached ? 'income' : 'gold'}
            />
            <span className="text-ink-soft text-xs">
              {pocket.progress.target_reached && <b className="text-income">Target tercapai · </b>}
              {formatRupiah(pocket.progress.collected)} dari {formatRupiah(pocket.target_amount)}
            </span>
          </span>
        )}
        <span className="text-ink-soft flex items-center gap-1.5 text-xs">
          <Monitor aria-hidden className="size-3.5" />
          {pocket.show_on_tv ? 'Tampil di layar TV' : 'Tidak tampil di TV'}
          {!isOpen(pocket) && ' · diarsipkan'}
        </span>
      </button>
    </Card>
  )
}
