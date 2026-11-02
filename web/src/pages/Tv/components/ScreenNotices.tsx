import { TvStage } from './TvStage'

// An admin blocked this TV on the Layar TV page. It checks back every 5 minutes, so it comes
// back by itself once unblocked.
export function BlockedScreen() {
  return (
    <TvStage className="bg-brand">
      <div className="text-on-brand m-auto flex flex-col items-center gap-4 text-center">
        <p className="text-tv-40 font-bold">Layar ini diblokir</p>
        <p className="text-on-brand-muted text-tv-26">
          Minta pengurus masjid membuka blokir di menu Layar TV.
        </p>
      </div>
    </TvStage>
  )
}

// The live connection has been down for a while: the TV shows what it has, and says so.
export function OfflineMark() {
  return (
    <div
      role="status"
      className="bg-expense-soft text-expense-ink text-tv-18 absolute right-8 bottom-24 rounded-full px-5 py-2 font-semibold"
    >
      Offline · menampilkan data tersimpan
    </div>
  )
}

// Browsers keep audio off until the page is tapped once; until then the alarm can't sound.
export function SoundPrompt() {
  return (
    <div
      role="status"
      className="bg-gold text-on-gold text-tv-18 absolute top-6 left-1/2 -translate-x-1/2 rounded-full px-5 py-2 font-semibold"
    >
      Ketuk layar sekali untuk menyalakan suara alarm
    </div>
  )
}
