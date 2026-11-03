import type { AlarmSound } from '@/types/iqamah'

/** Loudest gain at volume 100; Web Audio's 1.0 clips on cheap TV speakers. */
const MAX_GAIN = 0.6

type Tone = { at: number; frequency: number; duration: number; type: OscillatorType }

// Synthesised, so the TV needs no audio files to cache.
const PATTERNS: Record<AlarmSound, Tone[]> = {
  // three short beeps
  beep: [0, 0.3, 0.6].map((at) => ({ at, frequency: 880, duration: 0.18, type: 'square' })),
  // two bell strikes, each a tone and its octave fading out
  bell: [0, 1.2].flatMap((at) => [
    { at, frequency: 660, duration: 1.1, type: 'sine' as const },
    { at, frequency: 1320, duration: 0.8, type: 'sine' as const },
  ]),
}

/**
 * Plays the alarm with Web Audio. Browsers start audio suspended until the page gets a tap, so a
 * TV needs one tap after it boots (PRD); `unlock` is called on that tap.
 */
export function createAlarmPlayer() {
  const AudioCtor = window.AudioContext as typeof AudioContext | undefined
  const context = AudioCtor ? new AudioCtor() : null

  return {
    /** No Web Audio, or still waiting for the first tap. */
    get locked() {
      return context?.state !== 'running'
    },
    get supported() {
      return context !== null
    },
    unlock: () => context?.resume().catch(() => {}),
    play(sound: AlarmSound, volume: number) {
      if (context?.state !== 'running') return
      const start = context.currentTime
      for (const tone of PATTERNS[sound]) {
        const oscillator = context.createOscillator()
        const gain = context.createGain()
        oscillator.type = tone.type
        oscillator.frequency.value = tone.frequency
        // a short fade in and an exponential fade out avoid clicks
        gain.gain.setValueAtTime(0.0001, start + tone.at)
        gain.gain.exponentialRampToValueAtTime(
          (volume / 100) * MAX_GAIN + 0.0001,
          start + tone.at + 0.01,
        )
        gain.gain.exponentialRampToValueAtTime(0.0001, start + tone.at + tone.duration)
        oscillator.connect(gain).connect(context.destination)
        oscillator.start(start + tone.at)
        oscillator.stop(start + tone.at + tone.duration + 0.05)
      }
    },
  }
}

export type AlarmPlayer = ReturnType<typeof createAlarmPlayer>
