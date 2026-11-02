/**
 * Which slide is on screen `seconds` into the epoch, each slide staying for its own duration and
 * the list looping. Counted from the (server-synced) clock rather than a timer, so every TV
 * shows the same slide and a reload picks up where it was. -1 when there are no slides.
 */
export function slideIndexAt(durations: readonly number[], seconds: number) {
  const total = durations.reduce((sum, d) => sum + d, 0)
  if (total <= 0) return -1

  let t = Math.floor(seconds) % total
  for (const [i, d] of durations.entries()) {
    if (t < d) return i
    t -= d
  }
  return durations.length - 1 // unreachable
}
