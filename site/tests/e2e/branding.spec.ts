import { expect, test } from '@playwright/test'

for (const theme of ['dark', 'light'] as const) {
  test(`the identity and palette stay consistent in ${theme} mode`, async ({ page }) => {
    await page.emulateMedia({ colorScheme: theme })
    for (const route of ['/', '/features/', '/brand/', '/docs/', '/docs/getting-started/']) {
      await page.goto(route)
      const home = page.getByRole('link', { name: route.startsWith('/docs/') ? 'Envarium documentation' : 'Envarium home', exact: true }).first()
      await expect(home).toBeVisible()
      await expect(home.locator(`img.on-${theme === 'dark' ? 'dark' : 'light'}`)).toBeVisible()
      const colors = await page.locator('html').evaluate((element) => {
        const style = getComputedStyle(element)
        return ['blue', 'violet', 'canvas', 'panel', 'text'].map((name) => style.getPropertyValue(`--brand-${name}`).trim())
      })
      expect(colors).toEqual(['#78a6ff', '#9e7bff', '#0a0d12', '#11161e', '#edf3fb'])
      expect(await home.locator('img').evaluateAll((images) => images.every((image) => (image as HTMLImageElement).complete && (image as HTMLImageElement).naturalWidth > 0))).toBe(true)
    }
    await page.goto('/')
    const primary = page.locator('.button.primary').first()
    const style = await primary.evaluate((element) => ({ background: getComputedStyle(element).backgroundImage, color: getComputedStyle(element).color }))
    expect(style.background).toContain('rgb(120, 166, 255)')
    expect(style.background).toContain('rgb(158, 123, 255)')
    expect(style.color).toBe('rgb(10, 13, 18)')
  })
}

test('brand downloads and install icons resolve to real assets', async ({ page, request }) => {
  await page.goto('/brand/')
  const download = page.getByRole('link', { name: /Download the complete brand kit/ })
  await expect(download).toHaveAttribute('download', '')
  const archive = await request.get((await download.getAttribute('href'))!)
  expect(archive.ok()).toBe(true)
  expect((await archive.body()).subarray(0, 2).toString()).toBe('PK')
  await expect(page.locator('.interface-grid a')).toHaveCount(12)
  for (const href of await page.locator('a[download]').evaluateAll((links) => links.map((link) => link.getAttribute('href')!))) {
    expect((await request.get(href)).ok(), href).toBe(true)
  }
  const manifest = await (await request.get('/site.webmanifest')).json() as { icons: Array<{ src: string }> }
  for (const icon of manifest.icons) expect((await request.get(icon.src)).ok()).toBe(true)
  const sprite = await request.get('/brand/interface.svg')
  expect(sprite.ok()).toBe(true)
  expect(await sprite.text()).toContain('id="projects"')
  expect((await request.get('/social/default.png')).headers()['content-type']).toContain('image/png')
})

for (const width of [320, 375]) {
  test(`branding fits a ${width}px screen`, async ({ page }) => {
    await page.setViewportSize({ width, height: 812 })
    for (const route of ['/', '/brand/', '/docs/']) {
      await page.goto(route)
      const dimensions = await page.evaluate(() => ({ content: document.documentElement.scrollWidth, viewport: window.innerWidth }))
      expect(dimensions.content, `${route} at ${width}px`).toBeLessThanOrEqual(dimensions.viewport)
    }
  })
}
