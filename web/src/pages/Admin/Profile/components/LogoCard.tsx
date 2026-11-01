import { LoaderCircle } from 'lucide-react'
import { useState } from 'react'
import {
  Button,
  buttonStyles,
  Card,
  ConfirmDialog,
  MosqueLogo,
  Skeleton,
  useToast,
} from '@/components'
import { useDeleteMosqueLogo, useMosqueLogo, useUploadMosqueLogo } from '@/hooks/useMosqueProfile'
import { cn } from '@/utils/cn'
import { LOGO_TYPES, logoProblem } from '../utils'

// The logo has its own endpoint, so it saves as soon as a file is picked, apart from "Simpan".
export function LogoCard() {
  const logo = useMosqueLogo()
  const upload = useUploadMosqueLogo()
  const remove = useDeleteMosqueLogo()
  const toast = useToast()
  const [confirmDelete, setConfirmDelete] = useState(false)
  const busy = upload.isPending || remove.isPending

  const onFile = (file: File | undefined) => {
    if (!file) return
    const problem = logoProblem(file)
    if (problem) {
      toast({ tone: 'danger', message: problem })
      return
    }
    upload.mutate(file, {
      onSuccess: () => toast({ message: 'Logo masjid disimpan' }),
      onError: (error) => toast({ tone: 'danger', message: error.message }),
    })
  }

  const onDelete = () =>
    remove.mutate(undefined, {
      onSuccess: () => toast({ message: 'Logo masjid dihapus' }),
      onError: (error) => toast({ tone: 'danger', message: error.message }),
    })

  return (
    <Card className="flex items-center gap-4">
      {logo.isPending ? (
        <Skeleton className="size-18 shrink-0 rounded-2xl" />
      ) : logo.data ? (
        <img
          src={logo.data}
          alt="Logo masjid"
          className="size-18 shrink-0 rounded-2xl object-cover"
        />
      ) : (
        <div className="bg-brand flex size-18 shrink-0 items-center justify-center rounded-2xl">
          <MosqueLogo size={56} />
        </div>
      )}

      <div className="min-w-0 flex-1">
        <h2 className="text-base font-bold">Logo masjid</h2>
        <p className="text-ink-muted mt-0.5 text-xs">PNG/JPG/WebP, maks 2 MB, tampil persegi</p>
        {logo.isError && (
          <p className="text-expense mt-0.5 text-xs font-semibold">{logo.error.message}</p>
        )}
        <div className="mt-2 flex flex-wrap gap-2">
          {/* A label around a hidden input, so the button opens the file picker natively. */}
          <label
            className={cn(
              buttonStyles({ variant: 'secondary', size: 'sm' }),
              'focus-within:outline-gold cursor-pointer focus-within:outline-3 focus-within:outline-offset-2',
              busy && 'pointer-events-none opacity-50',
            )}
          >
            <input
              type="file"
              accept={LOGO_TYPES.join(',')}
              className="sr-only"
              disabled={busy}
              onChange={(event) => {
                onFile(event.currentTarget.files?.[0])
                event.currentTarget.value = ''
              }}
            />
            {upload.isPending && <LoaderCircle aria-hidden className="animate-spin" />}
            {logo.data ? 'Ganti logo' : 'Unggah logo'}
          </label>
          {logo.data && (
            <Button
              size="sm"
              variant="danger"
              disabled={busy}
              onClick={() => setConfirmDelete(true)}
            >
              Hapus
            </Button>
          )}
        </div>
      </div>

      <ConfirmDialog
        open={confirmDelete}
        onClose={() => setConfirmDelete(false)}
        title="Hapus logo masjid?"
        description="Logo bawaan akan dipakai sampai logo baru diunggah."
        confirmLabel="Hapus logo"
        danger
        onConfirm={onDelete}
      />
    </Card>
  )
}
