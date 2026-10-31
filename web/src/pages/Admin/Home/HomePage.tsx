import { useState } from 'react'
import {
  Avatar,
  Card,
  HeroHeader,
  IconButton,
  List,
  MenuTile,
  NavCard,
  PageBody,
  SectionHeader,
  Stat,
  StatusDot,
} from '@/components'
import { useMosqueProfile } from '@/hooks/useMosqueProfile'
import { useUser } from '@/hooks/useUsers'
import { ROUTES } from '@/route'
import { MENU } from '@/routes/menu'
import { useSession } from '@/stores/sessionStore'
import { canOpen } from '@/utils/auth/permissions'
import { roleLabels } from '@/utils/auth/roles'
import { AccountDrawer } from './drawers/AccountDrawer'
import { prayerTimes, recentActivity, screens } from './fixtures'
import { formatCountdown, nextPrayer } from './nextPrayer'
import { useNow } from './useNow'

// Home, "Beranda" in the UI (E3-02): the hub. Every other screen opens from the role-filtered "Kelola" grid.
export function HomePage() {
  const session = useSession()
  const mosque = useMosqueProfile()
  const [accountOpen, setAccountOpen] = useState(false)
  const now = useNow()
  const next = nextPrayer(now, prayerTimes)
  const role = roleLabels[session.role]

  // The API has no /me yet; only super admins may read a user, so others are greeted by role.
  const me = useUser(session.userId, { enabled: session.role === 'super_admin' })
  const name = me.data?.name

  const menu = MENU.filter((item) => canOpen(session.role, item.module))

  return (
    <>
      <HeroHeader>
        <div className="flex items-center justify-between gap-3">
          <div className="min-w-0">
            <p className="text-on-brand-muted truncate text-sm">
              Assalamu'alaikum{name && `, ${name}`} · {role}
            </p>
            <h1 className="mt-0.5 text-xl font-extrabold">{mosque.name}</h1>
          </div>
          <IconButton
            variant="brand"
            label="Akun saya"
            aria-haspopup="dialog"
            onClick={() => setAccountOpen(true)}
            icon={<Avatar name={name ?? role} tone="brand-2" />}
          />
        </div>
        <Card tone="brand-2" padding="md" className="flex items-center justify-between">
          <Stat label="Sholat berikutnya" value={`${next.name} ${next.time}`} size="lg" />
          <Stat
            label="dalam"
            value={
              <time dateTime={`PT${next.minutesLeft}M`}>{formatCountdown(next.minutesLeft)}</time>
            }
            size="lg"
            align="end"
            valueClassName="text-gold"
          />
        </Card>
      </HeroHeader>

      <PageBody className="gap-4.5 pt-4.5">
        {canOpen(session.role, 'screens') && (
          <NavCard
            to={ROUTES.screens}
            title="Layar TV"
            description={`${screens.online} online · ${screens.offline} offline sejak ${screens.offlineSince}`}
            trailing={
              <span className="flex gap-1.5">
                <StatusDot tone="success" />
                <StatusDot tone="danger" />
              </span>
            }
          />
        )}

        <section aria-labelledby="manage-title">
          <SectionHeader title={<span id="manage-title">Kelola</span>} className="mb-2.5" />
          <nav aria-labelledby="manage-title" className="grid grid-cols-4 gap-2">
            {menu.map((item) => (
              <MenuTile key={item.path} to={item.path} icon={<item.icon />} label={item.label} />
            ))}
          </nav>
        </section>

        <section aria-labelledby="activity-title">
          <SectionHeader
            title={<span id="activity-title">Aktivitas terakhir</span>}
            className="mb-2.5"
          />
          <Card padding="list">
            <List>
              {recentActivity.map((item) => (
                <li key={item.id} className="flex gap-2.5 py-2.5 text-xs">
                  <span className="flex-1 leading-snug">
                    <b>{item.actor}</b> {item.action}
                  </span>
                  <span className="text-ink-muted whitespace-nowrap">{item.time}</span>
                </li>
              ))}
            </List>
          </Card>
        </section>
      </PageBody>

      {accountOpen && (
        <AccountDrawer session={session} name={name} onClose={() => setAccountOpen(false)} />
      )}
    </>
  )
}
