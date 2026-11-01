import { UserPlus } from 'lucide-react'
import { useState } from 'react'
import {
  type Action,
  ActionBar,
  Button,
  Callout,
  Card,
  EmptyState,
  List,
  PageBody,
  PageHeader,
  SearchInput,
  SegmentedControl,
  Skeleton,
  useToast,
} from '@/components'
import { useUpdateUser, useUsers } from '@/hooks/useUsers'
import { useSession } from '@/stores/sessionStore'
import type { User } from '@/types/user'
import { ROLES, roleLabels } from '@/utils/auth/roles'
import { UserRow } from './components/UserRow'
import { AddUserDrawer } from './drawers/AddUserDrawer'
import { ChangeRoleDrawer } from './drawers/ChangeRoleDrawer'
import { EditUserDrawer } from './drawers/EditUserDrawer'
import { ResetPasswordDrawer } from './drawers/ResetPasswordDrawer'
import { DeactivateUserModal } from './modals/DeactivateUserModal'
import { filterUsers, isLastActiveSuperAdmin, sortUsers, type UserFilter } from './utils'

type Dialog =
  | { kind: 'add' }
  | { kind: 'edit' | 'role' | 'password' | 'deactivate'; user: User }
  | null

const ROLE_FILTERS = [
  { value: 'all', label: 'Semua' },
  ...ROLES.map((role) => ({ value: role, label: roleLabels[role] })),
] as const

const NO_FILTER: UserFilter = { query: '', role: 'all' }

// Pengguna (E2-07). Super admin only; the route is guarded by RequireAccess.
export function UsersPage() {
  const session = useSession()
  const users = useUsers()
  const update = useUpdateUser()
  const toast = useToast()
  const [dialog, setDialog] = useState<Dialog>(null)
  const [filter, setFilter] = useState<UserFilter>(NO_FILTER)
  const close = () => setDialog(null)

  const all = users.data ?? []
  const shown = filterUsers(sortUsers(all, session.userId), filter)
  const filtered = filter.query.trim() !== '' || filter.role !== 'all'

  const setStatus = (user: User, status: User['status']) =>
    update.mutate(
      { user, changes: { status } },
      {
        onSuccess: () =>
          toast(
            status === 'active'
              ? { message: `${user.name} diaktifkan kembali` }
              : { message: `${user.name} dinonaktifkan`, description: 'Sesi login-nya diakhiri.' },
          ),
        onError: (error) => toast({ tone: 'danger', message: error.message }),
      },
    )

  const actionsFor = (user: User): Action[] => {
    const locked = isLastActiveSuperAdmin(user, all)
    if (user.status === 'disabled') {
      return [
        { label: 'Aktifkan kembali', onSelect: () => setStatus(user, 'active') },
        { label: 'Ubah data', onSelect: () => setDialog({ kind: 'edit', user }) },
        { label: 'Atur ulang password', onSelect: () => setDialog({ kind: 'password', user }) },
      ]
    }
    return [
      { label: 'Ubah data', onSelect: () => setDialog({ kind: 'edit', user }) },
      { label: 'Ubah peran', onSelect: () => setDialog({ kind: 'role', user }), disabled: locked },
      { label: 'Atur ulang password', onSelect: () => setDialog({ kind: 'password', user }) },
      {
        label: 'Nonaktifkan akun',
        danger: true,
        disabled: locked,
        onSelect: () => setDialog({ kind: 'deactivate', user }),
      },
    ]
  }

  return (
    <>
      <PageHeader
        title="Pengguna"
        actions={
          users.data && (
            <span className="text-ink-muted text-xs font-semibold">{all.length} akun</span>
          )
        }
      />

      {/* Stays in view while the list scrolls under it. */}
      <div className="bg-canvas/95 sticky top-0 z-10 flex flex-col gap-2.5 px-5 pt-1 pb-3 backdrop-blur-sm">
        <SearchInput
          label="Cari pengguna"
          placeholder="Cari nama"
          value={filter.query}
          onClear={() => setFilter((f) => ({ ...f, query: '' }))}
          onChange={(event) => {
            // Read now: React clears currentTarget before the state updater runs.
            const query = event.currentTarget.value
            setFilter((f) => ({ ...f, query }))
          }}
        />
        <SegmentedControl
          label="Filter peran"
          variant="pill"
          value={filter.role}
          onChange={(role) => setFilter((f) => ({ ...f, role }))}
          options={ROLE_FILTERS}
          // Bleed to the screen edges so the row visibly scrolls.
          className="-mx-5 [scrollbar-width:none] flex-nowrap overflow-x-auto px-5"
        />
      </div>

      <PageBody className="pt-1">
        {users.isPending ? (
          <Card padding="list" aria-busy aria-label="Memuat pengguna">
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                className="border-line-soft flex items-center gap-3 border-b py-3 last:border-0"
              >
                <Skeleton className="size-10 rounded-full" />
                <div className="flex flex-1 flex-col gap-1.5">
                  <Skeleton className="h-3.5 w-2/5" />
                  <Skeleton className="h-3 w-3/5" />
                </div>
              </div>
            ))}
          </Card>
        ) : users.isError ? (
          <Callout
            tone="danger"
            title="Daftar pengguna gagal dimuat"
            action={
              <Button size="md" variant="secondary" onClick={() => users.refetch()}>
                Coba lagi
              </Button>
            }
          >
            {users.error.message}
          </Callout>
        ) : shown.length === 0 ? (
          <EmptyState
            title={filtered ? 'Tidak ada pengguna yang cocok' : 'Belum ada pengguna'}
            description={filtered ? 'Ubah kata kunci atau pilih peran lain.' : undefined}
            action={
              filtered && (
                <Button size="md" variant="secondary" onClick={() => setFilter(NO_FILTER)}>
                  Hapus filter
                </Button>
              )
            }
          />
        ) : (
          <>
            {filtered && (
              <p className="text-ink-muted text-xs" aria-live="polite">
                {shown.length} dari {all.length} pengguna
              </p>
            )}
            <Card padding="list">
              <List aria-label="Daftar pengguna">
                {shown.map((user) => (
                  <UserRow
                    key={user.id}
                    user={user}
                    isSelf={user.id === session.userId}
                    actions={actionsFor(user)}
                  />
                ))}
              </List>
            </Card>
            <p className="text-ink-muted text-xs">Minimal harus ada satu super admin aktif.</p>
          </>
        )}
      </PageBody>

      <ActionBar>
        <Button size="xl" block onClick={() => setDialog({ kind: 'add' })}>
          <UserPlus aria-hidden /> Tambah pengguna
        </Button>
      </ActionBar>

      {dialog?.kind === 'add' && <AddUserDrawer onClose={close} />}
      {dialog?.kind === 'edit' && <EditUserDrawer user={dialog.user} onClose={close} />}
      {dialog?.kind === 'role' && <ChangeRoleDrawer user={dialog.user} onClose={close} />}
      {dialog?.kind === 'password' && <ResetPasswordDrawer user={dialog.user} onClose={close} />}
      <DeactivateUserModal
        user={dialog?.kind === 'deactivate' ? dialog.user : null}
        onClose={close}
        onConfirm={(user) => setStatus(user, 'disabled')}
      />
    </>
  )
}
