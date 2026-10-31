import { Link } from 'react-router'
import { buttonStyles, EmptyState, HeroHeader, PageBody } from '@/components'

// Beranda (E3-02). Placeholder until auth and the mosque profile exist.
export function HomePage() {
  return (
    <>
      <HeroHeader>
        <div>
          <p className="text-sm text-on-brand-muted">Assalamu'alaikum</p>
          <h1 className="mt-0.5 text-xl font-extrabold">Panel Pengurus</h1>
        </div>
      </HeroHeader>
      <PageBody className="pt-5">
        <EmptyState
          title="Belum ada modul"
          description="Menu Kelola muncul di sini sesuai peran setelah modul dibuat."
          action={
            import.meta.env.DEV && (
              <Link to="/kit" className={buttonStyles({ variant: 'secondary', size: 'md' })}>
                Lihat komponen
              </Link>
            )
          }
        />
      </PageBody>
    </>
  )
}
