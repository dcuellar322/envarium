import { mkdir, readdir, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { expect, test } from '@playwright/test'

async function htmlRoutes(directory: string, prefix = ''): Promise<string[]> {
  const entries = await readdir(directory, { withFileTypes: true })
  return (await Promise.all(entries.map(async (entry) => {
    const path = `${prefix}/${entry.name}`
    if (entry.isDirectory()) return htmlRoutes(resolve(directory, entry.name), path)
    return entry.name.endsWith('.html') ? [path.replace(/index\.html$/, '')] : []
  }))).flat().sort()
}

test('every built page renders, fits, and has working internal links', async ({ page, request, browserName, isMobile }) => {
  test.skip(browserName !== 'chromium', 'full crawl runs on desktop and mobile Chromium; journeys cover other engines')
  test.setTimeout(240_000)
  const routes = await htmlRoutes(resolve('dist'))
  const viewport = isMobile ? 'mobile' : 'desktop'
  const output = resolve('../output/playwright/site-audit', viewport)
  const capture = process.env.SITE_AUDIT_SCREENSHOTS === '1'
  if (capture) await mkdir(output, { recursive: true })
  const pages = new Map<string, Set<string>>()
  const links = new Map<string, string>()
  const problems: string[] = []
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(`${page.url()}: ${error.message}`))
  for (const route of routes) {
    const response = await page.goto(route)
    if (!response?.ok() && route !== '/404.html') problems.push(`${route}: HTTP ${response?.status()}`)
    const headings = await page.locator('h1').count()
    if (headings !== 1) problems.push(`${route}: expected one h1, found ${headings}`)
    for (const image of await page.locator('img[loading="lazy"]').all()) await image.scrollIntoViewIfNeeded()
    await page.evaluate(() => Promise.all(Array.from(document.images).map((img) => img.decode().catch(() => undefined))))
    const facts = await page.evaluate(() => ({
      width: document.documentElement.scrollWidth,
      viewport: innerWidth,
      images: Array.from(document.images).filter((img) => !img.complete || img.naturalWidth === 0).map((img) => img.src),
      ids: Array.from(document.querySelectorAll('[id]')).map((element) => element.id),
      links: Array.from(document.querySelectorAll<HTMLAnchorElement>('a[href]')).map((link) => link.href),
    }))
    pages.set(route, new Set(facts.ids))
    if (facts.width > facts.viewport) problems.push(`${route}: overflows ${facts.width - facts.viewport}px`)
    for (const image of facts.images) problems.push(`${route}: broken image ${image}`)
    for (const href of facts.links) {
      const url = new URL(href)
      if (url.origin === new URL(page.url()).origin && !links.has(url.href)) links.set(url.href, route)
    }
    if (capture) {
      await page.evaluate(() => scrollTo(0, 0))
      await page.screenshot({ path: resolve(output, `${route === '/' ? 'home' : route.replace(/^\//, '').replace(/\/$/, '').replaceAll('/', '__')}.png`), fullPage: true, animations: 'disabled' })
    }
  }
  const checked = new Set<string>()
  for (const [href, source] of links) {
    const url = new URL(href)
    if (!checked.has(url.pathname)) {
      const response = await request.get(url.pathname)
      if (!response.ok() && url.pathname !== '/404.html') problems.push(`${source}: ${url.pathname} returns ${response.status()}`)
      checked.add(url.pathname)
    }
    if (url.hash && pages.has(url.pathname) && !pages.get(url.pathname)!.has(decodeURIComponent(url.hash.slice(1)))) {
      problems.push(`${source}: missing anchor ${url.pathname}${url.hash}`)
    }
  }
  if (capture) await writeFile(resolve(output, 'report.json'), JSON.stringify({ pages: routes.length, internalTargets: checked.size, problems, errors }, null, 2))
  expect([...problems, ...errors]).toEqual([])
})
