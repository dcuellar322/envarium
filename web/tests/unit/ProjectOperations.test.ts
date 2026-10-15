import { cleanup, fireEvent, render } from '@testing-library/vue'
import { afterEach, expect, test } from 'vitest'

import type { RuntimeObservation } from '../../src/api/generated/types.gen'
import ProjectLogsTab from '../../src/domains/projects/components/ProjectLogsTab.vue'
import ProjectServicesPanel from '../../src/domains/projects/components/ProjectServicesPanel.vue'

afterEach(cleanup)

const runtime: RuntimeObservation = {
  projectId: 'alpha',
  driver: 'compose',
  projectIdentity: 'alpha',
  state: 'partially_running',
  origin: 'envarium',
  services: [
    {
      id: 'api',
      runtimeName: 'alpha-api',
      state: 'running',
      health: 'healthy',
      ports: [],
      observedAt: '2026-09-22T12:00:00Z',
    },
    {
      id: 'worker',
      runtimeName: 'alpha-worker',
      state: 'paused',
      health: 'unknown',
      ports: [],
      observedAt: '2026-09-22T12:00:00Z',
    },
  ],
  observedAt: '2026-09-22T12:00:00Z',
}

test('targets lifecycle controls to one declared service', async () => {
  const view = render(ProjectServicesPanel, {
    props: { runtime, pending: false, actionPending: false, metrics: [] },
  })

  await fireEvent.click(view.getByRole('button', { name: 'Deploy api' }))
  await fireEvent.click(view.getByRole('button', { name: 'Resume worker' }))
  await fireEvent.click(view.getByRole('button', { name: 'Force stop api' }))

  expect(view.emitted('lifecycle')).toEqual([
    ['deploy', { services: ['api'] }],
    ['unpause', { services: ['worker'] }],
    ['kill', { services: ['api'] }],
  ])
})

test('filters live logs without changing the bounded source', async () => {
  const view = render(ProjectLogsTab, {
    props: {
      projectId: 'alpha',
      connection: 'connected',
      services: ['api', 'worker'],
      metrics: [],
      entries: [
        {
          sequence: 1,
          timestamp: '2026-09-22T12:00:00Z',
          projectId: 'alpha',
          serviceId: 'api',
          runId: 'run-1',
          source: 'docker',
          stream: 'stdout',
          level: 'info',
          message: '{"event":"ready","port":8080}',
          redacted: false,
          attributes: {},
        },
        {
          sequence: 2,
          timestamp: '2026-09-22T12:00:01Z',
          projectId: 'alpha',
          serviceId: 'worker',
          runId: 'run-1',
          source: 'docker',
          stream: 'stderr',
          level: 'error',
          message: 'queue unavailable',
          redacted: false,
          attributes: {},
        },
      ],
    },
  })

  expect(view.getByText('2 of 2 bounded entries')).toBeInTheDocument()
  expect(view.getByText('2 fields')).toBeInTheDocument()
  await fireEvent.update(view.getByPlaceholderText('Search logs…'), 'queue')
  expect(view.queryByText('{"event":"ready","port":8080}')).not.toBeInTheDocument()
  expect(view.getByText('queue unavailable')).toBeInTheDocument()
  expect(view.getByText('1 of 2 bounded entries')).toBeInTheDocument()
})
