import { describe, expect, it } from 'vitest'
import { setupTabs } from './tabs'

const render = () => {
  document.body.innerHTML = `
    <div data-tabs>
      <div role="tablist">
        <button role="tab" aria-controls="m-normal" aria-selected="true">Normal</button>
        <button role="tab" aria-controls="m-adzan" aria-selected="false">Adzan</button>
        <button role="tab" aria-controls="m-jumat" aria-selected="false">Jumat</button>
      </div>
      <div role="tabpanel" id="m-normal"><iframe data-src="https://demo.masjid.id"></iframe></div>
      <div role="tabpanel" id="m-adzan" hidden><iframe data-src="https://demo.masjid.id/tv"></iframe></div>
      <div role="tabpanel" id="m-jumat" hidden>jumat</div>
    </div>`
  setupTabs(document)
  return [...document.querySelectorAll<HTMLButtonElement>('[role="tab"]')]
}

const visible = () =>
  [...document.querySelectorAll<HTMLElement>('[role="tabpanel"]')]
    .filter((p) => !p.hidden)
    .map((p) => p.id)

const frameSrc = (panel: string) =>
  document.querySelector(`#${panel} iframe`)?.getAttribute('src') ?? null

describe('setupTabs', () => {
  it('shows the clicked panel', () => {
    const [, adzan] = render()
    adzan!.click()
    expect(visible()).toEqual(['m-adzan'])
    expect(adzan!.getAttribute('aria-selected')).toBe('true')
  })

  it('moves between tabs with the arrow keys, wrapping around', () => {
    const [normal] = render()
    normal!.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowLeft' }))
    expect(visible()).toEqual(['m-jumat'])
    expect(document.activeElement?.textContent).toBe('Jumat')
  })

  it('loads a panel’s demo frame only once the panel is shown', () => {
    const [, adzan] = render()
    expect(frameSrc('m-normal')).toBe('https://demo.masjid.id')
    expect(frameSrc('m-adzan')).toBeNull()

    adzan!.click()
    expect(frameSrc('m-adzan')).toBe('https://demo.masjid.id/tv')
  })
})
