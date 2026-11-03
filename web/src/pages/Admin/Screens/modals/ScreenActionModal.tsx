import { ConfirmDialog } from '@/components'
import type { Screen, ScreenMode } from '@/types/screen'
import { modeLabel } from '../utils'

/** An action that needs a confirm before it reaches the TV. */
export type ScreenAction =
  | { kind: 'mode'; screen: Screen; mode: ScreenMode }
  | { kind: 'reload'; screen: Screen }
  | { kind: 'reload-all' }
  | { kind: 'block'; screen: Screen }
  | { kind: 'unblock'; screen: Screen }

type Copy = { title: string; description: string; confirmLabel: string; danger?: boolean }

const screenName = (screen: Screen) => screen.name || 'layar baru ini'

const modeDescription = (mode: ScreenMode) =>
  mode === 'auto'
    ? 'Layar kembali menampilkan sesuai jadwal sholat.'
    : `${modeLabel(mode)} sampai pukul 00:00, lalu layar kembali ikut jadwal.`

const COPY: { [K in ScreenAction['kind']]: (action: Extract<ScreenAction, { kind: K }>) => Copy } =
  {
    mode: ({ screen, mode }) => ({
      title: `Ubah mode ${screenName(screen)}?`,
      description: modeDescription(mode),
      confirmLabel: 'Ubah mode',
    }),
    reload: ({ screen }) => ({
      title: `Muat ulang ${screenName(screen)}?`,
      description: 'Layar memuat ulang halaman dan tampil lagi dalam beberapa detik.',
      confirmLabel: 'Muat ulang',
    }),
    'reload-all': () => ({
      title: 'Muat ulang semua layar?',
      description:
        'Semua layar yang online memuat ulang halaman, misalnya setelah pembaruan aplikasi.',
      confirmLabel: 'Muat ulang',
    }),
    block: ({ screen }) => ({
      title: `Blokir ${screenName(screen)}?`,
      description:
        'Layar disembunyikan dari daftar dan tidak bisa terhubung. Bisa dibuka lagi dari daftar layar diblokir.',
      confirmLabel: 'Blokir',
      danger: true,
    }),
    unblock: ({ screen }) => ({
      title: `Buka blokir ${screenName(screen)}?`,
      description: 'Layar kembali ke daftar dan tersambung lagi dalam 5 menit.',
      confirmLabel: 'Buka blokir',
    }),
  }

type ScreenActionModalProps = {
  /** null keeps the modal closed. */
  action: ScreenAction | null
  onClose: () => void
  onConfirm: (action: ScreenAction) => void
}

export function ScreenActionModal({ action, onClose, onConfirm }: ScreenActionModalProps) {
  const copy = action && (COPY[action.kind] as (a: ScreenAction) => Copy)(action)

  return (
    <ConfirmDialog
      open={action !== null}
      onClose={onClose}
      title={copy?.title ?? ''}
      description={copy?.description}
      confirmLabel={copy?.confirmLabel ?? ''}
      danger={copy?.danger}
      onConfirm={() => action && onConfirm(action)}
    />
  )
}
