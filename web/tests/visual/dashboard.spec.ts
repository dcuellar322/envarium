import { expect, test } from '@playwright/test'

import { installAlphaMocks } from '../helpers/alphaMocks'
import { browserBootstrapPath } from '../helpers/browserSession'

test('dashboard aligns with the alpha command-center reference', async ({ page }) => {
  await installAlphaMocks(page)
  await page.goto(browserBootstrapPath())
  await expect(page.getByRole('heading', { name: 'Your development yard' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Alpha API' })).toBeVisible()
  await expect(page.getByText('Daemon connected')).toBeVisible()
  await expect(page).toHaveScreenshot('dashboard-alpha.png', {
    animations: 'disabled',
    fullPage: true,
  })
})

test('dashboard remains coherent at a narrow viewport', async ({ page }) => {
  await page.setViewportSize({ width: 720, height: 1050 })
  await installAlphaMocks(page)
  await page.goto(browserBootstrapPath())
  await expect(page.getByRole('heading', { name: 'Your development yard' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Alpha API' })).toBeVisible()
  await expect(page).toHaveScreenshot('dashboard-narrow.png', {
    animations: 'disabled',
    fullPage: true,
  })
})

test('empty dashboard has an actionable first-run state', async ({ page }) => {
  await installAlphaMocks(page, true)
  await page.goto(browserBootstrapPath())
  await expect(page.getByText('No projects registered')).toBeVisible()
  await expect(page.getByText('Daemon connected')).toBeVisible()
  await expect(page).toHaveScreenshot('dashboard-empty.png', {
    animations: 'disabled',
    fullPage: true,
  })
})

test('phone-width project cards wrap long paths and names without horizontal scrolling', async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 1050 })
  await installAlphaMocks(page)
  await page.route('**/api/v1/projects', (route) =>
    route.fulfill({
      json: [
        {
          id: 'alpha',
          slug: 'alpha-api',
          displayName: 'Customer application with a long repository name',
          primaryLocation:
            '/Users/developer/projects/organization/customer-application-with-a-long-repository-path',
          trustState: 'trusted',
          tags: [],
          manifestRevision: 1,
          createdAt: '2026-07-16T12:00:00Z',
          updatedAt: '2026-07-16T12:00:00Z',
        },
      ],
    }),
  )
  await page.goto(browserBootstrapPath('/projects'))
  await expect(
    page.getByRole('heading', { name: 'Customer application with a long repository name' }),
  ).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await expect(page).toHaveScreenshot('project-cards-phone.png', {
    animations: 'disabled',
    fullPage: true,
  })
})
