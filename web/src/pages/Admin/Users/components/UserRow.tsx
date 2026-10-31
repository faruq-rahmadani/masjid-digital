import { type Action, ActionMenu, Avatar, Badge, ListRow } from '@/components'
import type { User } from '@/types/user'
import { roleLabels, roleTones } from '@/utils/auth/roles'
import { formatPhone } from '@/utils/format/phone'

type UserRowProps = {
  user: User
  isSelf: boolean
  actions: Action[]
}

export function UserRow({ user, isSelf, actions }: UserRowProps) {
  const disabled = user.status === 'disabled'
  const contact = user.email || formatPhone(user.phone)

  return (
    <ListRow
      leading={<Avatar name={user.name} tone={disabled ? 'neutral' : roleTones[user.role]} />}
      title={
        <span className={disabled ? 'text-ink-muted' : 'font-bold'}>
          {user.name}
          {isSelf && ' (Anda)'}
          <span className="text-ink-muted ml-1.5 text-xs font-medium whitespace-nowrap tabular-nums">
            ID {user.id}
          </span>
        </span>
      }
      description={contact}
      trailing={
        <>
          {disabled ? (
            <Badge>Nonaktif</Badge>
          ) : (
            <Badge tone={roleTones[user.role]}>{roleLabels[user.role]}</Badge>
          )}
          {/* No menu on your own row: the API forbids changing your own role. */}
          {!isSelf && <ActionMenu label={user.name} actions={actions} />}
        </>
      }
      className="py-3"
    />
  )
}
