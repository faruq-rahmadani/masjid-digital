import { useEffect } from 'react'
import { MosqueLogo, SlideFrame } from '@/components'
import { usePublicSlides } from '@/hooks/useSlides'
import { useCashDisplay, useQuotes } from '@/hooks/useTvData'
import { slideImageUrl } from '@/services/slideService'
import type { FinanceDisplay } from '@/types/finance'
import type { IqamahSettings } from '@/types/iqamah'
import type { ScheduleItem } from '@/types/tv'
import { cn } from '@/utils/cn'
import { formatRupiah } from '@/utils/format/money'
import { formatClock, formatHijriDate, formatHms, formatLongDate } from '../clock'
import { RunningText } from '../components/RunningText'
import type { DisplayState } from '../displayMode'
import { slideIndexAt } from '../slideShow'
import type { TvContext } from '../types'

const QUOTE_SECONDS = 60
/** With several kas on the TV, the header shows one at a time (PRD: static columns, never a ticker). */
const CASH_SECONDS = 6

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
        <CashCard seconds={tv.now.getTime() / 1000} />
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
              ? `Khutbah ${settings.khutbah_min} menit setelah adzan`
              : `Iqamah ${settings.iqamah_min[state.next.key]} menit setelah adzan`}
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

const PERIOD_LABEL: Record<FinanceDisplay['period'], string> = {
  today: 'Hari ini',
  week: 'Minggu ini',
  month: 'Bulan ini',
}

// Rotation follows the clock like the quotes, so it needs no timer of its own.
function CashCard({ seconds }: { seconds: number }) {
  const cash = useCashDisplay().data
  if (!cash || cash.pockets.length === 0) return <div className="bg-canvas rounded-tv flex-1" />
  const index = Math.floor(seconds / CASH_SECONDS) % cash.pockets.length
  const kas = cash.pockets[index] ?? cash.pockets[0]
  if (!kas) return null

  return (
    <div className="bg-canvas text-ink rounded-tv flex min-w-0 flex-1 items-center gap-6 px-6">
      <div className="w-60 shrink-0">
        <div className="text-tv-16 text-ink-soft whitespace-nowrap">
          Laporan Kas · {PERIOD_LABEL[cash.period]}
        </div>
        <div className="text-tv-22 mt-1 line-clamp-2 leading-tight font-bold">{kas.name}</div>
        {cash.pockets.length > 1 && (
          <div className="mt-2 flex gap-1.5">
            {cash.pockets.map((p, i) => (
              <span
                key={p.id}
                className={cn('h-2 rounded-full', i === index ? 'bg-gold w-6' : 'bg-line w-2')}
              />
            ))}
          </div>
        )}
      </div>
      <div className="bg-line my-6 w-0.5 self-stretch" />
      <div className="grid min-w-0 flex-1 grid-cols-[1fr_1fr_1fr_1.3fr] items-center gap-4">
        <CashValue label="Saldo awal" value={formatRupiah(kas.opening)} />
        <CashValue label="Pemasukan" value={`+ ${formatRupiah(kas.in)}`} className="text-income" />
        <CashValue
          label="Pengeluaran"
          value={`- ${formatRupiah(kas.out)}`}
          className="text-expense"
        />
        <div className="bg-brand text-on-brand rounded-xl px-4.5 py-3">
          <div className="text-tv-16 text-on-brand-muted">Saldo akhir</div>
          <div className="text-tv-28 mt-0.5 font-extrabold whitespace-nowrap">
            {formatRupiah(kas.closing)}
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
  const slides = usePublicSlides().data ?? []
  const index = slideIndexAt(
    slides.map((s) => s.duration_seconds),
    seconds,
  )
  const slide = slides[index]
  usePreloadSlides(slides)

  return (
    <SlideFrame
      className="min-w-0 flex-1"
      src={slide && slideImageUrl(slide.id)}
      title={slide?.title ?? ''}
      caption={slide?.caption ?? ''}
      showCaption={slide?.show_caption ?? false}
      aside={
        slides.length > 1 && (
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
        )
      }
    />
  )
}

/**
 * Fetch every slide's image as soon as the list arrives, so each one is in the browser's cache
 * (kept for good: the images are immutable) before its turn and every TV switches at the same
 * second without a blank frame on a slow connection.
 */
function usePreloadSlides(slides: readonly { id: number }[]) {
  const ids = slides.map((s) => s.id).join()
  useEffect(() => {
    for (const id of ids ? ids.split(',') : []) {
      new Image().src = slideImageUrl(Number(id))
    }
  }, [ids])
}
