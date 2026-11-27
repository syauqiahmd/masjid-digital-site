import { describe, expect, it, vi } from 'vitest'
import {
  DEFAULT_TRANSPARENCY_URL,
  formatMonth,
  formatRupiah,
  setupTransparency,
  statTexts,
  type Transparency,
  transparencyUrl,
} from './transparency'

const report: Transparency = {
  deposit: 18_750_000,
  monthly_cost: 1_200_000,
  safe_until: '2028-03-14',
  updated_at: '2026-11-27T07:00:00Z',
}

describe('transparencyUrl', () => {
  it('defaults to the site origin', () => {
    expect(transparencyUrl({})).toBe(DEFAULT_TRANSPARENCY_URL)
    expect(transparencyUrl({ VITE_TRANSPARENCY_URL: ' https://api.x/t ' })).toBe('https://api.x/t')
  })
})

describe('formatting', () => {
  it('writes rupiah the short way', () => {
    expect(formatRupiah(18_750_000)).toBe('Rp18,75 jt')
    expect(formatRupiah(1_200_000)).toBe('Rp1,2 jt')
    expect(formatRupiah(2_500_000_000)).toBe('Rp2,5 M')
    expect(formatRupiah(750_000)).toBe('Rp750.000')
    expect(formatRupiah(-50_000)).toBe('−Rp50.000')
  })

  it('writes the month in Indonesian', () => {
    expect(formatMonth('2028-03-14')).toBe('Maret 2028')
  })

  it('fills every stat', () => {
    expect(statTexts(report)).toEqual({
      deposit: 'Rp18,75 jt',
      'monthly-cost': 'Rp1,2 jt',
      'safe-until': 'Maret 2028',
      'updated-at': '27 November 2026',
    })
    expect(statTexts({ ...report, safe_until: null })['safe-until']).toBe('Belum ada biaya')
  })
})

function page() {
  document.body.innerHTML = `
    <section data-transparency>
      <b data-stat="deposit">—</b>
      <b data-stat="safe-until">—</b>
      <p data-transparency-ready hidden>Diperbarui <span data-stat="updated-at">—</span>.</p>
      <p data-transparency-error hidden>Belum tersedia.</p>
    </section>`
  return document
}

describe('setupTransparency', () => {
  it('fills the numbers from the API', async () => {
    const doc = page()
    const send = vi.fn(async () => Response.json({ data: report }))
    await setupTransparency(doc, '/api/public/v1/transparency', send)

    expect(send).toHaveBeenCalledWith('/api/public/v1/transparency', expect.anything())
    expect(doc.querySelector('[data-stat="deposit"]')?.textContent).toBe('Rp18,75 jt')
    expect(doc.querySelector('[data-stat="safe-until"]')?.textContent).toBe('Maret 2028')
    expect(doc.querySelector('[data-stat="updated-at"]')?.textContent).toBe('27 November 2026')
    expect(doc.querySelector<HTMLElement>('[data-transparency-ready]')?.hidden).toBe(false)
    expect(doc.querySelector<HTMLElement>('[data-transparency-error]')?.hidden).toBe(true)
  })

  it('keeps the dashes and says so when the API is down', async () => {
    const doc = page()
    const send = vi.fn(async () => Response.json({ code: 'UNAVAILABLE' }, { status: 503 }))
    await setupTransparency(doc, '/x', send)

    expect(doc.querySelector('[data-stat="deposit"]')?.textContent).toBe('—')
    expect(doc.querySelector<HTMLElement>('[data-transparency-error]')?.hidden).toBe(false)
    expect(doc.querySelector<HTMLElement>('[data-transparency-ready]')?.hidden).toBe(true)
  })

  it('treats a network failure the same', async () => {
    const doc = page()
    await setupTransparency(
      doc,
      '/x',
      vi.fn(async () => Promise.reject(new TypeError('offline'))),
    )
    expect(doc.querySelector<HTMLElement>('[data-transparency-error]')?.hidden).toBe(false)
  })
})
