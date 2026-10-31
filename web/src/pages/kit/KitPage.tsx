import { ActionBar, Button, PageBody, PageHeader } from '@/components'
import { ActionsSection } from './ActionsSection'
import { DataDisplaySection } from './DataDisplaySection'
import { FeedbackSection } from './FeedbackSection'
import { FormsSection } from './FormsSection'
import { FoundationsSection } from './FoundationsSection'
import { LayoutSection } from './LayoutSection'
import { NavigationSection } from './NavigationSection'
import { OverlaysSection } from './OverlaysSection'
import { GROUPS } from './groups'

/*
 * Dev-only design-system reference (route /kit). Mirrors docs/design/design-system.md, group by
 * group, with live components and the mockup's placeholder data.
 */
export function KitPage() {
  return (
    <>
      <PageHeader title="Design system" meta="Masjid Digital · panel pengurus" backTo="/" />
      <nav
        aria-label="Bagian design system"
        className="sticky top-0 z-20 border-b border-line-soft bg-canvas/95 backdrop-blur-sm"
      >
        <ul className="flex [scrollbar-width:none] gap-1.5 overflow-x-auto px-5 py-2">
          {GROUPS.map((group) => (
            <li key={group.id} className="shrink-0">
              <a
                href={`#${group.id}`}
                className="inline-flex h-9 items-center rounded-full px-3 text-sm font-semibold text-ink-muted hover:bg-track hover:text-ink"
              >
                {group.label}
              </a>
            </li>
          ))}
        </ul>
      </nav>
      <PageBody className="gap-12 pt-6">
        <FoundationsSection />
        <ActionsSection />
        <FormsSection />
        <DataDisplaySection />
        <FeedbackSection />
        <OverlaysSection />
        <NavigationSection />
        <LayoutSection />
      </PageBody>
      <ActionBar>
        <Button size="xl" block>
          Simpan dan kirim ke layar
        </Button>
      </ActionBar>
    </>
  )
}
