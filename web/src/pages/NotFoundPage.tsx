import { Link } from 'react-router'
import { buttonStyles, EmptyState, PageBody, PageHeader } from '@/components'

export function NotFoundPage() {
  return (
    <>
      <PageHeader title="Halaman tidak ditemukan" backTo="/" />
      <PageBody>
        <EmptyState
          title="Halaman ini tidak ada"
          description="Alamatnya mungkin salah atau halamannya sudah dipindah."
          action={
            <Link to="/" className={buttonStyles({ size: 'md' })}>
              Ke Beranda
            </Link>
          }
        />
      </PageBody>
    </>
  )
}
