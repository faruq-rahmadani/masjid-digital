import { useState } from 'react'
import {
  type Action,
  ActionMenu,
  Badge,
  Button,
  Callout,
  Card,
  ConfirmDialog,
  Field,
  Input,
  List,
  ListRow,
  PageBody,
  PageHeader,
  SectionHeader,
  Sheet,
  Skeleton,
  Stack,
  Tabs,
  useToast,
} from '@/components'
import { useCategories, useFinanceMutation } from '@/hooks/useFinance'
import {
  archiveCategory,
  createCategory,
  deleteCategory,
  renameCategory,
  restoreCategory,
} from '@/services/financeService'
import type { Category, CategoryScope, CategoryType } from '@/types/finance'
import { canArchiveCategory, isPindah } from '@/utils/finance'

const SCOPES: { scope: CategoryScope; title: string }[] = [
  { scope: 'umum_terikat', title: 'Kas Umum & kas terikat' },
  { scope: 'zakat', title: 'Zakat' },
  { scope: 'fidyah', title: 'Fidyah' },
]

type Pending =
  | { kind: 'delete' | 'archive' | 'restore'; category: Category }
  | { kind: 'create'; name: string; type: CategoryType }

// Kategori (E7-08): per tab (pemasukan, pengeluaran) grouped by kas kind. System categories carry
// the syariah rules: asnaf may be hidden, never renamed or deleted. Own ones are for Kas Umum and
// terikat kas; a used one is archived instead of deleted.
export function CategoriesPage() {
  const categories = useCategories()
  const [tab, setTab] = useState<CategoryType>('in')
  const [renaming, setRenaming] = useState<Category | null>(null)
  const [pending, setPending] = useState<Pending | null>(null)
  const toast = useToast()
  const create = useFinanceMutation(createCategory)
  const remove = useFinanceMutation(deleteCategory)
  const archive = useFinanceMutation(archiveCategory)
  const restore = useFinanceMutation(restoreCategory)

  const report = (message: string) => ({
    onSuccess: () => toast({ message }),
    onError: (e: Error) => toast({ tone: 'danger', message: e.message }),
  })
  const run = () => {
    if (!pending) return
    setPending(null)
    if (pending.kind === 'create') {
      return create.mutate(
        { name: pending.name, type: pending.type },
        report('Kategori ditambahkan'),
      )
    }
    const { category } = pending
    if (pending.kind === 'delete') return remove.mutate(category.id, report('Kategori dihapus'))
    if (pending.kind === 'archive')
      return archive.mutate(category.id, report('Kategori diarsipkan'))
    restore.mutate(category.id, report('Kategori dipulihkan'))
  }

  const content = (type: CategoryType) => {
    const list = (categories.data ?? []).filter((c) => c.type === type && !isPindah(c))
    const archived = list.filter((c) => c.archived_at !== null)
    return (
      <Stack gap="md" className="pt-3.5">
        <AddCategory type={type} onAdd={(name) => setPending({ kind: 'create', name, type })} />
        {SCOPES.map(({ scope, title }) => {
          const rows = list.filter((c) => c.scope === scope && c.archived_at === null)
          if (rows.length === 0) return null
          return (
            <section key={scope}>
              <SectionHeader title={title} className="mb-2" />
              <Card padding="list">
                <List>
                  {rows.map((c) => (
                    <CategoryRow
                      key={c.id}
                      category={c}
                      actions={actionsFor(c, setRenaming, (kind) =>
                        setPending({ kind, category: c }),
                      )}
                    />
                  ))}
                </List>
              </Card>
            </section>
          )
        })}
        {archived.length > 0 && (
          <section>
            <SectionHeader title={`Diarsipkan (${archived.length})`} className="mb-2" />
            <Card padding="list">
              <List>
                {archived.map((c) => (
                  <CategoryRow
                    key={c.id}
                    category={c}
                    actions={[
                      {
                        label: 'Pulihkan',
                        onSelect: () => setPending({ kind: 'restore', category: c }),
                      },
                    ]}
                  />
                ))}
              </List>
            </Card>
          </section>
        )}
      </Stack>
    )
  }

  return (
    <>
      <PageHeader title="Kategori" />
      <PageBody>
        {categories.error && <Callout tone="danger">{categories.error.message}</Callout>}
        {!categories.data && !categories.error && <Skeleton className="h-40" />}
        {categories.data && (
          <Tabs
            label="Jenis kategori"
            value={tab}
            onChange={setTab}
            items={[
              { value: 'in', label: 'Pemasukan', content: content('in') },
              { value: 'out', label: 'Pengeluaran', content: content('out') },
            ]}
          />
        )}
      </PageBody>

      {renaming && <RenameSheet category={renaming} onClose={() => setRenaming(null)} />}
      <ConfirmDialog
        open={pending !== null}
        onClose={() => setPending(null)}
        title={confirmTitle(pending)}
        confirmLabel={confirmLabel(pending)}
        danger={pending?.kind === 'delete' || pending?.kind === 'archive'}
        onConfirm={run}
      />
    </>
  )
}

function actionsFor(
  c: Category,
  rename: (c: Category) => void,
  ask: (kind: 'delete' | 'archive') => void,
): Action[] {
  const own = c.code === null
  const actions: Action[] = []
  if (own) actions.push({ label: 'Ganti nama', onSelect: () => rename(c) })
  if (own && c.transaction_count === 0)
    actions.push({ label: 'Hapus', danger: true, onSelect: () => ask('delete') })
  if (canArchiveCategory(c) && (c.transaction_count > 0 || !own)) {
    actions.push({ label: 'Arsipkan', danger: true, onSelect: () => ask('archive') })
  }
  return actions
}

function CategoryRow({ category, actions }: { category: Category; actions: Action[] }) {
  const parts = [`${category.transaction_count} transaksi`, category.archived_at && 'diarsipkan']
  return (
    <ListRow
      title={
        <span className="flex items-center gap-2">
          {category.name}
          {category.code && (
            <Badge tone="neutral" size="sm">
              Sistem
            </Badge>
          )}
        </span>
      }
      description={parts.filter(Boolean).join(' · ')}
      trailing={actions.length > 0 && <ActionMenu label={category.name} actions={actions} />}
    />
  )
}

function AddCategory({ type, onAdd }: { type: CategoryType; onAdd: (name: string) => void }) {
  const [name, setName] = useState('')
  return (
    <form
      className="flex items-end gap-2"
      onSubmit={(e) => {
        e.preventDefault()
        if (!name.trim()) return
        onAdd(name.trim())
        setName('')
      }}
    >
      <Field
        label={`Kategori ${type === 'in' ? 'pemasukan' : 'pengeluaran'} baru`}
        className="flex-1"
      >
        <Input
          placeholder={type === 'in' ? 'Infaq Idul Adha' : 'Perbaikan sound'}
          maxLength={60}
          value={name}
          onChange={(e) => setName(e.currentTarget.value)}
        />
      </Field>
      <Button type="submit" size="lg">
        Tambah
      </Button>
    </form>
  )
}

function RenameSheet({ category, onClose }: { category: Category; onClose: () => void }) {
  const [name, setName] = useState(category.name)
  const [confirming, setConfirming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const rename = useFinanceMutation((next: string) => renameCategory(category.id, next))
  const toast = useToast()
  return (
    <Sheet
      open
      onClose={onClose}
      title="Ganti nama kategori"
      footer={
        <Button
          size="xl"
          block
          loading={rename.isPending}
          onClick={() => name.trim() && setConfirming(true)}
        >
          Simpan
        </Button>
      }
    >
      <Stack gap="sm">
        <Field label="Nama kategori" hint="Transaksi lama ikut memakai nama baru">
          <Input maxLength={60} value={name} onChange={(e) => setName(e.currentTarget.value)} />
        </Field>
        {error && <Callout tone="danger">{error}</Callout>}
      </Stack>
      <ConfirmDialog
        open={confirming}
        onClose={() => setConfirming(false)}
        title="Simpan nama kategori?"
        description={`"${category.name}" menjadi "${name.trim()}".`}
        confirmLabel="Simpan"
        onConfirm={() => {
          setConfirming(false)
          rename.mutate(name.trim(), {
            onSuccess: () => {
              toast({ message: 'Nama kategori disimpan' })
              onClose()
            },
            onError: (e) => setError(e.message),
          })
        }}
      />
    </Sheet>
  )
}

function confirmTitle(pending: Pending | null) {
  if (!pending) return ''
  if (pending.kind === 'create') return `Tambah kategori "${pending.name}"?`
  if (pending.kind === 'delete') return `Hapus kategori "${pending.category.name}"?`
  if (pending.kind === 'archive') return `Arsipkan "${pending.category.name}"?`
  return `Pulihkan "${pending.category.name}"?`
}

function confirmLabel(pending: Pending | null) {
  if (pending?.kind === 'delete') return 'Hapus'
  if (pending?.kind === 'archive') return 'Arsipkan'
  if (pending?.kind === 'restore') return 'Pulihkan'
  return 'Tambah'
}
