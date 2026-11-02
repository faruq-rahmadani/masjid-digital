import { render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { loopPosition, scrollSeconds, Ticker } from './Ticker'

describe('loopPosition', () => {
  it('is where the loop is at that moment, the same on every screen', () => {
    expect(loopPosition(95_500, 30)).toBe(5_500) // 95.5 s into 30 s loops
    expect(loopPosition(90_000, 30)).toBe(0)
  })
})

describe('Ticker', () => {
  const original = HTMLElement.prototype.getAnimations
  afterEach(() => {
    HTMLElement.prototype.getAnimations = original
  })

  it('jumps its scroll to where the server clock says the loop is', () => {
    // jsdom runs no CSS animations: stand in for the marquee's
    const animation = { currentTime: 0 as number | null }
    HTMLElement.prototype.getAnimations = () => [animation as unknown as Animation]
    const texts = ['Kajian Tafsir setiap Kamis']
    const serverNow = 1_793_596_812_345

    render(<Ticker texts={texts} speed="medium" syncTo={() => serverNow} />)

    expect(animation.currentTime).toBe(
      loopPosition(serverNow, scrollSeconds(texts.join(), 'medium')),
    )
  })

  it('starts at the top when not synced (the admin preview)', () => {
    const animation = { currentTime: 0 as number | null }
    HTMLElement.prototype.getAnimations = () => [animation as unknown as Animation]

    render(<Ticker texts={['Infaq']} speed="fast" />)

    expect(animation.currentTime).toBe(0)
  })
})
