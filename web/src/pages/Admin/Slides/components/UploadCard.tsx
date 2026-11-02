import { useState } from 'react'
import { UploadTile, useToast } from '@/components'
import { NewSlideDrawer } from '../drawers/NewSlideDrawer'
import { SLIDE_TYPES, slideProblem } from '../utils'

/**
 * Picking an image opens the new-slide sheet to fill in its title, caption and duration; the
 * upload happens on Simpan. The file is checked here first so a bad one never opens the sheet.
 */
export function UploadCard({ onUploaded }: { onUploaded: () => void }) {
  const [file, setFile] = useState<File | null>(null)
  const toast = useToast()

  const pick = ([picked]: File[]) => {
    if (!picked) return
    const problem = slideProblem(picked)
    if (problem) toast({ tone: 'danger', message: problem, description: picked.name })
    else setFile(picked)
  }

  return (
    <>
      <UploadTile
        label="Upload gambar"
        hint="JPG/PNG/WebP, maks 5 MB · disarankan 16:9"
        accept={SLIDE_TYPES.join(',')}
        onFiles={pick}
      />
      {file && (
        <NewSlideDrawer
          file={file}
          onClose={() => setFile(null)}
          onCreated={() => {
            setFile(null)
            onUploaded()
          }}
        />
      )}
    </>
  )
}
