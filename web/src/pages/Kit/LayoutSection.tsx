import {
  Avatar,
  Card,
  HeroHeader,
  IconButton,
  MosqueLogo,
  PageBody,
  Stack,
  Stat,
} from '@/components'
import { Group, Specimen } from '@/pages/Kit/Specimen'

export function LayoutSection() {
  return (
    <Group
      id="layout"
      description="One 390px column (max 480, centered). 20px gutters, 14px between cards, safe areas respected."
    >
      <Specimen name="HeroHeader" usage="Green header that opens Home; runs under the status bar.">
        <HeroHeader className="-mx-5 pt-6">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-on-brand-muted text-sm">
                Assalamu'alaikum, [Ketua DKM] · Super admin
              </p>
              <p className="mt-0.5 text-xl font-extrabold">Masjid [NAMA MASJID]</p>
            </div>
            <IconButton
              variant="brand"
              label="Akun saya"
              icon={<Avatar name="[Ketua DKM]" className="bg-transparent" />}
            />
          </div>
          <Card tone="brand-2" padding="md" className="flex items-center justify-between">
            <Stat label="Sholat berikutnya" value="Maghrib 18:15" size="lg" />
            <Stat label="dalam" value="00:47" size="lg" align="end" valueClassName="text-gold" />
          </Card>
        </HeroHeader>
      </Specimen>

      <Specimen
        name="PageBody · Stack"
        usage="PageBody is the screen column. Stack gaps: xs 6 (label→control), sm 10 (inside cards), md 14 (cards), lg 20 (sections)."
      >
        <PageBody className="border-dash -mx-0 rounded-xl border-2 border-dashed px-3 py-3">
          <Stack gap="sm">
            <Card padding="md">Card</Card>
            <Card padding="md">Card</Card>
          </Stack>
        </PageBody>
      </Specimen>

      <Specimen
        name="ActionBar"
        usage="Sticky bottom bar holding the screen's primary action. Pinned to the bottom of this page."
      >
        <p className="text-ink-muted text-xs">See “Simpan dan kirim ke layar” below.</p>
      </Specimen>

      <Specimen
        name="MosqueLogo"
        usage="Placeholder mark until a mosque uploads its logo."
        surface="brand"
      >
        <div className="flex items-center gap-3">
          <MosqueLogo size={56} />
          <div>
            <p className="text-xl font-extrabold">Panel Pengurus</p>
            <p className="text-on-brand-muted text-sm">Masjid [NAMA MASJID]</p>
          </div>
        </div>
      </Specimen>
    </Group>
  )
}
