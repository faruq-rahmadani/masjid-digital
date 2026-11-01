import { useCallback, useEffect, useRef, useState } from 'react'

const TICK_MS = 100

/**
 * A clock that starts at `initial` and runs `speed` times faster than real time. It can pause and
 * jump; each change re-anchors it at the simulated time reached so far.
 */
export function useSimulatedClock(initial: Date) {
  const [speed, setSpeed] = useState(1)
  const [paused, setPaused] = useState(false)
  const [now, setNow] = useState(initial)
  const anchor = useRef({ simulated: initial.getTime(), real: Date.now() })

  const simulatedTime = useCallback(
    () => anchor.current.simulated + (Date.now() - anchor.current.real) * (paused ? 0 : speed),
    [paused, speed],
  )

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date(simulatedTime())), TICK_MS)
    return () => clearInterval(timer)
  }, [simulatedTime])

  const reanchor = (simulated: number) => {
    anchor.current = { simulated, real: Date.now() }
  }

  return {
    now,
    speed,
    paused,
    changeSpeed: (next: number) => {
      reanchor(simulatedTime())
      setSpeed(next)
    },
    togglePaused: () => {
      reanchor(simulatedTime())
      setPaused((value) => !value)
    },
    jump: (date: Date) => {
      reanchor(date.getTime())
      setNow(date)
    },
  }
}
