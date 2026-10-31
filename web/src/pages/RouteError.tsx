import { isRouteErrorResponse, useRouteError } from 'react-router'
import { Button } from '@/components'

// Rendered in place of the whole shell when a route throws or fails to load.
export function RouteError() {
  const error = useRouteError()
  const detail = isRouteErrorResponse(error) ? `${error.status} ${error.statusText}` : undefined

  return (
    <main role="alert" className="max-w-app pt-safe-12 mx-auto flex flex-col gap-2 px-5 pb-12">
      <h1 className="text-xl font-extrabold">Halaman gagal dimuat</h1>
      <p className="text-ink-muted">Periksa koneksi internet lalu muat ulang.</p>
      {detail && <p className="text-ink-muted text-xs">{detail}</p>}
      <Button className="mt-4 self-start" onClick={() => window.location.reload()}>
        Muat ulang
      </Button>
    </main>
  )
}
