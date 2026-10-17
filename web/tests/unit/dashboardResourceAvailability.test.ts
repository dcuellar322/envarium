import { cleanup, render } from '@testing-library/vue'
import { afterEach, expect, test } from 'vitest'

import type { RuntimeMetricSample } from '../../src/api/generated/types.gen'
import DashboardProjectCard from '../../src/domains/dashboard/components/DashboardProjectCard.vue'
import type { ProjectSnapshot } from '../../src/domains/dashboard/api'

afterEach(cleanup)
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
