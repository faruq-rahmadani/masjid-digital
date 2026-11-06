import { describe, expect, it } from 'vitest'
import { applyDemoLinks, demoUrls } from './demoLinks'

describe('demoUrls', () => {
  it('derives the TV demo from the panel demo', () => {
    expect(demoUrls({ VITE_DEMO_URL: 'https://demo.masjid.id/' })).toEqual({
      panel: 'https://demo.masjid.id/',
      tv: 'https://demo.masjid.id/tv',
    })
  })

  it('takes an explicit TV demo, and nothing when unset', () => {
    expect(
      demoUrls({
        VITE_DEMO_URL: 'https://demo.masjid.id',
        VITE_DEMO_TV_URL: 'https://tv.masjid.id',
      }).tv,
    ).toBe('https://tv.masjid.id')
    expect(demoUrls({ VITE_DEMO_URL: '  ' })).toEqual({ panel: null, tv: null })
  })
})

describe('applyDemoLinks', () => {
  it('points the buttons at the demo and hides the ones without one', () => {
    document.body.innerHTML = `
      <a data-demo="panel" href="#">Coba demo</a>
      <a data-demo="tv" href="#">Lihat layar TV</a>`
    applyDemoLinks(document, { panel: 'https://demo.masjid.id', tv: null })

    const [panel, tv] = document.querySelectorAll('a')
    expect(panel!.getAttribute('href')).toBe('https://demo.masjid.id')
    expect(panel!.hidden).toBe(false)
    expect(tv!.hidden).toBe(true)
  })

  it('gives the demo frames their URL and hides the demo section without a demo', () => {
    document.body.innerHTML = `
      <section data-demo-only><iframe data-demo-frame="tv"></iframe></section>`
    applyDemoLinks(document, { panel: 'https://demo.masjid.id', tv: 'https://demo.masjid.id/tv' })
    expect(document.querySelector('iframe')!.dataset.src).toBe('https://demo.masjid.id/tv')
    expect(document.querySelector('section')!.hidden).toBe(false)

    applyDemoLinks(document, { panel: null, tv: null })
    expect(document.querySelector('section')!.hidden).toBe(true)
  })
})
