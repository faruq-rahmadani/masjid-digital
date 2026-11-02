/** List filter for records that can be on or off the TV (running texts, slides). */
export type ActiveFilter = 'active' | 'inactive' | 'all'

export const matchesFilter = (item: { active: boolean }, filter: ActiveFilter) =>
  filter === 'all' || item.active === (filter === 'active')

/**
 * The API saves the order of every item, but a filtered list only reorders the ones it shows:
 * put `moved` (the shown ids in their new order) back into the slots the shown items held.
 */
export function mergeOrder(all: readonly number[], moved: readonly number[]) {
  const shown = new Set(moved)
  let next = 0
  return all.map((id) => (shown.has(id) ? (moved[next++] ?? id) : id))
}
