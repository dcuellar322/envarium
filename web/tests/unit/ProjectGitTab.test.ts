import { cleanup, render, screen } from '@testing-library/vue'
import { afterEach, expect, test } from 'vitest'
import ProjectGitTab from '../../src/domains/projects/components/ProjectGitTab.vue'
import type { ProjectEnvironment } from '../../src/api/generated/types.gen'
afterEach(cleanup)
test('native port guidance preserves the registered worktree actions', () => {
  const environment = {
    id: 'env-native',
    projectId: 'p',
    name: 'feature',
    path: '/repo/feature',
    branch: 'feature',
    head: 'abcd',
    primary: false,
    detached: false,
    bare: false,
    locked: false,
    state: 'inactive',
    registeredAt: '2026-10-18T00:00:00Z',
    lastObservedAt: '2026-10-18T00:00:00Z',
    updatedAt: '2026-10-18T00:00:00Z',
    availability: 'available',
    allocation: {
      portLeases: [],
      composeProjectName: 'fixture',
      portLeaseNamespace: 'fixture',
      portOffset: 1,
    },
    hostname: 'feature.localhost',
  } satisfies ProjectEnvironment
  render(ProjectGitTab, {
    props: {
      driver: 'process',
      environments: [environment],
      environmentsPending: false,
      environmentsError: false,
      registrationPending: false,
    },
    global: { stubs: { CheckoutControls: true } },
  })
  expect(screen.getByText(/Native worktrees use independent ports/)).toBeInTheDocument()
  // Persisted checkouts stay visible even before a new Git inventory arrives.
  expect(screen.getByRole('button', { name: 'Review start' })).toBeInTheDocument()
  expect(screen.getByRole('button', { name: 'Review stop' })).toBeEnabled()
})
