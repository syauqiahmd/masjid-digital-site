/**
 * The demo stage: the admin panel (a phone) and the TV side by side, so a visitor saves a change
 * on one and sees it on the other. Until `[data-demo-start]` is clicked the frames show still
 * screenshots (`[data-demo-placeholder]`): loading both apps for everyone who scrolls by would be
 * heavy, and every TV opened registers a new screen on the demo server.
 *
 * Markup: `[data-demo-stage]` holding `iframe[data-demo-frame]` (their URL in data-src, set by
 * applyDemoLinks), the placeholders, the start button and `[data-demo-running]` notes.
 */
export function setupDemoStage(root: ParentNode) {
  for (const stage of root.querySelectorAll<HTMLElement>('[data-demo-stage]')) {
    const start = stage.querySelector<HTMLButtonElement>('[data-demo-start]')
    start?.addEventListener('click', () => {
      for (const frame of stage.querySelectorAll<HTMLIFrameElement>('iframe[data-demo-frame]')) {
        if (!frame.getAttribute('src') && frame.dataset.src) frame.src = frame.dataset.src
        frame.hidden = false
      }
      for (const element of stage.querySelectorAll<HTMLElement>('[data-demo-placeholder]')) {
        element.hidden = true
      }
      for (const element of stage.querySelectorAll<HTMLElement>('[data-demo-running]')) {
        element.hidden = false
      }
      start.hidden = true
    })
  }
}

export type DemoAccountEnv = { VITE_DEMO_PHONE?: string; VITE_DEMO_PASSWORD?: string }

/** The demo's login, or null when either half is not set. */
export function demoAccount(env: DemoAccountEnv): { phone: string; password: string } | null {
  const phone = env.VITE_DEMO_PHONE?.trim()
  const password = env.VITE_DEMO_PASSWORD?.trim()
  return phone && password ? { phone, password } : null
}

/**
 * Shows the demo's login in `[data-demo-account]`, each value with a `[data-copy]` button: the
 * panel runs on the demo's own domain, so this page cannot type into its login form. Without an
 * account the box stays hidden.
 */
export function applyDemoAccount(
  root: ParentNode,
  account: { phone: string; password: string } | null,
  clipboard: Pick<Clipboard, 'writeText'> | undefined = navigator.clipboard,
) {
  const box = root.querySelector<HTMLElement>('[data-demo-account]')
  if (!box || !account) return

  const values: Record<string, string> = { phone: account.phone, password: account.password }
  for (const [name, value] of Object.entries(values)) {
    const element = box.querySelector<HTMLElement>(`[data-demo-account-${name}]`)
    if (element) element.textContent = value
  }

  for (const button of box.querySelectorAll<HTMLButtonElement>('[data-copy]')) {
    const value = values[button.dataset.copy ?? '']
    if (!value || !clipboard) {
      button.hidden = true
      continue
    }
    const label = button.textContent
    button.addEventListener('click', async () => {
      try {
        await clipboard.writeText(value)
        button.textContent = 'Tersalin'
      } catch {
        button.textContent = 'Gagal menyalin'
      }
      setTimeout(() => {
        button.textContent = label
      }, 2000)
    })
  }
  box.hidden = false
}
