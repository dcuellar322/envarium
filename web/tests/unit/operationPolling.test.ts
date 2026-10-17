import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { cleanup, render } from '@testing-library/vue'
import { defineComponent, h, nextTick } from 'vue'
import { afterEach, expect, test, vi } from 'vitest'

import OperationCenter from '../../src/app/components/OperationCenter.vue'
import { providePresentationActivity } from '../../src/domains/system/composables/usePresentationActivity'
const api = vi.hoisted(() => ({
  loadOperations: vi.fn(),
  requestOperationCancellation: vi.fn(),
}))
vi.mock('../../src/domains/operations/api', () => api)
let client: QueryClient
afterEach(() => {
  cleanup()
  client?.clear()
  vi.useRealTimers()
  vi.clearAllMocks()
  delete (window as Window & { __envariumNativeActive?: boolean }).__envariumNativeActive
})
test('idle connected operations avoid rapid polling and disconnection restores it', async () => {
  vi.useFakeTimers()
  api.loadOperations.mockResolvedValue([])
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const view = render(OperationCenter, {
    props: { connected: true },
    global: { plugins: [[VueQueryPlugin, { queryClient: client }]] },
  })
  await vi.advanceTimersByTimeAsync(0)
  await nextTick()
  expect(api.loadOperations).toHaveBeenCalledTimes(1)
  await vi.advanceTimersByTimeAsync(14_999)
  expect(api.loadOperations).toHaveBeenCalledTimes(1)
  await vi.advanceTimersByTimeAsync(1)
  expect(api.loadOperations).toHaveBeenCalledTimes(2)
  await view.rerender({ connected: false })
  await vi.advanceTimersByTimeAsync(1_500)
  expect(api.loadOperations).toHaveBeenCalledTimes(3)
})
test('active operations keep rapid polling while connected', async () => {
  vi.useFakeTimers()
  api.loadOperations.mockResolvedValue([
    { id: 'one', kind: 'runtime.start', state: 'running', updatedAt: '2026-10-17T00:00:00Z' },
  ])
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(OperationCenter, {
    props: { connected: true },
    global: { plugins: [[VueQueryPlugin, { queryClient: client }]] },
  })
  await vi.advanceTimersByTimeAsync(0)
  await vi.advanceTimersByTimeAsync(1_500)
  expect(api.loadOperations).toHaveBeenCalledTimes(2)
})

test('inactive native presentation pauses operation polling and resumes fresh data', async () => {
  vi.useFakeTimers()
  api.loadOperations.mockResolvedValue([])
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    defineComponent({
      setup() {
        providePresentationActivity()
        return () => h(OperationCenter, { connected: false })
      },
    }),
    { global: { plugins: [[VueQueryPlugin, { queryClient: client }]] } },
  )
  await vi.advanceTimersByTimeAsync(0)
  expect(api.loadOperations).toHaveBeenCalledTimes(1)
  const native = window as Window & { __envariumNativeActive?: boolean }
  native.__envariumNativeActive = false
  window.dispatchEvent(new Event('envarium:activity'))
  await nextTick()
  await vi.advanceTimersByTimeAsync(15_000)
  expect(api.loadOperations).toHaveBeenCalledTimes(1)
  native.__envariumNativeActive = true
  window.dispatchEvent(new Event('envarium:activity'))
  await nextTick()
  await vi.advanceTimersByTimeAsync(0)
  expect(api.loadOperations).toHaveBeenCalledTimes(2)
})
