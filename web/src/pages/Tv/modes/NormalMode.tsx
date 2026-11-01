import { Image } from 'lucide-react'
import { MosqueLogo } from '@/components'
import { useCashSummary, useQuotes, useSlides } from '@/hooks/useTvData'
import type { IqamahSettings, ScheduleItem } from '@/types/tv'
import { cn } from '@/utils/cn'
import { formatRupiah } from '@/utils/format/money'
import { formatClock, formatHijriDate, formatHms, formatLongDate, formatShortDate } from '../clock'
import { RunningText } from '../components/RunningText'
import type { DisplayState } from '../displayMode'
import type { TvContext } from '../types'

const SLIDE_SECONDS = 10
const QUOTE_SECONDS = 60

type NormalModeProps = {
  tv: TvContext
  state: Extract<DisplayState, { mode: 'normal' }>
  schedule: readonly ScheduleItem[]
  settings: IqamahSettings
}

// Normal (E8-01, tv-main.html): header with kas, clock and quote, slide, countdown, 7 tiles, INFO.
export function NormalMode({ tv, state, schedule, settings }: NormalModeProps) {
  return (
    <div className="bg-brand text-on-brand flex min-h-0 flex-1 flex-col gap-6 p-8">
      <div className="flex h-30 shrink-0 gap-6">
        <MosqueCard mosque={tv.mosque} />
        <CashCard />
      </div>

      <div className="flex min-h-0 flex-1 gap-6">
        <div className="flex min-h-0 w-130 shrink-0 flex-col gap-6">
          <div className="pt-3 text-center">
            <div className="text-tv-32 font-bold">{formatLongDate(tv.now, tv.timeZone)}</div>
            <div className="text-tv-24 text-on-brand-muted mt-1.5">
              {formatHijriDate(tv.now, tv.timeZone)}
            </div>
            <div className="text-tv-196 mt-4 leading-none font-extrabold tracking-tighter">
              {formatClock(tv.clock.seconds)}
            </div>
            <div className="text-tv-22 text-gold mt-2 font-semibold tracking-widest">{tv.zone}</div>
          </div>
          <QuoteCard seconds={tv.now.getTime() / 1000} />
        </div>
        <SlideShow seconds={tv.now.getTime() / 1000} />
      </div>

      <div className="flex h-47.5 shrink-0 gap-4">
        <div className="bg-canvas text-ink rounded-tv mr-2 flex w-130 shrink-0 flex-col items-center justify-center">
          <div className="text-tv-24 font-bold">
            Menuju {state.jumat ? 'Sholat Jumat' : state.next.name}
          </div>
          <div className="text-tv-84 leading-none font-extrabold tracking-tight">
            {formatHms(state.secondsLeft)}
          </div>
          <div className="text-tv-18 text-ink-soft mt-1">
            {state.jumat
              ? `Khutbah ${settings.khutbah_minutes} menit setelah adzan`
              : `Iqamah ${settings.iqamah_minutes[state.next.key]} menit setelah adzan`}
          </div>
        </div>
        <ol className="grid flex-1 grid-cols-7 gap-4">
          {schedule.map((item) => {
            const next = item.key === state.next.key
            const name = next && state.jumat ? 'Jumat' : item.name
            return (
              <li
                key={item.key}
                className={cn(
                  'rounded-tv flex flex-col items-center justify-center',
                  next
                    ? 'bg-gold text-on-gold outline-canvas gap-1 outline-4 -outline-offset-4'
                    : 'bg-canvas gap-1.5',
                )}
              >
                {next && <div className="text-tv-16 font-bold tracking-widest">BERIKUTNYA</div>}
                <div
                  className={cn(
                    'text-tv-26',
                    next ? 'font-bold' : 'text-tv-tile-label font-semibold',
                  )}
                >
                  {name}
                </div>
                <div className={cn('text-tv-54 font-extrabold', !next && 'text-ink')}>
                  {item.time}
                </div>
              </li>
            )
          })}
        </ol>
      </div>

      <RunningText />
    </div>
  )
}

function MosqueCard({ mosque }: { mosque: TvContext['mosque'] }) {
  return (
    <div className="bg-canvas text-ink rounded-tv flex w-130 shrink-0 items-center gap-5 px-7">
      {mosque.logo ? (
        <img src={mosque.logo} alt="" className="size-18 shrink-0 rounded-2xl object-cover" />
      ) : (
        <MosqueLogo size={72} className="shrink-0" />
      )}
      <div className="min-w-0">
        <div className="text-tv-30 line-clamp-2 leading-tight font-extrabold">{mosque.name}</div>
        {mosque.address && (
          <div className="text-tv-18 text-ink-soft mt-1.5 truncate">{mosque.address}</div>
        )}
      </div>
    </div>
  )
}

function CashCard() {
  const cash = useCashSummary().data
  if (!cash) return <div className="bg-canvas rounded-tv flex-1" />

  return (
    <div className="bg-canvas text-ink rounded-tv flex min-w-0 flex-1 items-center gap-6 px-6">
      <div className="w-37.5 shrink-0">
        <div className="text-tv-16 text-ink-soft">Laporan Kas</div>
        <div className="text-tv-22 mt-1 font-bold">Per {formatShortDate(cash.as_of)}</div>
      </div>
      <div className="bg-line my-6 w-0.5 self-stretch" />
      <div className="grid min-w-0 flex-1 grid-cols-[1fr_1fr_1fr_1.3fr] items-center gap-4">
        <CashValue label="Saldo awal" value={formatRupiah(cash.opening_balance)} />
        <CashValue
          label="Pemasukan"
          value={`+ ${formatRupiah(cash.income)}`}
          className="text-income"
        />
        <CashValue
          label="Pengeluaran"
          value={`- ${formatRupiah(cash.expense)}`}
          className="text-expense"
        />
        <div className="bg-brand text-on-brand rounded-xl px-4.5 py-3">
          <div className="text-tv-16 text-on-brand-muted">Saldo akhir</div>
          <div className="text-tv-28 mt-0.5 font-extrabold whitespace-nowrap">
            {formatRupiah(cash.closing_balance)}
          </div>
        </div>
      </div>
    </div>
  )
}

function CashValue({
  label,
  value,
  className,
}: {
  label: string
  value: string
  className?: string
}) {
  return (
    <div className="min-w-0">
      <div className="text-tv-16 text-ink-soft">{label}</div>
      <div className={cn('text-tv-28 mt-1 truncate font-bold', className)}>{value}</div>
    </div>
  )
}

// Rotation follows the clock (index = time / interval), so it needs no timer of its own.
function QuoteCard({ seconds }: { seconds: number }) {
  const quotes = useQuotes().data ?? []
  const quote = quotes[Math.floor(seconds / QUOTE_SECONDS) % Math.max(1, quotes.length)]

  return (
    <div className="bg-brand-2 rounded-tv flex min-h-0 flex-1 flex-col justify-center gap-4 overflow-hidden px-9 py-6 text-center">
      {quote && (
        <>
          <div className="text-tv-26 line-clamp-4 leading-snug">"{quote.text}"</div>
          <div className="text-tv-20 text-gold font-semibold">{quote.source}</div>
        </>
      )}
    </div>
  )
}

function SlideShow({ seconds }: { seconds: number }) {
  const slides = useSlides().data ?? []
  const index = Math.floor(seconds / SLIDE_SECONDS) % Math.max(1, slides.length)
  const slide = slides[index]

  return (
    <div className="bg-tv-media rounded-tv-lg relative min-w-0 flex-1 overflow-hidden">
      <div className="text-on-brand-muted absolute inset-x-0 top-0 bottom-35 flex items-center justify-center">
        {slide?.image_url ? (
          <img src={slide.image_url} alt="" className="size-full object-cover" />
        ) : (
          <Image aria-hidden className="size-24" strokeWidth={1.5} />
        )}
      </div>
      {slide && (
        <div className="bg-tv-caption absolute inset-x-0 bottom-0 flex h-35 items-center justify-between gap-6 px-10">
          <div className="min-w-0">
            <div className="text-tv-34 truncate font-bold">{slide.title}</div>
            <div className="text-tv-22 text-on-brand-muted mt-1.5 truncate">{slide.caption}</div>
          </div>
          {slides.length > 1 && (
            <div className="flex shrink-0 items-center gap-2.5" aria-hidden>
              {slides.map((s, i) => (
                <div
                  key={s.id}
                  className={cn(
                    'h-2.5 rounded-full',
                    i === index ? 'bg-gold w-9' : 'bg-tv-dot w-2.5',
                  )}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
