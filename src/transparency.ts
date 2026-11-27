export type TransparencyEnv = { VITE_TRANSPARENCY_URL?: string }

/** The API's route, on the site's own origin unless VITE_TRANSPARENCY_URL says otherwise. */
export const DEFAULT_TRANSPARENCY_URL = '/api/public/v1/transparency'

export function transparencyUrl(env: TransparencyEnv): string {
  return env.VITE_TRANSPARENCY_URL?.trim() || DEFAULT_TRANSPARENCY_URL
}

/** GET /api/public/v1/transparency's `data`: whole rupiah, and the day the deposit runs out. */
export type Transparency = {
  deposit: number
  monthly_cost: number
  safe_until: string | null
  updated_at: string
}

const decimal = (n: number, digits: number) =>
  n.toLocaleString('id-ID', { maximumFractionDigits: digits })

/** Rupiah the short way, as people say it: Rp18,75 jt, Rp1,2 M, Rp750.000. */
export function formatRupiah(amount: number): string {
  const sign = amount < 0 ? '−' : ''
  const n = Math.abs(amount)
  if (n >= 1e9) return `${sign}Rp${decimal(n / 1e9, 2)} M`
  if (n >= 1e6) return `${sign}Rp${decimal(n / 1e6, 2)} jt`
  return `${sign}Rp${decimal(n, 0)}`
}

/** "2028-03-14" as "Maret 2028". */
export function formatMonth(date: string): string {
  const [year, month] = date.split('-').map(Number)
  if (!year || !month) return date
  return new Date(Date.UTC(year, month - 1, 1)).toLocaleDateString('id-ID', {
    month: 'long',
    year: 'numeric',
    timeZone: 'UTC',
  })
}

/** "2026-11-27T07:00:00Z" as "27 November 2026", in WIB like the team reads it. */
export function formatDay(timestamp: string): string {
  const date = new Date(timestamp)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    timeZone: 'Asia/Jakarta',
  })
}

/** The text of each `[data-stat]` element, by its value. */
export function statTexts(report: Transparency): Record<string, string> {
  return {
    deposit: formatRupiah(report.deposit),
    'monthly-cost': formatRupiah(report.monthly_cost),
    'safe-until': report.safe_until ? formatMonth(report.safe_until) : 'Belum ada biaya',
    'updated-at': formatDay(report.updated_at),
  }
}

/**
 * Fills `[data-transparency]` from the API: each `[data-stat]` gets its number, and
 * `[data-transparency-ready]` / `[data-transparency-error]` show which way it went. Until then
 * (and without JavaScript) the page shows its placeholder dashes.
 */
export async function setupTransparency(
  root: ParentNode,
  url: string,
  send: typeof fetch = fetch,
): Promise<void> {
  const section = root.querySelector<HTMLElement>('[data-transparency]')
  if (!section) return

  const show = (selector: string) => {
    const element = section.querySelector<HTMLElement>(selector)
    if (element) element.hidden = false
  }

  try {
    const res = await send(url, { headers: { Accept: 'application/json' } })
    if (!res.ok) throw new Error(`status ${res.status}`)
    const body: unknown = await res.json()
    const report = (body as { data?: Transparency } | null)?.data
    if (!report || typeof report.deposit !== 'number') throw new Error('unexpected body')

    const texts = statTexts(report)
    for (const element of section.querySelectorAll<HTMLElement>('[data-stat]')) {
      const text = texts[element.dataset.stat ?? '']
      if (text !== undefined) element.textContent = text
    }
    show('[data-transparency-ready]')
  } catch {
    show('[data-transparency-error]')
  }
}
