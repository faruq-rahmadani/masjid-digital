import { addDays, weekdayOf } from '@/utils/prayer/calendar'

const WEEK = 7

/** The next `count` Fridays from `today` ("YYYY-MM-DD"), today included when it is a Friday. */
export function nextFridays(today: string, count: number) {
  const offset = Array.from({ length: WEEK }, (_, i) => i).find(
    (i) => weekdayOf(addDays(today, i)) === 'friday',
  ) as number
  return Array.from({ length: count }, (_, i) => addDays(today, offset + i * WEEK))
}
