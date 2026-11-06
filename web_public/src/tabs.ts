/**
 * Tabs: the hero's TV modes and the demo picker. Buttons switch the visible panel, arrow keys
 * move between them, and a panel's demo frames (`iframe[data-src]`) only start loading once the
 * panel is shown, so the page doesn't open the whole demo app for every visitor.
 *
 * Markup: `[data-tabs]` holding `[role=tab][aria-controls]` buttons and `[role=tabpanel]`s.
 */
export function setupTabs(root: ParentNode) {
  for (const widget of root.querySelectorAll<HTMLElement>('[data-tabs]')) {
    const tabs = [...widget.querySelectorAll<HTMLButtonElement>('[role="tab"]')]
    const select = (index: number) => {
      tabs.forEach((tab, i) => {
        const selected = i === index
        tab.setAttribute('aria-selected', String(selected))
        tab.tabIndex = selected ? 0 : -1
        const panel = widget.querySelector<HTMLElement>(`#${tab.getAttribute('aria-controls')}`)
        if (!panel) return
        panel.hidden = !selected
        if (selected) loadFrames(panel)
      })
    }
    const step: Record<string, (i: number) => number> = {
      ArrowRight: (i) => (i + 1) % tabs.length,
      ArrowDown: (i) => (i + 1) % tabs.length,
      ArrowLeft: (i) => (i - 1 + tabs.length) % tabs.length,
      ArrowUp: (i) => (i - 1 + tabs.length) % tabs.length,
      Home: () => 0,
      End: () => tabs.length - 1,
    }

    tabs.forEach((tab, i) => {
      tab.addEventListener('click', () => select(i))
      tab.addEventListener('keydown', (e) => {
        const next = step[e.key]?.(i)
        if (next === undefined) return
        e.preventDefault()
        select(next)
        tabs[next]?.focus()
      })
    })
    select(
      Math.max(
        0,
        tabs.findIndex((tab) => tab.getAttribute('aria-selected') === 'true'),
      ),
    )
  }
}

/** Starts the panel's frames; `loading="lazy"` on them still waits until they near the viewport. */
function loadFrames(panel: HTMLElement) {
  for (const frame of panel.querySelectorAll<HTMLIFrameElement>('iframe[data-src]')) {
    if (!frame.getAttribute('src')) frame.src = frame.dataset.src ?? ''
  }
}
