import { cleanup, render, screen, fireEvent } from '@testing-library/vue'
import { afterEach, expect, test } from 'vitest'
import GitWorktreeList from '../../src/domains/projects/components/GitWorktreeList.vue'
import type { ProjectEnvironment } from '../../src/api/generated/types.gen'

afterEach(cleanup)
const environment: ProjectEnvironment = {
  id: 'env-feature',
  projectId: 'p',
  name: 'feature',
  path: '/repo/feature',
  branch: 'feature',
  head: 'abcd',
  primary: false,
  detached: false,
  bare: false,
  locked: false,
  hostname: 'feature.localhost',
  availability: 'available',
  state: 'inactive',
  registeredAt: '2026-10-17T00:00:00Z',
  lastObservedAt: '2026-10-17T00:00:00Z',
  updatedAt: '2026-10-17T00:00:00Z',
  allocation: {
    composeProjectName: 'fixture-feature',
    portLeaseNamespace: 'fixture',
    portOffset: 100,
    portLeases: [],
  },
}
const worktree = {
  path: environment.path,
  branch: 'feature',
  head: 'abcd',
  detached: false,
  bare: false,
  locked: false,
}
test('only observed Compose checkouts offer isolated starts; native cleanup stays available', async () => {
  const view = render(GitWorktreeList, {
    props: {
      worktrees: [worktree],
      environments: [environment],
      pending: false,
      driver: 'process',
    },
  })
  expect(screen.getByRole('button', { name: 'Review start' })).toBeDisabled()
  expect(screen.getByText(/Isolated worktree starts currently support Compose/)).toBeInTheDocument()
  await fireEvent.click(screen.getByRole('button', { name: 'Review stop' }))
  expect(view.emitted().action).toEqual([['env-feature', 'stop']])
  await view.rerender({ driver: 'compose' })
  expect(screen.getByRole('button', { name: 'Review start' })).toBeEnabled()
  await view.rerender({ worktrees: [] })
  expect(screen.getByRole('button', { name: 'Review start' })).toBeDisabled()
  expect(screen.getByRole('status')).toHaveTextContent('not returned by Git')
})
