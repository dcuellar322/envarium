import AxeBuilder from '@axe-core/playwright'
import { expect, test } from '@playwright/test'
import { publicRoutes } from '../public-routes'

// Axe opens temporary pages; static audit traces and JSON cover failures
// without recording a separate video for every scanned document.
test.use({ video: 'off' })

const tags = ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa']
const templates = ['/', '/download/', '/download/cli/', '/download/verify/', '/features/projects/', '/docs/', '/docs/cli/', '/docs/start/native-process/', '/community/', '/contributing/', '/integrations/codex/', '/roadmap/', '/brand/', '/about/', '/security/', '/privacy/']

for (const theme of ['dark', 'light'] as const) {
  for (const route of templates) {
    test(`${route} has no axe violations in ${theme} mode`, async ({ page, browserName }) => {
      test.skip(browserName !== 'chromium', 'axe runs once; behavioral journeys cover the cross-browser matrix')
      await page.emulateMedia({ colorScheme: theme })
      await page.goto(route)
      const { violations } = await new AxeBuilder({ page }).withTags(tags).analyze()
      expect(violations, JSON.stringify(violations, null, 2)).toEqual([])
    })
  }
}

test('every public page passes mobile axe checks', async ({ page, browserName, isMobile }, testInfo) => {
  test.skip(browserName !== 'chromium' || !isMobile, 'one complete mobile crawl complements themed template checks')
  test.setTimeout(420_000)
  await page.emulateMedia({ colorScheme: 'light' })
  const results = []
  for (const route of await publicRoutes()) {
    await page.goto(route)
    const { violations } = await new AxeBuilder({ page }).withTags(tags).analyze()
    results.push({ route, violations })
  }
  await testInfo.attach('public-page-accessibility.json', { body: JSON.stringify(results, null, 2), contentType: 'application/json' })
  const failures = results.filter(({ violations }) => violations.length)
  expect(failures, JSON.stringify(failures, null, 2)).toEqual([])
})

test('expanded mobile navigation passes axe checks', async ({ page, browserName, isMobile }) => {
  test.skip(browserName !== 'chromium' || !isMobile, 'mobile navigation state')
  await page.goto('/')
  await page.locator('details.mobile-menu summary').click()
  const { violations } = await new AxeBuilder({ page }).withTags(tags).analyze()
  expect(violations, JSON.stringify(violations, null, 2)).toEqual([])
})

test('documentation search results pass axe checks', async ({ page, browserName }) => {
  test.skip(browserName !== 'chromium', 'desktop and mobile search states')
  await page.goto('/docs/?search')
  await page.getByRole('textbox', { name: 'Search' }).fill('Envarium CLI contract')
  await expect(page.getByRole('link', { name: /Envarium CLI contract/ }).first()).toBeVisible()
  const { violations } = await new AxeBuilder({ page }).withTags(tags).analyze()
  expect(violations, JSON.stringify(violations, null, 2)).toEqual([])
})
