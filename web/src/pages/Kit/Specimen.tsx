import type { ReactNode } from 'react'
import { GROUPS, type GroupId } from '@/pages/Kit/groups'
import { cn } from '@/utils/cn'

// Building blocks for the /kit reference page.

type GroupProps = { id: GroupId; description: string; children: ReactNode }

export function Group({ id, description, children }: GroupProps) {
  const index = GROUPS.findIndex((g) => g.id === id)
  const group = GROUPS[index]
  return (
    <section id={id} aria-labelledby={`${id}-title`} className="flex scroll-mt-28 flex-col gap-6">
      <header className="border-line border-b pb-3">
        <p className="text-gold-ink font-mono text-xs font-semibold">
          {String(index + 1).padStart(2, '0')}
        </p>
        <h2 id={`${id}-title`} className="text-xl font-extrabold">
          {group?.label}
        </h2>
        <p className="text-ink-muted mt-0.5 text-sm text-pretty">{description}</p>
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
        <p className="text-ink-muted text-xs text-pretty">{usage}</p>
      </div>
      <div
        className={cn(
          'flex flex-col gap-3',
          surface === 'card' && 'bg-card rounded-xl p-4',
          surface === 'brand' && 'bg-brand text-on-brand rounded-3xl p-5',
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
