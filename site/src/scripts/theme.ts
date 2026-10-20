const storageKey = 'envarium-site-theme'
const root = document.documentElement
let stored: string | null = null
try {
  stored = window.localStorage.getItem(storageKey)
} catch {
  // Browser storage can be disabled; the system theme still applies.
}

if (stored === 'dark' || stored === 'light') {
  root.dataset.theme = stored
}

document.querySelectorAll<HTMLButtonElement>('[data-theme-toggle]').forEach((button) => {
  button.addEventListener('click', () => {
    const current = root.dataset.theme
    const systemLight = window.matchMedia('(prefers-color-scheme: light)').matches
    const next = current === 'dark' || (!current && !systemLight) ? 'light' : 'dark'
    root.dataset.theme = next
    try {
      window.localStorage.setItem(storageKey, next)
    } catch {
      // Keep the selected theme for this page when persistence is unavailable.
    }
    button.setAttribute('aria-label', `Use ${next === 'dark' ? 'light' : 'dark'} theme`)
  })
})
