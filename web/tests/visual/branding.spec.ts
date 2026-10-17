import { expect, test } from '@playwright/test'

import { installAlphaMocks } from '../helpers/alphaMocks'
import { browserBootstrapPath } from '../helpers/browserSession'

test('the shell loads the approved wordmark, favicon, and domain icons', async ({
  page,
  request,
}) => {
  await installAlphaMocks(page)
  await page.goto(browserBootstrapPath())
  const brand = page.getByRole('link', { name: 'Envarium dashboard' })
  await expect(brand.locator('.brand-wordmark')).toBeVisible()
  expect(
    await brand
      .locator('.brand-wordmark')
      .evaluate((image) => (image as HTMLImageElement).naturalWidth),
  ).toBeGreaterThan(0)
  await expect(
    page.getByRole('link', { name: 'Projects', exact: true }).locator('use'),
  ).toHaveAttribute('href', '/brand/interface.svg#projects')
  const colors = await page.locator('html').evaluate((element) => {
    const style = getComputedStyle(element)
    return ['--accent', '--accent-2', '--bg', '--panel', '--text'].map((name) =>
      style.getPropertyValue(name).trim(),
    )
  })
  expect(colors).toEqual(['#78a6ff', '#9e7bff', '#0a0d12', '#11161e', '#edf3fb'])
  expect((await request.get('/favicon.svg')).ok()).toBe(true)
  expect((await request.get('/brand/interface.svg')).ok()).toBe(true)
  await page.setViewportSize({ width: 720, height: 1050 })
  await expect(brand.locator('.brand-mark').first()).toBeVisible()
  await expect(brand.locator('.brand-wordmark')).not.toBeVisible()
})

test('high contrast preserves a legible logo and navigation at both sidebar sizes', async ({
  page,
}) => {
  await installAlphaMocks(page)
  await page.route('**/api/v1/settings', (route) =>
    route.fulfill({
      json: {
        settings: {
          appearance: { density: 'comfortable', timeDisplay: 'relative', theme: 'high-contrast' },
        },
        pendingRestart: [],
      },
    }),
  )
  await page.goto(browserBootstrapPath())
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'high-contrast')
  const brand = page.getByRole('link', { name: 'Envarium dashboard' })
  await expect(brand.locator('.brand-contrast')).toBeVisible()
  await expect(brand.locator('.brand-wordmark')).not.toBeVisible()
  await expect(page.getByRole('heading', { name: 'Alpha API' })).toBeVisible()
  await expect(page).toHaveScreenshot('branding-high-contrast.png', {
    animations: 'disabled',
    fullPage: true,
  })
  await page.setViewportSize({ width: 720, height: 1050 })
  await expect(brand.locator('.brand-mark--contrast')).toBeVisible()
  await expect(brand.locator('.brand-mark:not(.brand-mark--contrast)')).not.toBeVisible()
})
