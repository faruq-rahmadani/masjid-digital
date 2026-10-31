import { EmptyState, PageBody, PageHeader, TextLink } from '@/components'

// Placeholder for menu items whose screen isn't built yet (see features/home/menu.ts).
export function ComingSoonPage({ title }: { title: string }) {
  return (
    <>
      <PageHeader title={title} backTo="/" />
      <PageBody>
        <EmptyState
          title="Segera hadir"
          description={`Menu ${title} sedang dibuat.`}
          action={<TextLink to="/">Kembali ke Beranda</TextLink>}
        />
      </PageBody>
    </>
  )
}
