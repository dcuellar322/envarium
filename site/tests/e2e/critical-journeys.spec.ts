import { expect, test } from '@playwright/test'

test('homepage routes to trusted downloads and the first-project guide', async ({ page }) => {
  await page.goto('/')
  await expect(page).toHaveTitle(/Envarium — Local Development Command Center/)
  await expect(page.getByRole('heading', { name: 'Your local development command center.' })).toBeVisible()
  await page.getByRole('link', { name: /Download/ }).first().click()
  await expect(page).toHaveURL(/\/download\/$/)
  await expect(page.getByText('No published stable package for this selection.')).toBeVisible()
  await page.goto('/')
  await page.getByRole('link', { name: 'First-project guide' }).click()
  await expect(page).toHaveURL(/\/docs\/start\/$/)
})

test('download selection keeps every platform visible and links integrity evidence', async ({ page }) => {
  await page.goto('/download/')
  await page.locator('input[name="platform"][value="linux"]').check()
  await page.locator('input[name="architecture"][value="amd64"]').check()
  await page.getByRole('button', { name: 'Update recommendation' }).click()
  await expect(page).toHaveURL(/platform=linux/)
  await expect(page).toHaveURL(/architecture=amd64/)
  await expect(page.locator('[data-download-result] h2').first()).toHaveText('Linux · amd64 · stable')
  const platformGrid = page.locator('.platform-grid')
  for (const platform of ['macOS', 'Windows', 'Linux', 'WSL2', 'Headless CLI']) {
    await expect(platformGrid.getByRole('link', { name: new RegExp(`^${platform}`) })).toBeVisible()
  }
  await page.getByRole('link', { name: 'Verification guide' }).click()
  await expect(page.getByRole('heading', { name: 'Verify an Envarium download' })).toBeVisible()
})

test('documentation starts at a tested first-project outcome', async ({ page }) => {
  await page.goto('/docs/')
  await expect(page.getByRole('heading', { level: 1, name: 'Documentation' })).toBeVisible()
  await page.getByRole('link', { name: /Add your first project/ }).first().click()
  await expect(page).toHaveURL(/\/docs\/getting-started\/$/)
  await expect(page.getByText('envarium project add')).toBeVisible()
})

test('documentation search finds the CLI contract under the production CSP', async ({ page }) => {
  const response = await page.goto('/docs/')
  const policy = response?.headers()['content-security-policy'] ?? ''
  expect(policy).toContain("'wasm-unsafe-eval'")
  expect(policy).not.toContain("'unsafe-eval'")
  const openSearch = page.getByRole('button', { name: 'Search' })
  await expect(openSearch).toBeEnabled()
  await openSearch.click()
  const search = page.getByRole('textbox', { name: 'Search' })
  await search.fill('Envarium CLI contract')
  await page.getByRole('link', { name: /Envarium CLI contract/ }).first().click()
  await expect(page).toHaveURL(/\/docs\/cli\/$/)
})

test('Codex integration documents setup and authorization failures', async ({ page }) => {
  await page.goto('/integrations/codex/')
  await expect(page.getByRole('heading', { level: 1, name: /Codex/ })).toBeVisible()
  await expect(page.getByText('envarium agent install codex --profile observe')).toBeVisible()
  await expect(page.getByText(/authorization/i).first()).toBeVisible()
  await expect(page.getByText(/remov/i).first()).toBeVisible()
})

test('generic MCP setup provides a read-only stdio configuration and reference', async ({ page }) => {
  await page.goto('/integrations/mcp/')
  const configuration = JSON.parse(await page.locator('[data-mcp-config]').innerText()) as { type: string; command: string; args: string[] }
  expect(configuration.type).toBe('stdio')
  expect(configuration.command).toMatch(/^\//)
  expect(configuration.args.slice(0, 2)).toEqual(['mcp', 'serve'])
  expect(configuration.args[configuration.args.indexOf('--profile') + 1]).toBe('observe')
  await page.getByRole('link', { name: 'Open reference', exact: true }).click()
  await expect(page).toHaveURL(/\/docs\/mcp\/$/)
  await expect(page.getByRole('heading', { name: 'Install', exact: true })).toBeVisible()
})

test('unpublished changelog offers source builds and an empty release feed', async ({ page, request }) => {
  await page.goto('/changelog/')
  await expect(page.getByRole('heading', { name: 'No public releases yet' })).toBeVisible()
  await page.getByRole('link', { name: 'Desktop build instructions' }).click()
  await expect(page).toHaveURL(/\/docs\/desktop-installation\/#build-from-source$/)
  const feed = await request.get('/changelog/rss.xml')
  expect(feed.ok()).toBe(true)
  const xml = await feed.text()
  expect(xml).toContain('<rss')
  expect(xml).not.toContain('<item>')
})

test('community routes questions, bugs, proposals, and security correctly', async ({ page }) => {
  await page.goto('/community/')
  await expect(page.getByRole('link', { name: /Ask a question/ })).toHaveAttribute('href', /question\.yml/)
  await expect(page.getByRole('link', { name: /Open a bug report/ })).toHaveAttribute('href', /bug_report\.yml/)
  await expect(page.getByRole('link', { name: /Propose a feature/ })).toHaveAttribute('href', /feature_request\.yml/)
  await expect(page.getByRole('link', { name: /Report privately/ })).toHaveAttribute('href', /security\/advisories\/new/)
  await expect(page.getByText('Never post credentials')).toBeVisible()
})

test('mobile navigation exposes primary destinations', async ({ page, isMobile }) => {
  test.skip(!isMobile, 'mobile-only journey')
  await page.goto('/')
  const menu = page.locator('details.mobile-menu')
  await menu.locator('summary').click()
  await expect(menu.getByRole('link', { name: 'Community' })).toBeVisible()
})

test('theme selection persists across marketing pages', async ({ page, isMobile }) => {
  test.skip(Boolean(isMobile), 'theme control is intentionally hidden on narrow screens')
  await page.goto('/')
  const toggle = page.getByRole('button', { name: 'Change color theme' })
  await toggle.click()
  const theme = await page.locator('html').getAttribute('data-theme')
  expect(theme).toMatch(/dark|light/)
  await page.goto('/features/')
  await expect(page.locator('html')).toHaveAttribute('data-theme', theme!)
})


test('download deep links, channels, invalid values, and history match the result', async ({ page }) => {
  await page.goto('/download/?platform=linux&architecture=amd64&channel=beta')
  const result = page.locator('[data-download-result]')
  await expect(result.getByRole('heading', { name: 'Linux · amd64 · beta' })).toBeVisible()
  await expect(result.getByText('No published beta package for this selection.')).toBeVisible()
  await page.getByRole('combobox', { name: 'Release channel' }).selectOption('nightly')
  await page.getByRole('button', { name: 'Update recommendation' }).click()
  await expect(result.getByRole('heading', { name: 'Linux · amd64 · nightly' })).toBeVisible()
  await page.goBack()
  await expect(result.getByRole('heading', { name: 'Linux · amd64 · beta' })).toBeVisible()
  await page.goto('/download/?platform=invalid&architecture=invalid&channel=invalid')
  await expect(result.getByRole('heading', { name: 'macOS · arm64 · stable' })).toBeVisible()
})

test('Windows and WSL2 are coming soon throughout installation', async ({ page }) => {
  for (const platform of ['windows', 'wsl']) {
    await page.goto(`/download/?platform=${platform}`)
    const result = page.locator('[data-download-result]')
    await expect(result.getByRole('heading').first()).toContainText('coming soon')
    await expect(result.locator('a[href*="releases/download/"]')).toHaveCount(0)
    await page.goto(`/download/${platform}/`)
    await expect(page.getByRole('heading', { level: 1 })).toContainText('coming soon')
    await expect(page.locator('a[href*="releases/download/"]')).toHaveCount(0)
  }
})

test('unpublished desktop packages lead to usable source build instructions', async ({ page }) => {
  for (const platform of ['macos', 'linux']) {
    await page.goto(`/download/${platform}/`)
    await expect(page.locator('a[href*="releases/download/"]')).toHaveCount(0)
    await expect(page.locator('aside ol')).toHaveCount(0)
    await page.getByRole('link', { name: 'CLI build instructions' }).click()
    await expect(page).toHaveURL(/\/download\/cli\/#build-from-source$/)
    await expect(page.getByText('make build', { exact: false })).toBeVisible()
    await page.goto(`/download/${platform}/`)
    await page.getByRole('link', { name: 'Desktop build instructions' }).click()
    await expect(page).toHaveURL(/\/docs\/desktop-installation\/#build-from-source$/)
    await expect(page.getByText('make desktop-build', { exact: false })).toBeVisible()
  }
})

test('search navigation opens search, handles no matches, and closes with Escape', async ({ page, isMobile }) => {
  await page.goto('/')
  if (isMobile) {
    await page.locator('details.mobile-menu summary').click()
    await page.locator('details.mobile-menu').getByRole('link', { name: 'Search documentation' }).click()
  } else await page.getByRole('link', { name: 'Search documentation', exact: true }).click()
  await expect(page.getByRole('dialog', { name: 'Search' })).toBeVisible()
  const search = page.getByRole('textbox', { name: 'Search' })
  await expect(search).toBeFocused()
  await search.fill('zzznomatchenvarium987')
  await expect(page.getByText(/No results for/)).toBeVisible()
  await search.fill('Envarium CLI contract')
  await page.getByRole('link', { name: /Envarium CLI contract/ }).first().click()
  await expect(page).toHaveURL(/\/docs\/cli\/$/)
  await page.getByRole('button', { name: 'Search', exact: true }).click()
  await page.keyboard.press('Escape')
  await expect(page.getByRole('dialog', { name: 'Search' })).not.toBeVisible()
})

test('installation remains useful without JavaScript', async ({ browser }) => {
  const context = await browser.newContext({ javaScriptEnabled: false })
  const page = await context.newPage()
  await page.goto('/download/')
  await expect(page.getByText('The package selector requires JavaScript.', { exact: false })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Update recommendation' })).not.toBeVisible()
  await page.locator('.platform-grid').getByRole('link', { name: /^Windows/ }).click()
  await expect(page.getByRole('heading', { level: 1 })).toContainText('coming soon')
  await context.close()
})


for (const blocked of [false, true]) {
  test(`docs copy link reports ${blocked ? 'blocked clipboard access' : 'success'}`, async ({ page }) => {
    await page.addInitScript((deny) => {
      Object.defineProperty(navigator, 'clipboard', { value: {
        writeText: () => deny ? Promise.reject(new Error('Permission denied')) : Promise.resolve(),
      } })
    }, blocked)
    await page.goto('/docs/cli/')
    await page.getByRole('button', { name: 'Copy canonical link' }).click()
    await expect(page.locator('[data-copy-status]')).toHaveText(blocked
      ? 'Copy was blocked by your browser. Copy the address from the address bar instead.'
      : 'Canonical link copied.')
  })
}
