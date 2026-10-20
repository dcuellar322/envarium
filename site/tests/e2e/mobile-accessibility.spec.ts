import { expect, test } from '@playwright/test'
import { features } from '../../src/content/marketing/features'
import { publicRoutes } from '../public-routes'

test('mobile navigation supports touch, Escape, outside clicks and focus dismissal', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 667 })
  await page.goto('/')
  const menu = page.locator('details.mobile-menu')
  const toggle = menu.locator('summary')
  const size = await toggle.boundingBox()
  expect(size?.width).toBeGreaterThanOrEqual(44)
  expect(size?.height).toBeGreaterThanOrEqual(44)
  await toggle.click()
  await page.keyboard.press('Escape')
  await expect(menu).not.toHaveAttribute('open', '')
  await expect(toggle).toBeFocused()
  await toggle.click()
  await page.getByRole('heading', { level: 1 }).click({ position: { x: 2, y: 2 } })
  await expect(menu).not.toHaveAttribute('open', '')
  await toggle.click()
  await menu.getByRole('link', { name: 'GitHub', exact: true }).focus()
  await page.keyboard.press('Tab')
  await expect(menu).not.toHaveAttribute('open', '')
  await page.goto('/community/')
  await toggle.click()
  await expect(menu.getByRole('link', { name: 'Community', exact: true })).toHaveAttribute('aria-current', 'page')
  await page.setViewportSize({ width: 1280, height: 800 })
  await expect(menu).not.toHaveAttribute('open', '')
  await page.setViewportSize({ width: 375, height: 667 })
  await expect(menu).not.toHaveAttribute('open', '')
})

test('landscape mobile navigation keeps the final link reachable in the viewport', async ({ page }) => {
  await page.setViewportSize({ width: 667, height: 320 })
  await page.goto('/')
  const menu = page.locator('details.mobile-menu')
  await menu.locator('summary').click()
  const last = menu.getByRole('link', { name: 'GitHub', exact: true })
  await last.scrollIntoViewIfNeeded()
  const bounds = await last.boundingBox()
  expect(bounds!.y).toBeGreaterThanOrEqual(0)
  expect(bounds!.y + bounds!.height).toBeLessThanOrEqual(320)
})

test('mobile navigation remains usable without JavaScript', async ({ browser }) => {
  const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 320, height: 568 } })
  try {
    const page = await context.newPage()
    await page.goto('/')
    const menu = page.locator('details.mobile-menu')
    await menu.locator('summary').click()
    await menu.getByRole('link', { name: 'Community' }).click()
    await expect(page).toHaveURL(/\/community\/$/)
  } finally { await context.close() }
})

test('skip navigation places keyboard focus in the main content', async ({ page, browserName }) => {
  await page.goto('/')
  // macOS WebKit uses Option-Tab to include links in keyboard navigation.
  await page.keyboard.press(browserName === 'webkit' && process.platform === 'darwin' ? 'Alt+Tab' : 'Tab')
  await expect(page.getByRole('link', { name: 'Skip to content' })).toBeFocused()
  await page.keyboard.press('Enter')
  await expect(page.locator('main')).toBeFocused()
})

for (const blocked of ['read', 'write'] as const) {
  test(`theme control works when storage ${blocked} is blocked`, async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 })
    await page.addInitScript((mode) => {
      if (mode === 'read') Object.defineProperty(window, 'localStorage', { get() { throw new DOMException('Blocked', 'SecurityError') } })
      else Storage.prototype.setItem = () => { throw new DOMException('Blocked', 'SecurityError') }
    }, blocked)
    const errors: string[] = []
    page.on('pageerror', (error) => errors.push(error.message))
    await page.goto('/')
    const toggle = page.locator('[data-theme-toggle]')
    await toggle.click()
    const theme = await page.locator('html').getAttribute('data-theme')
    expect(theme).toMatch(/dark|light/)
    await toggle.click()
    await expect(page.locator('html')).toHaveAttribute('data-theme', theme === 'dark' ? 'light' : 'dark')
    expect(errors).toEqual([])
  })
}

for (const route of ['/', '/download/']) {
  test(`${route} keeps its layout when scripts arrive late`, async ({ page }) => {
    await page.setViewportSize({ width: 412, height: 823 })
    let release!: () => void
    const scripts = new Promise<void>((resolve) => { release = resolve })
    const inlineModules = new Map<string, string>()
    await page.route('**/__audit-module-*.js', async (request) => {
      await scripts
      await request.fulfill({ contentType: 'text/javascript', body: inlineModules.get(new URL(request.request().url()).pathname)! })
    })
    await page.route('**/_astro/*.js', async (request) => { await scripts; await request.continue() })
    await page.route((url) => url.pathname === route, async (request) => {
      const response = await request.fetch()
      // Astro may inline small modules. Serve their unchanged code as delayed
      // same-origin modules so this check exercises both build strategies.
      const html = (await response.text()).replace(/<script([^>]*type="module"[^>]*)>([\s\S]*?)<\/script>/g, (tag, attributes: string, code: string) => {
        if (/\bsrc=/.test(attributes) || !code) return tag
        const path = `/__audit-module-${inlineModules.size}.js`
        inlineModules.set(path, code)
        return `<script${attributes} src="${path}"></script>`
      })
      await request.fulfill({ response, body: html })
    })
    try {
      await page.goto(route, { waitUntil: 'commit' })
      const target = page.locator(route === '/' ? '.product-frame' : '[data-download-result]')
      if (route === '/download/') await expect(page.locator('[data-download-selector]')).toHaveCSS('display', 'grid')
      else await expect(page.locator('.hero')).toHaveCSS('display', 'grid')
      const before = await target.boundingBox()
      release()
      await page.waitForLoadState('networkidle')
      const after = await target.boundingBox()
      expect(Math.abs(after!.y - before!.y)).toBeLessThanOrEqual(1)
    } finally { release(); await page.unrouteAll({ behavior: 'wait' }) }
  })
}

test('long verification commands scroll with the keyboard on phones', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 568 })
  await page.goto('/download/verify/')
  const command = page.locator('pre').nth(1)
  await expect(command).toHaveAttribute('tabindex', '0')
  await command.focus()
  await page.keyboard.press('ArrowRight')
  await expect.poll(() => command.evaluate((element) => element.scrollLeft)).toBeGreaterThan(0)
})

test('documentation code and tables scroll with the keyboard without JavaScript', async ({ browser }) => {
  const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 320, height: 568 } })
  try {
    const page = await context.newPage()
    await page.goto('/docs/mcp/')
    for (const selector of ['.sl-markdown-content pre', '.sl-markdown-content table']) {
      const content = page.locator(selector).first()
      await expect(content).toHaveAttribute('tabindex', '0')
      await content.focus()
      // Native WebKit scrolling settles after focus; repeated key presses
      // exercise keyboard scrolling without depending on that animation time.
      await expect.poll(async () => {
        await page.keyboard.press('ArrowRight')
        return content.evaluate((element) => element.scrollLeft)
      }).toBeGreaterThan(0)
    }
  } finally { await context.close() }
})

test('feature screenshots preserve their natural aspect ratio on phones', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 667 })
  for (const feature of features) {
    await page.goto(`/features/${feature.slug}/`)
    const image = page.locator('.product-shot img')
    await image.scrollIntoViewIfNeeded()
    const ratios = await image.evaluate(async (element) => {
      const image = element as HTMLImageElement
      await image.decode()
      const bounds = image.getBoundingClientRect()
      return { natural: image.naturalWidth / image.naturalHeight, rendered: bounds.width / bounds.height }
    })
    expect(ratios.rendered, feature.slug).toBeCloseTo(ratios.natural, 2)
  }
})

test('every public page fits a 320px viewport', async ({ page, browserName, isMobile }) => {
  test.skip(browserName !== 'chromium' || isMobile, 'one narrow crawl complements all-engine mobile journeys')
  test.setTimeout(180_000)
  await page.setViewportSize({ width: 320, height: 568 })
  const failures: string[] = []
  for (const route of await publicRoutes()) {
    await page.goto(route)
    if (await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)) failures.push(route)
  }
  expect(failures).toEqual([])
})

test('structured metadata does not advertise coming-soon platforms or an unpublished version', async ({ page }) => {
  await page.goto('/')
  const data = JSON.parse(await page.locator('script[type="application/ld+json"]').innerText()) as Array<{ '@type': string; operatingSystem?: string; softwareVersion?: string }>
  const app = data.find((entry) => entry['@type'] === 'SoftwareApplication')!
  expect(app.operatingSystem).not.toMatch(/Windows|WSL/)
  if (await page.getByText('No stable binaries yet', { exact: false }).count()) expect(app.softwareVersion).toBeUndefined()
})
