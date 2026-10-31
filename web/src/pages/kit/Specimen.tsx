import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'
import { GROUPS, type GroupId } from './groups'

// Building blocks for the /kit reference page.

type GroupProps = { id: GroupId; description: string; children: ReactNode }

export function Group({ id, description, children }: GroupProps) {
  const index = GROUPS.findIndex((g) => g.id === id)
  const group = GROUPS[index]
  return (
    <section id={id} aria-labelledby={`${id}-title`} className="flex scroll-mt-28 flex-col gap-6">
      <header className="border-b border-line pb-3">
        <p className="font-mono text-xs font-semibold text-gold-ink">
          {String(index + 1).padStart(2, '0')}
        </p>
        <h2 id={`${id}-title`} className="text-xl font-extrabold">
          {group?.label}
        </h2>
        <p className="mt-0.5 text-sm text-pretty text-ink-muted">{description}</p>
      </header>
      {children}
    </section>
  )
}

type SpecimenProps = {
  name: string
  usage: string
  /** `card` previews on white, `canvas` on the page ground, `brand` on green. */
  surface?: 'canvas' | 'card' | 'brand'
  className?: string
  children: ReactNode
}

export function Specimen({ name, usage, surface = 'canvas', className, children }: SpecimenProps) {
  return (
    <article className="flex flex-col gap-2.5">
      <div>
        <h3 className="font-mono text-sm font-semibold">{name}</h3>
        <p className="text-xs text-pretty text-ink-muted">{usage}</p>
      </div>
      <div
        className={cn(
          'flex flex-col gap-3',
          surface === 'card' && 'rounded-xl bg-card p-4',
          surface === 'brand' && 'rounded-3xl bg-brand p-5 text-on-brand',
          className,
        )}
      >
        {children}
      </div>
    </article>
  )
}

export function Row({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={cn('flex flex-wrap items-center gap-2', className)}>{children}</div>
}
