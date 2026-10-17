import { expect, test } from '@playwright/test'

for (const [name, route] of [['homepage', '/'], ['download', '/download/'], ['docs', '/docs/'], ['community', '/community/']] as const) {
  test(`${name} visual baseline`, async ({ page, browserName, isMobile }) => {
    test.skip(browserName !== 'chromium' || Boolean(isMobile), 'single deterministic visual project')
    await page.goto(route)
    await expect(page).toHaveScreenshot(`${name}.png`, { fullPage: true })
  })
}

for (const theme of ['dark', 'light'] as const) {
  test(`brand kit in ${theme} mode`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: theme })
    await page.goto('/brand/')
    await expect(page).toHaveScreenshot(`brand-${theme}.png`, { fullPage: true })
  })
}
