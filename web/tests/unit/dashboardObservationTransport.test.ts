import { afterEach, expect, test, vi } from 'vitest'
const sdk = vi.hoisted(() => ({ getProjectObservations: vi.fn() }))
vi.mock('../../src/api/generated/sdk.gen', () => sdk)
import { loadProjectObservations } from '../../src/domains/dashboard/observations'
afterEach(() => vi.clearAllMocks())
test('large dashboards split bounded requests and preserve the cancellation signal', async () => {
  sdk.getProjectObservations.mockImplementation(async ({ query }) => ({
    data: { projects: query.projectId.map((projectId: string) => ({ projectId })) },
  }))
  const controller = new AbortController()
  const ids = Array.from({ length: 130 }, (_, i) => String(i))
  const result = await loadProjectObservations('runtime', ids, controller.signal)
  expect(result.map((item) => item.projectId)).toEqual(ids)
  expect(
    sdk.getProjectObservations.mock.calls.map(([options]) => options.query.projectId.length),
  ).toEqual([64, 64, 2])
  expect(
    sdk.getProjectObservations.mock.calls.every(
      ([options]) => options.signal === controller.signal,
    ),
  ).toBe(true)
  controller.abort()
  await expect(loadProjectObservations('runtime', ids, controller.signal)).rejects.toThrow()
  expect(sdk.getProjectObservations).toHaveBeenCalledTimes(3)
})
