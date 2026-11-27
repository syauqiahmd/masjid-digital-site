import './styles.css'
import { applyDemoAccount, demoAccount, setupDemoStage } from './demo'
import { applyDemoLinks, demoUrls } from './demoLinks'
import { inquiryUrl, setupInquiryForm } from './inquiry'
import { setupPlanPicker } from './plan'
import { setupTabs } from './tabs'
import { setupTransparency, transparencyUrl } from './transparency'

// demo links first: the demo stage loads its frames from the URLs they set
applyDemoLinks(document, demoUrls(import.meta.env))
setupTabs(document)
setupDemoStage(document)
applyDemoAccount(document, demoAccount(import.meta.env))
setupInquiryForm(document, inquiryUrl(import.meta.env))
setupPlanPicker(document)
void setupTransparency(document, transparencyUrl(import.meta.env))

// The phone menu: one button opens and closes the nav, and a picked link closes it.
const toggle = document.querySelector<HTMLButtonElement>('[data-menu-toggle]')
const menu = document.getElementById(toggle?.getAttribute('aria-controls') ?? '')
const setMenu = (open: boolean) => {
  toggle?.setAttribute('aria-expanded', String(open))
  menu?.classList.toggle('max-lg:hidden', !open)
}
toggle?.addEventListener('click', () => setMenu(toggle.getAttribute('aria-expanded') !== 'true'))
menu?.addEventListener('click', (e) => {
  if ((e.target as Element).closest('a')) setMenu(false)
})

for (const year of document.querySelectorAll('[data-year]')) {
  year.textContent = String(new Date().getFullYear())
}
