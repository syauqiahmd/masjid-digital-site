import { beforeEach, describe, expect, it, vi } from 'vitest'
import { applyDemoAccount, demoAccount, setupDemoStage } from './demo'

beforeEach(() => {
  document.body.innerHTML = `
    <div data-demo-stage>
      <img data-demo-placeholder src="/screenshots/login.png" alt="" />
      <iframe data-demo-frame="panel" data-src="https://demo.masjid.id" hidden></iframe>
      <img data-demo-placeholder src="/screenshots/tv.png" alt="" />
      <iframe data-demo-frame="tv" data-src="https://demo.masjid.id/tv" hidden></iframe>
      <button type="button" data-demo-start>Mulai demo</button>
      <p data-demo-running hidden>Ubah running text di HP.</p>
    </div>`
  setupDemoStage(document)
})

const frames = () => [...document.querySelectorAll('iframe')]

describe('setupDemoStage', () => {
  it('loads nothing until the visitor starts the demo', () => {
    expect(frames().map((f) => f.getAttribute('src'))).toEqual([null, null])
  })

  it('starts both apps side by side on click', () => {
    document.querySelector<HTMLButtonElement>('[data-demo-start]')!.click()

    expect(frames().map((f) => f.getAttribute('src'))).toEqual([
      'https://demo.masjid.id',
      'https://demo.masjid.id/tv',
    ])
    expect(frames().every((f) => !f.hidden)).toBe(true)
    expect(
      [...document.querySelectorAll<HTMLElement>('[data-demo-placeholder]')].every((p) => p.hidden),
    ).toBe(true)
    expect(document.querySelector<HTMLElement>('[data-demo-running]')!.hidden).toBe(false)
    expect(document.querySelector<HTMLElement>('[data-demo-start]')!.hidden).toBe(true)
  })
})

describe('applyDemoAccount', () => {
  const box = () => document.querySelector<HTMLElement>('[data-demo-account]')!
  beforeEach(() => {
    document.body.innerHTML = `
      <div data-demo-account hidden>
        <span data-demo-account-phone></span><button data-copy="phone">Salin</button>
        <span data-demo-account-password></span><button data-copy="password">Salin</button>
      </div>`
  })

  it('reads the account from the env, only when both halves are set', () => {
    expect(demoAccount({ VITE_DEMO_PHONE: ' 0812 ', VITE_DEMO_PASSWORD: 'rahasia' })).toEqual({
      phone: '0812',
      password: 'rahasia',
    })
    expect(demoAccount({ VITE_DEMO_PHONE: '0812' })).toBeNull()
  })

  it('stays hidden without an account', () => {
    applyDemoAccount(document, null)
    expect(box().hidden).toBe(true)
  })

  it('shows the account and copies each value', async () => {
    const writeText = vi.fn(async () => {})
    applyDemoAccount(document, { phone: '081234567890', password: 'rahasia' }, { writeText })

    expect(box().hidden).toBe(false)
    expect(document.querySelector('[data-demo-account-phone]')!.textContent).toBe('081234567890')

    const copyPassword = document.querySelector<HTMLButtonElement>('[data-copy="password"]')!
    copyPassword.click()
    await Promise.resolve()
    expect(writeText).toHaveBeenCalledWith('rahasia')
    await vi.waitFor(() => expect(copyPassword.textContent).toBe('Tersalin'))
  })
})
