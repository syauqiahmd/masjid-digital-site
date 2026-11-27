/**
 * The Daftar form's plan picker: the radios `input[name="plan"]` show the fields of the plan
 * picked (`fieldset[data-plan-fields="gratis" | "donatur"]`) and disable the other plan's, so the
 * browser neither checks nor sends them. The plan buttons elsewhere on the page (`[data-plan]`)
 * pick their plan on the way to the form.
 *
 * It also tidies two inputs as they are typed: the subdomain to what a DNS label allows, and the
 * donation to rupiah with thousands dots (1.500.000).
 */
export function setupPlanPicker(root: ParentNode) {
  const form = root.querySelector<HTMLFormElement>('form[data-inquiry]')
  if (!form) return
  const radios = [...form.querySelectorAll<HTMLInputElement>('input[name="plan"]')]

  const sync = () => {
    const picked = radios.find((radio) => radio.checked)?.value
    for (const fields of form.querySelectorAll<HTMLFieldSetElement>('[data-plan-fields]')) {
      const shown = fields.dataset.planFields === picked
      fields.hidden = !shown
      fields.disabled = !shown
    }
  }
  for (const radio of radios) radio.addEventListener('change', sync)
  form.addEventListener('reset', () => setTimeout(sync))

  for (const link of root.querySelectorAll<HTMLElement>('[data-plan]')) {
    link.addEventListener('click', () => {
      const radio = radios.find((r) => r.value === link.dataset.plan)
      if (!radio) return
      radio.checked = true
      sync()
    })
  }

  form
    .querySelector<HTMLInputElement>('input[name="subdomain"]')
    ?.addEventListener('input', (e) => {
      const input = e.target as HTMLInputElement
      input.value = input.value.toLowerCase().replace(/[^a-z0-9-]/g, '')
    })
  form.querySelector<HTMLInputElement>('input[name="donation"]')?.addEventListener('input', (e) => {
    const input = e.target as HTMLInputElement
    const amount = Number(input.value.replace(/\D/g, ''))
    input.value = amount ? amount.toLocaleString('id-ID') : ''
  })

  sync()
}
