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
  Skeleton,
  Stat,
  StatusDot,
} from '@/components'
import { useRecentActivity, useScreens } from '@/hooks/useHomeData'
import { useMosqueProfile } from '@/hooks/useMosqueProfile'
import { useNow } from '@/hooks/useNow'
import { useDaySchedule } from '@/hooks/usePrayerTimes'
import { useProfile } from '@/hooks/useProfile'
import { ROUTES } from '@/route'
import { MENU } from '@/routes/menu'
import { useSession } from '@/stores/sessionStore'
import type { Screen } from '@/types/home'
import { canOpen } from '@/utils/auth/permissions'
import { roleLabels } from '@/utils/auth/roles'
import { formatActivityTime, formatTimeAgo } from '@/utils/format/date'
import { PRAYER_KEYS } from '@/utils/prayer/calculate'
import { localDate } from '@/utils/prayer/calendar'
import { AccountDrawer } from './drawers/AccountDrawer'
import { ChangePasswordDrawer } from './drawers/ChangePasswordDrawer'
import { formatCountdown, nextPrayer } from './nextPrayer'

// Home, "Beranda" in the UI (E3-02): the hub. Every other screen opens from the role-filtered "Kelola" grid.
export function HomePage() {
  const session = useSession()
  const mosque = useMosqueProfile()
  const [drawer, setDrawer] = useState<'account' | 'password' | null>(null)
  const now = useNow()
  const timezone = mosque.data?.timezone ?? 'Asia/Jakarta'
  const schedule = useDaySchedule(localDate(now, timezone))
  const prayers = schedule.data?.filter((e) => (PRAYER_KEYS as readonly string[]).includes(e.key))
  const next = prayers?.length ? nextPrayer(now, prayers, timezone) : null
  // no location saved yet: nothing to compute the times from
  const noLocation = schedule.data === null
  const screens = useScreens()
  const activity = useRecentActivity()
  const role = roleLabels[session.role]

  const profile = useProfile()
  const name = profile.data?.name
  const mosqueName = mosque.data?.name ?? 'Masjid Digital'

  const menu = MENU.filter((item) => canOpen(session.role, item.module))

  return (
    <>
      <HeroHeader>
        <div className="flex items-center justify-between gap-3">
          <div className="min-w-0">
            <p className="text-on-brand-muted truncate text-sm">
              Assalamu'alaikum{name && `, ${name}`} · {role}
            </p>
            <h1 className="mt-0.5 text-xl font-extrabold">{mosqueName}</h1>
          </div>
          <IconButton
            variant="brand"
            label="Akun saya"
            aria-haspopup="dialog"
            onClick={() => setDrawer('account')}
            icon={<Avatar name={name ?? role} tone="brand-2" />}
          />
        </div>
        <Card tone="brand-2" padding="md" className="flex items-center justify-between">
          <Stat
            label="Sholat berikutnya"
            value={<NextPrayerName next={next} noLocation={noLocation} />}
            size="lg"
          />
          <Stat
            label="dalam"
            value={<NextPrayerCountdown next={next} noLocation={noLocation} />}
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
            description={screens.data ? screenSummary(screens.data, now) : 'Memuat…'}
            trailing={
              <span className="flex gap-1.5">
                {screens.data?.map((s) => (
                  <StatusDot key={s.id} tone={s.status === 'online' ? 'success' : 'danger'} />
                ))}
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
              {activity.data?.map((item) => (
                <li key={item.id} className="flex gap-2.5 py-2.5 text-xs">
                  <span className="flex-1 leading-snug">
                    <b>{item.actor_name}</b> {item.summary}
                  </span>
                  <span className="text-ink-muted whitespace-nowrap">
                    {formatActivityTime(item.created_at, now)}
                  </span>
                </li>
              ))}
            </List>
          </Card>
        </section>
      </PageBody>

      {drawer === 'account' && (
        <AccountDrawer
          session={session}
          name={name}
          onChangePassword={() => setDrawer('password')}
          onClose={() => setDrawer(null)}
        />
      )}
      {drawer === 'password' && <ChangePasswordDrawer onClose={() => setDrawer(null)} />}
    </>
  )
}

function screenSummary(screens: Screen[], now: Date) {
  if (screens.length === 0) return 'Belum ada layar terpasang'
  const online = screens.filter((s) => s.status === 'online').length
  const offline = screens.filter((s) => s.status === 'offline')
  if (offline.length === 0) return `${online} online`
  const oldest = offline.reduce((a, b) => (a.last_seen_at < b.last_seen_at ? a : b))
  return `${online} online · ${offline.length} offline sejak ${formatTimeAgo(oldest.last_seen_at, now)}`
}

type NextPrayerProps = {
  next: ReturnType<typeof nextPrayer> | null
  /** No location saved yet: nothing to compute the times from. */
  noLocation: boolean
}

function NextPrayerName({ next, noLocation }: NextPrayerProps) {
  if (next) return `${next.name} ${next.time}`
  if (noLocation) {
    return (
      <span className="text-on-brand-muted text-sm font-semibold">Isi lokasi di Profil masjid</span>
    )
  }
  return <Skeleton className="bg-brand mt-1 h-6 w-32" />
}

function NextPrayerCountdown({ next, noLocation }: NextPrayerProps) {
  if (next) {
    return <time dateTime={`PT${next.minutesLeft}M`}>{formatCountdown(next.minutesLeft)}</time>
  }
  if (noLocation) return '—'
  return <Skeleton className="bg-brand mt-1 ml-auto h-6 w-14" />
}
