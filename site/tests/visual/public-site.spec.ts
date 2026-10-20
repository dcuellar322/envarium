import { expect, test } from '@playwright/test'

for (const [name, route] of [['homepage', '/'], ['download', '/download/'], ['docs', '/docs/'], ['community', '/community/'], ['projects-feature', '/features/projects/']] as const) {
  test(`${name} visual baseline`, async ({ page, browserName, isMobile }) => {
    test.skip(browserName !== 'chromium' || Boolean(isMobile), 'single deterministic visual project')
    expect((await page.goto(route))?.ok()).toBe(true)
    await expect(page).toHaveTitle(/Envarium/)
    await expect(page).toHaveScreenshot(`${name}.png`, { fullPage: true })
  })
}

for (const theme of ['dark', 'light'] as const) {
  test(`brand kit in ${theme} mode`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: theme })
    expect((await page.goto('/brand/'))?.ok()).toBe(true)
    await expect(page).toHaveTitle(/Envarium/)
    await expect(page).toHaveScreenshot(`brand-${theme}.png`, { fullPage: true })
  })
}
