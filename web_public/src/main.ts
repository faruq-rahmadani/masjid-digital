import './styles.css'
import { applyDemoLinks, demoUrls } from './demoLinks'
import { setupTabs } from './tabs'

// demo links first: the tabs load the visible demo frame from the URLs they set
applyDemoLinks(document, demoUrls(import.meta.env))
setupTabs(document)

// The phone menu: one button opens and closes the nav, and a picked link closes it.
const toggle = document.querySelector<HTMLButtonElement>('[data-menu-toggle]')
const menu = document.getElementById(toggle?.getAttribute('aria-controls') ?? '')
const setMenu = (open: boolean) => {
  toggle?.setAttribute('aria-expanded', String(open))
  menu?.classList.toggle('max-md:hidden', !open)
}
toggle?.addEventListener('click', () => setMenu(toggle.getAttribute('aria-expanded') !== 'true'))
menu?.addEventListener('click', (e) => {
  if ((e.target as Element).closest('a')) setMenu(false)
})

for (const year of document.querySelectorAll('[data-year]')) {
  year.textContent = String(new Date().getFullYear())
}
