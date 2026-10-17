/* eslint-disable vue/one-component-per-file */
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { cleanup, render, waitFor } from '@testing-library/vue'
import { defineComponent, h, nextTick, ref, type Ref } from 'vue'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import type { Project } from '../../src/api/generated/types.gen'

const api = vi.hoisted(() => ({
  loadProjectRuntime: vi.fn(),
  loadProjectGitStatus: vi.fn(),
  loadProjectActions: vi.fn(),
  loadProjectMetrics: vi.fn(),
}))
const batch = vi.hoisted(() => ({ loadProjectObservations: vi.fn() }))
const resources = vi.hoisted(() => ({ loadResourceOverview: vi.fn() }))
vi.mock('../../src/domains/projects/api', () => api)
vi.mock('../../src/domains/dashboard/observations', () => batch)
vi.mock('../../src/domains/resources/api', () => resources)
import { useDashboardObservations } from '../../src/domains/dashboard/composables/useDashboardObservations'
import { providePresentationActivity } from '../../src/domains/system/composables/usePresentationActivity'
import type { ProjectSnapshot } from '../../src/domains/dashboard/api'

const project: Project = {
  id: 'one',
  slug: 'one',
  displayName: 'One',
  primaryLocation: '/one',
  trustState: 'trusted',
  manifestRevision: 1,
  tags: [],
  createdAt: '2026-10-17T00:00:00Z',
  updatedAt: '2026-10-17T00:00:00Z',
}
type NativeWindow = Window & { __envariumNativeActive?: boolean }
let client: QueryClient
let observed: Ref<ProjectSnapshot[]>
function nativeActivity(active: boolean) {
  ;(window as NativeWindow).__envariumNativeActive = active
  window.dispatchEvent(new Event('envarium:activity'))
}
function mountDashboard(items = [project]) {
  client = new QueryClient({ defaultOptions: { queries: { retry: false, staleTime: 10_000 } } })
  const Content = defineComponent({
    setup() {
      const snapshots = useDashboardObservations(ref(items))
      observed = snapshots.data
      return () => h('div', snapshots.data.value[0]?.runtime?.state ?? 'pending')
    },
  })
  return render(
    defineComponent({
      setup() {
        providePresentationActivity()
        return () => h(Content)
      },
    }),
    { global: { plugins: [[VueQueryPlugin, { queryClient: client }]] } },
  )
}
beforeEach(() => {
  vi.clearAllMocks()
  batch.loadProjectObservations.mockImplementation(
    (kind: string, ids: readonly string[], signal?: AbortSignal) =>
      Promise.all(
        ids.map(async (id) => {
          if (kind === 'runtime')
            return { projectId: id, runtime: await api.loadProjectRuntime(id, signal) }
          if (kind === 'git-status')
            return { projectId: id, git: await api.loadProjectGitStatus(id, signal) }
          return { projectId: id, actions: await api.loadProjectActions(id, signal) }
        }),
      ),
  )

  api.loadProjectRuntime.mockResolvedValue({ state: 'running' })
  api.loadProjectGitStatus.mockResolvedValue({ branch: 'main' })
  api.loadProjectActions.mockResolvedValue({ actions: [] })
  resources.loadResourceOverview.mockResolvedValue({
    projects: [
      {
        projectId: 'one',
        active: true,
        metric: { cpuAvailable: true, memoryAvailable: true },
        services: [],
      },
    ],
  })
})
afterEach(() => {
  cleanup()
  client?.clear()
  delete (window as NativeWindow).__envariumNativeActive
})
test('runtime invalidations never re-read Git, actions or direct per-project metrics', async () => {
  mountDashboard()
  await waitFor(() => expect(api.loadProjectGitStatus).toHaveBeenCalledTimes(1))
  await waitFor(() => expect(client.isFetching()).toBe(0))
  for (let i = 0; i < 4; i++)
    await client.invalidateQueries({ queryKey: ['project-observations', 'runtime'] })
  expect(api.loadProjectRuntime).toHaveBeenCalledTimes(5)
  expect(api.loadProjectGitStatus).toHaveBeenCalledTimes(1)
  expect(api.loadProjectActions).toHaveBeenCalledTimes(1)
  expect(resources.loadResourceOverview).toHaveBeenCalledTimes(1)
  expect(api.loadProjectMetrics).not.toHaveBeenCalled()
})
test('hidden native presentation suppresses invalidation fetches and refreshes on return', async () => {
  mountDashboard()
  await waitFor(() => expect(api.loadProjectGitStatus).toHaveBeenCalledTimes(1))
  nativeActivity(false)
  await nextTick()
  await waitFor(() => expect(client.isFetching()).toBe(0))
  await client.invalidateQueries()
  expect(api.loadProjectGitStatus).toHaveBeenCalledTimes(1)
  expect(api.loadProjectRuntime).toHaveBeenCalledTimes(1)
  nativeActivity(true)
  await waitFor(() => expect(api.loadProjectGitStatus).toHaveBeenCalledTimes(2))
  expect(api.loadProjectRuntime).toHaveBeenCalledTimes(2)
})
test('hiding aborts an in-flight Git request', async () => {
  let received: AbortSignal | undefined
  api.loadProjectGitStatus.mockImplementationOnce((_id: string, signal: AbortSignal) => {
    received = signal
    return new Promise((_resolve, reject) =>
      signal.addEventListener('abort', () => reject(new DOMException('aborted', 'AbortError'))),
    )
  })
  mountDashboard()
  await waitFor(() => expect(received).toBeDefined())
  nativeActivity(false)
  await nextTick()
  await waitFor(() => expect(received?.aborted).toBe(true))
})
test('pending repositories never trigger trusted observations', async () => {
  mountDashboard([{ ...project, trustState: 'pending' }])
  await waitFor(() => expect(resources.loadResourceOverview).toHaveBeenCalledTimes(1))
  expect(api.loadProjectRuntime).not.toHaveBeenCalled()
  expect(api.loadProjectGitStatus).not.toHaveBeenCalled()
  expect(api.loadProjectActions).not.toHaveBeenCalled()
})

test('fetch progress and unrelated project changes preserve card props', async () => {
  mountDashboard([project, { ...project, id: 'two' }])
  await waitFor(() => expect(observed.value.every((item) => item.git)).toBe(true))
  const before = observed.value
  let finish: (value: unknown) => void = () => undefined
  api.loadProjectRuntime.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        finish = resolve
      }),
  )
  const fetching = client.invalidateQueries({ queryKey: ['project-observations', 'runtime'] })
  await waitFor(() => expect(client.isFetching()).toBe(1))
  expect(observed.value[0]).toBe(before[0])
  expect(observed.value[1]).toBe(before[1])
  finish({ state: 'stopped' })
  await fetching
  await waitFor(() => expect(observed.value[0]?.runtime?.state).toBe('stopped'))
  expect(observed.value[1]).toBe(before[1])
})

test('many project cards share three kind-specific transport requests', async () => {
  mountDashboard(Array.from({ length: 11 }, (_, i) => ({ ...project, id: String(i) })))
  await waitFor(() => expect(observed.value.every((item) => item.git)).toBe(true))
  expect(batch.loadProjectObservations).toHaveBeenCalledTimes(3)
  for (const [kind, ids, signal] of batch.loadProjectObservations.mock.calls) {
    expect(['runtime', 'git-status', 'actions']).toContain(kind)
    expect(ids).toHaveLength(11)
    expect(signal).toBeInstanceOf(AbortSignal)
  }
})
