import {
  AlignLeft,
  Bell,
  Check,
  ChevronLeft,
  ChevronRight,
  Clock,
  Copy,
  CreditCard,
  EllipsisVertical,
  House,
  Image,
  Minus,
  Monitor,
  Plus,
  Search,
  Settings,
  Upload,
  Users,
  X,
} from 'lucide-react'
import { useState } from 'react'
import { Group, Specimen } from '@/pages/Kit/Specimen'
import { cn } from '@/utils/cn'

const COLORS = [
  {
    group: 'Grounds',
    tokens: [
      ['canvas', 'Admin page background'],
      ['card', 'Cards, inputs'],
      ['track', 'Segmented-control track'],
      ['highlight', 'Row being edited'],
      ['media', 'Image placeholder'],
    ],
  },
  {
    group: 'Text',
    tokens: [
      ['ink', 'Primary text'],
      ['ink-soft', 'Field labels'],
      ['ink-muted', 'Secondary text, meta'],
    ],
  },
  {
    group: 'Lines',
    tokens: [
      ['line', 'Input and button borders'],
      ['line-soft', 'Row dividers'],
      ['dash', 'Dashed upload borders'],
    ],
  },
  {
    group: 'Brand',
    tokens: [
      ['brand', 'Primary actions, headers'],
      ['brand-2', 'Cards on brand'],
      ['on-brand', 'Text on brand'],
      ['on-brand-muted', 'Secondary text on brand'],
    ],
  },
  {
    group: 'Accent',
    tokens: [
      ['gold', 'Next prayer, focus ring'],
      ['on-gold', 'Text on gold'],
      ['gold-soft', 'Gold badge ground'],
      ['gold-ink', 'Gold badge text'],
      ['notice', 'Notice callout ground'],
      ['notice-ink', 'Notice callout text'],
    ],
  },
  {
    group: 'Money & status',
    tokens: [
      ['income', 'Pemasukan, online'],
      ['income-soft', 'Success ground'],
      ['expense', 'Pengeluaran, offline'],
      ['expense-soft', 'Danger ground'],
      ['expense-line', 'Destructive borders'],
      ['expense-ink', 'Danger text'],
    ],
  },
] as const

const TYPE = [
  ['Amount', 'text-4xl font-extrabold', '36 / 800', 'Rp 750.000'],
  ['Display', 'text-3xl font-extrabold', '30 / 800', 'Rp 92.058.114'],
  ['Stat', 'text-2xl font-extrabold', '22 / 800', 'Maghrib 18:15'],
  ['Page title', 'text-xl font-extrabold', '20 / 800', 'Keuangan'],
  ['Primary action', 'text-lg font-bold', '16 / 700', 'Simpan dan kirim ke layar'],
  ['Section title', 'text-md font-bold', '15 / 700', 'Aktivitas terakhir'],
  ['Body strong', 'text-base font-semibold', '14 / 600', 'Kotak amal ba’da Dzuhur'],
  ['Body', 'text-base', '14 / 400', 'Layar tetap menampilkan jadwal dari cache.'],
  ['Small', 'text-sm font-semibold', '13 / 600', 'Ganti logo'],
  ['Label', 'text-xs font-semibold text-ink-soft', '12 / 600', 'Email atau nomor HP'],
  ['Meta', 'text-xs text-ink-muted', '12 / 400', '28 Okt · Infaq transfer'],
  [
    'Caption caps',
    'text-2xs font-bold uppercase tracking-wide text-ink-muted',
    '11 / 700',
    'Koreksi menit',
  ],
] as const

const SPACING = [
  ['1', 4, 'Icon to text in tight spots'],
  ['1.5', 6, 'Label to control'],
  ['2', 8, 'Chips, menu tiles'],
  ['2.5', 10, 'Buttons side by side'],
  ['3', 12, 'Fields inside a card'],
  ['3.5', 14, 'Cards on a page'],
  ['4', 16, 'Card padding'],
  ['5', 20, 'Page gutter, page sections'],
  ['6', 24, 'Hero padding'],
] as const

// Literal class names so Tailwind can see them.
const RADII = [
  ['rounded-sm', '8', 'Tags, small stepper'],
  ['rounded-md', '10', 'Inputs, small buttons'],
  ['rounded-lg', '12', 'Large buttons'],
  ['rounded-xl', '14', 'Cards, primary action'],
  ['rounded-2xl', '16', 'Cards on brand'],
  ['rounded-3xl', '18', 'Saldo card'],
  ['rounded-t-sheet', '24', 'Sheet top corners'],
  ['rounded-full', '∞', 'Pills, avatars'],
] as const

const ICONS = [
  [ChevronLeft, 'ChevronLeft'],
  [ChevronRight, 'ChevronRight'],
  [X, 'X'],
  [Search, 'Search'],
  [Settings, 'Settings'],
  [EllipsisVertical, 'EllipsisVertical'],
  [Plus, 'Plus'],
  [Minus, 'Minus'],
  [Check, 'Check'],
  [Copy, 'Copy'],
  [Upload, 'Upload'],
  [Image, 'Image'],
  [AlignLeft, 'AlignLeft'],
  [CreditCard, 'CreditCard'],
  [Clock, 'Clock'],
  [Bell, 'Bell'],
  [House, 'House'],
  [Users, 'Users'],
  [Monitor, 'Monitor'],
] as const

// Read the hex from the live CSS variable so the swatches can't drift from the tokens.
function readTokenValues(names: readonly string[]): Record<string, string> {
  const style = getComputedStyle(document.documentElement)
  return Object.fromEntries(
    names.map((n) => [n, style.getPropertyValue(`--color-${n}`).trim().toUpperCase()]),
  )
}

const ALL_COLOR_NAMES = COLORS.flatMap((g) => g.tokens.map(([name]) => name))

export function FoundationsSection() {
  const [hex] = useState(() => readTokenValues(ALL_COLOR_NAMES))

  return (
    <Group
      id="foundations"
      description="Tokens every component is built from. Change a value in styles/index.css, never in a component."
    >
      <Specimen name="Color" usage="Use by role, not by look. Money colors mean money only.">
        {COLORS.map((group) => (
          <div key={group.group}>
            <p className="text-2xs text-ink-muted mb-1.5 font-bold tracking-wide uppercase">
              {group.group}
            </p>
            <ul className="divide-line-soft bg-card divide-y rounded-xl px-3">
              {group.tokens.map(([name, role]) => (
                <li key={name} className="flex items-center gap-3 py-2">
                  <span
                    aria-hidden
                    className="border-line-soft size-9 shrink-0 rounded-md border"
                    style={{ background: `var(--color-${name})` }}
                  />
                  <span className="min-w-0 flex-1">
                    <span className="block font-mono text-sm font-semibold">{name}</span>
                    <span className="text-ink-muted block text-xs">{role}</span>
                  </span>
                  <span className="text-ink-muted font-mono text-xs">{hex[name]}</span>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </Specimen>

      <Specimen
        name="Typography"
        usage="Plus Jakarta Sans, tabular numerals. Use these roles; text-sm is 13px in this project."
        surface="card"
        className="gap-0 py-1"
      >
        {TYPE.map(([role, classes, spec, sample]) => (
          <div key={role} className="border-line-soft border-b py-2.5 last:border-0">
            <div className="text-2xs text-ink-muted mb-0.5 flex justify-between font-mono">
              <span>{role}</span>
              <span>{spec}</span>
            </div>
            <p className={cn('truncate', classes)}>{sample}</p>
          </div>
        ))}
      </Specimen>

      <Specimen name="Spacing" usage="4px grid. Pick the step by relationship, not by eye.">
        <ul className="bg-card flex flex-col gap-2 rounded-xl p-4">
          {SPACING.map(([step, px, use]) => (
            <li key={step} className="grid grid-cols-[4.5rem_1fr] items-center gap-3">
              <span className="font-mono text-xs whitespace-nowrap">
                {step} · {px}
              </span>
              <span className="flex items-center gap-2">
                <span aria-hidden className="bg-gold h-3 rounded-sm" style={{ width: px * 2 }} />
                <span className="text-ink-muted truncate text-xs">{use}</span>
              </span>
            </li>
          ))}
        </ul>
      </Specimen>

      <Specimen name="Radius" usage="Bigger surfaces get bigger corners.">
        <div className="grid grid-cols-4 gap-2">
          {RADII.map(([name, px, use]) => (
            <div key={name} className="flex flex-col items-center gap-1 text-center">
              <span aria-hidden className={cn('border-line bg-card size-14 border-2', name)} />
              <span className="text-2xs font-mono font-semibold">
                {name.replace('rounded-', '')} · {px}
              </span>
              <span className="text-2xs text-ink-muted leading-tight">{use}</span>
            </div>
          ))}
        </div>
      </Specimen>

      <Specimen name="Elevation" usage="Cards are flat. Only layers above the page cast a shadow.">
        <div className="grid grid-cols-3 gap-3 text-center text-xs">
          <div className="bg-card rounded-xl px-2 py-5">Flat (card)</div>
          <div className="bg-card shadow-overlay rounded-xl px-2 py-5">overlay</div>
          <div className="bg-card shadow-toast rounded-xl px-2 py-5">toast</div>
        </div>
      </Specimen>

      <Specimen
        name="Iconography"
        usage="lucide-react, 20px in buttons (22px in menu tiles), stroke 2 (1.8 in tiles). Always aria-hidden next to text."
        surface="card"
      >
        <div className="grid grid-cols-4 gap-y-3">
          {ICONS.map(([Icon, name]) => (
            <div key={name} className="flex flex-col items-center gap-1">
              <Icon aria-hidden className="size-5" />
              <span className="text-2xs text-ink-muted max-w-full truncate">{name}</span>
            </div>
          ))}
        </div>
      </Specimen>

      <Specimen name="Motion" usage="Short and functional; off under prefers-reduced-motion.">
        <dl className="divide-line-soft bg-card divide-y rounded-xl px-4 text-sm">
          {[
            ['150ms', 'Color and opacity transitions (default)'],
            ['160ms ease-out', 'Toast enters'],
            ['220ms ease-out', 'Sheet slides up'],
          ].map(([value, use]) => (
            <div key={value} className="flex justify-between gap-3 py-2.5">
              <dt className="font-mono font-semibold">{value}</dt>
              <dd className="text-ink-muted text-right">{use}</dd>
            </div>
          ))}
        </dl>
      </Specimen>
    </Group>
  )
}
