import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { cleanup, render, waitFor } from '@testing-library/vue'
import { defineComponent, computed } from 'vue'
import { afterEach, expect, test, vi } from 'vitest'
import { useCheckoutTransition } from '../../src/domains/projects/composables/useCheckoutTransition'
import { trackOperation } from '../../src/domains/operations/store'
import type { Operation } from '../../src/api/generated/types.gen'
const api = vi.hoisted(() => ({
  loadProjectRuntime: vi.fn(),
  previewRuntimeAction: vi.fn(),
  runRuntimeAction: vi.fn(),
  previewBranchSwitch: vi.fn(),
  switchBranch: vi.fn(),
  registerEnvironments: vi.fn(),
}))
vi.mock('../../src/domains/projects/api', () => api)
vi.mock('../../src/domains/projects/gitApi', () => api)
vi.mock('../../src/domains/environments/api', () => api)
afterEach(() => {
  cleanup()
  vi.resetAllMocks()
})
let transition: ReturnType<typeof useCheckoutTransition>
function setup() {
  render(
    defineComponent({
      setup() {
        transition = useCheckoutTransition(
          computed(() => 'p'),
          vi.fn(),
        )
        return () => null
      },
    }),
    {
      global: {
        plugins: [
          [
            VueQueryPlugin,
            { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) },
          ],
        ],
      },
    },
  )
  api.loadProjectRuntime.mockResolvedValue({ state: 'running' })
  api.previewRuntimeAction.mockImplementation(async (projectId, action) => ({
    projectId,
    action,
    checkoutRevision: 'reviewed',
    manifestHash: 'hash',
  }))
  api.previewBranchSwitch.mockResolvedValue({ projectId: 'env-a', toBranch: 'feature' })
}
function operation(id: string, state = 'queued') {
  return {
    id,
    projectId: 'env-a',
    kind: 'runtime.stop',
    state,
    cancellationRequested: false,
    requestedAt: '2026-10-18T00:00:00Z',
    updatedAt: '2026-10-18T00:00:00Z',
  } as Operation
}
test('reviews stop, branch, then fresh start and waits for each successful operation', async () => {
  setup()
  api.runRuntimeAction.mockResolvedValueOnce(operation('transition-stop'))
  api.switchBranch.mockResolvedValueOnce(operation('transition-branch'))
  await transition.prepare(['env-a'], 'env-a', 'feature')
  expect(transition.runtimePlan.value?.action).toBe('stop')
  expect(api.previewBranchSwitch).not.toHaveBeenCalled()
  await transition.confirm()
  expect(api.runRuntimeAction).toHaveBeenCalledWith('env-a', 'stop', {
    confirmRisk: true,
    expectedCheckoutRevision: 'reviewed',
    expectedManifestHash: 'hash',
  })
  expect(api.previewBranchSwitch).not.toHaveBeenCalled()
  trackOperation(operation('transition-stop', 'succeeded'))
  await waitFor(() => expect(transition.branchPlan.value?.toBranch).toBe('feature'))
  expect(api.previewRuntimeAction).toHaveBeenCalledTimes(1)
  await transition.confirm()
  trackOperation(operation('transition-branch', 'succeeded'))
  await waitFor(() => expect(transition.runtimePlan.value?.action).toBe('start'))
  expect(api.registerEnvironments).toHaveBeenCalledWith('p')
  expect(api.previewRuntimeAction).toHaveBeenLastCalledWith('env-a', 'start')
})
test('failure stops the sequence; cancellation of review performs no mutation', async () => {
  setup()
  api.runRuntimeAction.mockResolvedValueOnce(operation('transition-fail'))
  await transition.prepare(['env-a'], 'env-b')
  transition.cancel()
  expect(api.runRuntimeAction).not.toHaveBeenCalled()
  await transition.prepare(['env-a'], 'env-b')
  await transition.confirm()
  trackOperation(operation('transition-fail', 'failed'))
  await waitFor(() => expect(transition.error.value).toContain('failed'))
  expect(api.previewRuntimeAction).toHaveBeenCalledTimes(2)
  expect(transition.runtimePlan.value).toBeUndefined()
})
test('leaving checkout context cancels future stages while queued work remains tracked', async () => {
  setup()
  api.runRuntimeAction.mockResolvedValueOnce(operation('transition-unmount'))
  await transition.prepare(['env-a'], 'env-b')
  await transition.confirm()
  cleanup()
  trackOperation(operation('transition-unmount', 'succeeded'))
  await new Promise((resolve) => setTimeout(resolve, 0))
  expect(api.previewRuntimeAction).toHaveBeenCalledTimes(1)
})
test('failed port reconciliation leaves the new branch stopped without preparing a start', async () => {
  setup()
  api.loadProjectRuntime.mockResolvedValue({ state: 'stopped' })
  api.switchBranch.mockResolvedValueOnce(operation('transition-lease-fail'))
  api.registerEnvironments.mockRejectedValueOnce(new Error('Port allocation unavailable'))
  await transition.prepare(['env-a'], 'env-a', 'feature')
  await transition.confirm()
  trackOperation(operation('transition-lease-fail', 'succeeded'))
  await waitFor(() => expect(transition.error.value).toContain('Port allocation unavailable'))
  expect(api.previewRuntimeAction).not.toHaveBeenCalled()
  expect(transition.runtimePlan.value).toBeUndefined()
})
