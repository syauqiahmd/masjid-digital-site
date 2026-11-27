import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { setupTabs } from './tabs'

const render = () => {
  document.body.innerHTML = `
    <div data-tabs>
      <div role="tablist">
        <button role="tab" aria-controls="m-normal" aria-selected="true">Normal</button>
        <button role="tab" aria-controls="m-adzan" aria-selected="false">Adzan</button>
        <button role="tab" aria-controls="m-jumat" aria-selected="false">Jumat</button>
      </div>
      <div role="tabpanel" id="m-normal">normal</div>
      <div role="tabpanel" id="m-adzan" hidden>adzan</div>
      <div role="tabpanel" id="m-jumat" hidden>jumat</div>
    </div>`
  setupTabs(document)
  return [...document.querySelectorAll<HTMLButtonElement>('[role="tab"]')]
}

const visible = () =>
  [...document.querySelectorAll<HTMLElement>('[role="tabpanel"]')]
    .filter((p) => !p.hidden)
    .map((p) => p.id)

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
})

describe('setupTabs autoplay', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    document.body.innerHTML = `
      <div data-tabs data-tabs-autoplay="5000">
        <div role="tablist">
          <button role="tab" aria-controls="a-normal" aria-selected="true">Normal</button>
          <button role="tab" aria-controls="a-adzan" aria-selected="false">Adzan</button>
          <button role="tab" aria-controls="a-iqamah" aria-selected="false">Iqamah</button>
        </div>
        <div role="tabpanel" id="a-normal">normal</div>
        <div role="tabpanel" id="a-adzan" hidden>adzan</div>
        <div role="tabpanel" id="a-iqamah" hidden>iqamah</div>
      </div>`
    setupTabs(document)
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  const tick = async (ms: number) => {
    await vi.advanceTimersByTimeAsync(ms)
  }
  const widget = () => document.querySelector<HTMLElement>('[data-tabs]')!
  const tab = (name: string) =>
    [...document.querySelectorAll<HTMLButtonElement>('[role="tab"]')].find(
      (t) => t.textContent === name,
    )!

  it('moves to the next mode every interval, wrapping around', async () => {
    await tick(4999)
    expect(visible()).toEqual(['a-normal'])
    await tick(1)
    expect(visible()).toEqual(['a-adzan'])
    await tick(10_000)
    expect(visible()).toEqual(['a-normal'])
  })

  it('stops for good once a tab is picked', async () => {
    tab('Iqamah').click()
    await tick(30_000)
    expect(visible()).toEqual(['a-iqamah'])
  })

  it('pauses while the pointer is over it', async () => {
    widget().dispatchEvent(new Event('pointerenter'))
    await tick(12_000)
    expect(visible()).toEqual(['a-normal'])

    widget().dispatchEvent(new Event('pointerleave'))
    await tick(5000)
    expect(visible()).toEqual(['a-adzan'])
  })
})
