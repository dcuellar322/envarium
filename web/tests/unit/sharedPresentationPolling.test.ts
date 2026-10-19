/* eslint-disable vue/one-component-per-file */
import { focusManager, QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { cleanup, render } from '@testing-library/vue'
import { defineComponent, h, nextTick, ref } from 'vue'
import { afterEach, expect, test, vi } from 'vitest'

const api = vi.hoisted(() => ({ host: vi.fn(), projects: vi.fn() }))
vi.mock('../../src/domains/system/api', () => ({ loadHostObservation: api.host }))
vi.mock('../../src/domains/projects/api', () => ({ loadProjects: api.projects }))
import {
  provideHostObservation,
  useHostObservation,
} from '../../src/domains/system/composables/useHostObservation'
import {
  provideProjectCatalog,
  useProjectCatalog,
} from '../../src/domains/projects/composables/useProjectCatalog'

let client: QueryClient
afterEach(() => {
  cleanup()
  client?.clear()
  focusManager.setFocused(undefined)
  vi.useRealTimers()
  vi.clearAllMocks()
})

test('shell consumers share one timer per fact, pause hidden, and retain errors and fresh data', async () => {
  vi.useFakeTimers()
  focusManager.setFocused(true)
  api.host.mockResolvedValue({ cpuPercent: 1 })
  api.projects.mockResolvedValue([{ id: 'one' }])
  const active = ref(true)
  const laterConsumer = ref(false)
  const reads: ReturnType<typeof useHostObservation>[] = []
  const Consumer = defineComponent({
    setup() {
      const host = useHostObservation()
      const projects = useProjectCatalog()
      reads.push(host)
      return () => h('p', `${host.data.value?.cpuPercent}:${projects.data.value?.length}`)
    },
  })
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    defineComponent({
      setup() {
        provideHostObservation(active)
        provideProjectCatalog(active)
        return () => h('div', [h(Consumer), laterConsumer.value ? h(Consumer) : null])
      },
    }),
    { global: { plugins: [[VueQueryPlugin, { queryClient: client }]] } },
  )
  await vi.waitFor(() => expect(api.host).toHaveBeenCalledTimes(1))
  await vi.advanceTimersByTimeAsync(1_000)
  laterConsumer.value = true
  await nextTick()
  expect(reads[0]).toBe(reads[1])
  await vi.advanceTimersByTimeAsync(14_000)
  expect(api.host).toHaveBeenCalledTimes(2)
  expect(api.projects).toHaveBeenCalledTimes(2)
  await vi.advanceTimersByTimeAsync(1_000)
  expect(api.host).toHaveBeenCalledTimes(2)
  expect(api.projects).toHaveBeenCalledTimes(2)

  active.value = false
  await nextTick()
  await vi.advanceTimersByTimeAsync(45_000)
  expect(api.host).toHaveBeenCalledTimes(2)
  expect(api.projects).toHaveBeenCalledTimes(2)
  api.host.mockRejectedValueOnce(new Error('offline'))
  active.value = true
  await nextTick()
  await vi.advanceTimersByTimeAsync(1)
  expect(reads[0]?.isError.value).toBe(true)
  expect(reads[0]?.data.value?.cpuPercent).toBe(1)
  api.host.mockResolvedValue({ cpuPercent: 9 })
  await vi.advanceTimersByTimeAsync(15_000)
  expect(reads[1]?.data.value?.cpuPercent).toBe(9)
  expect(reads[1]?.isError.value).toBe(false)
})
