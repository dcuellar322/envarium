import { cleanup, render } from '@testing-library/vue'
import { afterEach, expect, test, vi } from 'vitest'
import * as format from '../../src/lib/format'

import type { RuntimeMetricSample } from '../../src/api/generated/types.gen'
import DashboardProjectCard from '../../src/domains/dashboard/components/DashboardProjectCard.vue'
import type { ProjectSnapshot } from '../../src/domains/dashboard/api'

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})
const project: ProjectSnapshot['project'] = {
  id: 'one',
  slug: 'one',
  displayName: 'One',
  primaryLocation: '/one',
  trustState: 'trusted',
  tags: [],
  manifestRevision: 1,
  createdAt: '2026-10-17T00:00:00Z',
  updatedAt: '2026-10-17T00:00:00Z',
}
test('unavailable samples cannot appear as measured zero CPU or memory', () => {
  const metric = {
    cpuPercent: 0,
    cpuAvailable: false,
    memoryBytes: 0,
    memoryAvailable: false,
  } as RuntimeMetricSample
  const view = render(DashboardProjectCard, {
    props: { snapshot: { project, metrics: [metric], warnings: ['Resource samples are partial'] } },
    global: { stubs: { RouterLink: true } },
  })
  expect(view.getByText('— · —')).toBeVisible()
  expect(view.queryByText('0 B · 0.0%')).toBeNull()
  expect(view.getByText('Partial data')).toBeVisible()
})
test('known stopped projects can show zero without inventing unavailable samples', () => {
  const view = render(DashboardProjectCard, {
    props: { snapshot: { project, metrics: [], warnings: [] } },
    global: { stubs: { RouterLink: true } },
  })
  expect(view.getByText('0 B · 0.0%')).toBeVisible()
})

test('timestamp-only observations skip card rendering while actions use the latest snapshot', async () => {
  const initials = vi.spyOn(format, 'projectInitials')
  const metric = {
    cpuPercent: 2,
    cpuAvailable: true,
    memoryBytes: 1024,
    memoryAvailable: true,
  } as RuntimeMetricSample
  const first: ProjectSnapshot = { project, metrics: [metric], warnings: [] }
  const view = render(DashboardProjectCard, {
    props: { snapshot: first },
    global: { stubs: { RouterLink: true } },
  })
  expect(initials).toHaveBeenCalledTimes(1)
  const current = { ...first, metrics: [{ ...metric, timestamp: '2026-10-18T01:00:00Z' }] }
  await view.rerender({ snapshot: current })
  expect(initials).toHaveBeenCalledTimes(1)
  view.getByRole('button', { name: 'Start' }).click()
  expect(view.emitted()).toMatchObject({ runtime: [[current, 'start']] })
  await view.rerender({ snapshot: { ...current, metrics: [{ ...metric, cpuPercent: 5 }] } })
  expect(view.getByText('1.0 KB · 5.0%')).toBeVisible()
  expect(initials).toHaveBeenCalledTimes(2)
  await view.rerender({ pending: true })
  expect(view.getByRole('button', { name: 'Start' })).toBeDisabled()
  await view.rerender({ snapshot: { ...current, warnings: ['Metrics unavailable'] } })
  expect(view.getByText('Partial data')).toHaveAttribute('title', 'Metrics unavailable')
})
