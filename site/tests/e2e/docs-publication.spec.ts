import { readdir } from 'node:fs/promises'
import { expect, test } from '@playwright/test'
import { internalDocumentation, isPublicDocumentation } from '../../src/lib/docs/publication'

const internalRoutes = (await readdir('../docs', { recursive: true }))
  .filter((path) => /\.mdx?$/.test(path) && !isPublicDocumentation(path))
  .map((path) => `/docs/${path.replace(/\.mdx?$/, '')}`)

function isInternalRoute(path: string): boolean {
  return internalDocumentation.some((entry) => path.startsWith(`/docs/${entry.replace(/\.mdx?$/, '')}`))
}

test('maintainer records have no published routes, sitemap entries, or AI exports', async ({ request, browserName, isMobile }) => {
  test.skip(browserName !== 'chromium' || isMobile, 'one HTTP check covers the shared static output')
  test.setTimeout(120_000)
  for (const route of internalRoutes) {
    for (const suffix of ['/', '.md']) {
      const response = await request.get(`${route}${suffix}`)
      expect(response.status(), `${route}${suffix}`).toBe(404)
    }
  }
  const index = await request.get('/sitemap-index.xml')
  expect(index.ok()).toBe(true)
  const sitemapUrls = [...(await index.text()).matchAll(/<loc>([^<]+)<\/loc>/g)].map((match) => match[1]!)
  const sitemaps = await Promise.all(sitemapUrls.map(async (url) => {
    const response = await request.get(new URL(url).pathname)
    expect(response.ok()).toBe(true)
    return response.text()
  }))
  for (const name of ['llms.txt', 'llms-full.txt']) {
    const response = await request.get(`/${name}`)
    expect(response.ok()).toBe(true)
    const content = await response.text()
    expect(content).toContain('Envarium CLI contract')
    expect(content, name).not.toMatch(/\bphases?\b/i)
    for (const route of internalRoutes) expect(content, name).not.toContain(route)
  }
  for (const route of internalRoutes) expect(sitemaps.join('\n')).not.toContain(`${route}/`)
})

test('search indexes public guides without internal phase reports or maintainer logs', async ({ page }) => {
  await page.goto('/docs/')
  const results = await page.evaluate(async () => {
    // Exercise the deployed index directly, including all results, rather than
    // checking only the first visible result in the search dialog.
    const modulePath = '/pagefind/pagefind.js'
    const pagefind = await import(/* @vite-ignore */ modulePath)
    const searches = await Promise.all(['Envarium CLI contract', 'Phase', 'acceptance', 'launch checklist', 'rebrand', 'Arcane', '"Copy-ready Markdown"'].map(async (query) => {
      const search = await pagefind.search(query)
      return { query, urls: (await Promise.all(search.results.map((result: { data: () => Promise<{ url: string }> }) => result.data()))).map((result: { url: string }) => result.url) }
    }))
    return searches
  })
  expect(results[0]?.urls.some((url: string) => new URL(url, page.url()).pathname === '/docs/cli/')).toBe(true)
  expect(results.find((result) => result.query === '"Copy-ready Markdown"')?.urls).toEqual([])
  for (const result of results) {
    for (const url of result.urls) expect(isInternalRoute(new URL(url, page.url()).pathname), `${result.query}: ${url}`).toBe(false)
  }
  await page.getByRole('button', { name: 'Search', exact: true }).click()
  await page.getByRole('textbox', { name: 'Search' }).fill('Envarium CLI contract')
  await page.getByRole('link', { name: /Envarium CLI contract/ }).first().click()
  await expect(page).toHaveURL(/\/docs\/cli\/$/)
})
