import { beforeEach, describe, expect, it } from 'vitest'
import { setupPlanPicker } from './plan'

beforeEach(() => {
  document.body.innerHTML = `
    <a href="#daftar" data-plan="donatur">Jadi donatur</a>
    <form data-inquiry>
      <input type="radio" name="plan" value="gratis" checked />
      <input type="radio" name="plan" value="donatur" />
      <fieldset data-plan-fields="gratis"><input name="subdomain" /></fieldset>
      <fieldset data-plan-fields="donatur"><input name="donation" required /></fieldset>
    </form>`
  setupPlanPicker(document)
})

const fields = (plan: string) =>
  document.querySelector<HTMLFieldSetElement>(`[data-plan-fields="${plan}"]`)!
const input = (name: string) => document.querySelector<HTMLInputElement>(`[name="${name}"]`)!
const type = (name: string, value: string) => {
  input(name).value = value
  input(name).dispatchEvent(new Event('input'))
}

describe('setupPlanPicker', () => {
  it('shows only the picked plan’s fields, and disables the other’s', () => {
    expect(fields('gratis').hidden).toBe(false)
    expect(fields('donatur').hidden).toBe(true)
    expect(fields('donatur').disabled).toBe(true)

    const donatur = document.querySelector<HTMLInputElement>('input[value="donatur"]')!
    donatur.checked = true
    donatur.dispatchEvent(new Event('change'))
    expect(fields('gratis').disabled).toBe(true)
    expect(fields('donatur').hidden).toBe(false)
  })

  it('picks the plan of a plan button', () => {
    document.querySelector<HTMLElement>('[data-plan="donatur"]')!.click()
    expect(document.querySelector<HTMLInputElement>('input[value="donatur"]')!.checked).toBe(true)
    expect(fields('donatur').disabled).toBe(false)
  })

  it('tidies the subdomain and the donation as they are typed', () => {
    type('subdomain', 'Al Ikhlas_1')
    expect(input('subdomain').value).toBe('alikhlas1')
    type('donation', 'Rp1500000')
    expect(input('donation').value).toBe('1.500.000')
  })
})
