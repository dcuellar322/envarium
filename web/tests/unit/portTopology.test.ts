import { expect, test } from 'vitest'
import type { PortFact } from '../../src/api/generated/types.gen'
import { portTopology } from '../../src/domains/ports/topology'
const fact = (id: string, updates: Partial<PortFact> = {}): PortFact => ({
  id,
  kind: 'binding',
  projectId: 'app',
  projectName: 'Application',
  serviceId: 'web',
  host: '127.0.0.1',
  port: 8080,
  protocol: 'tcp',
  source: 'compose',
  evidence: 'published port',
  observedAt: '2026-10-17T12:00:00Z',
  ...updates,
})
test('groups application and services, merges IPv4/IPv6 and declaration evidence', () => {
  const facts = [
    fact('v4', { target: 80 }),
    fact('v6', { host: '::', target: 80 }),
    fact('decl', { kind: 'declaration' }),
    fact('db', { serviceId: 'database', port: 5432 }),
    fact('host', { projectId: undefined, projectName: undefined, processId: 20, port: 9000 }),
    fact('inactive', {
      projectId: 'stopped',
      projectName: 'Stopped',
      kind: 'reservation',
      port: 8000,
    }),
  ]
  const active = portTopology(facts, new Set(['v6']))
  expect(active).toHaveLength(1)
  expect(active[0]?.services).toHaveLength(2)
  const endpoint = active[0]?.services.find((item) => item.id === 'web')?.endpoints[0]
  expect(endpoint).toMatchObject({
    port: 8080,
    hosts: ['127.0.0.1', '::'],
    targets: [80],
    active: true,
    conflict: true,
  })
  const all = portTopology(facts, new Set(), { configured: true, host: true, search: '' })
  expect(all).toHaveLength(3)
  expect(
    all.find((app) => app.id === 'app')?.services.find((service) => service.id === 'web')
      ?.endpoints,
  ).toHaveLength(1)
  expect(all.find((app) => app.id === 'stopped')?.services[0]?.endpoints[0]?.active).toBe(false)
  expect(
    portTopology(facts, new Set(), { configured: false, host: false, search: '5432' })[0]
      ?.services[0]?.id,
  ).toBe('database')
  expect(
    portTopology(facts, new Set(), { configured: false, host: false, search: 'missing' }),
  ).toEqual([])
})
test('conflict attribution respects fact identity and protocol', () => {
  const apps = portTopology([fact('tcp'), fact('udp', { protocol: 'udp' })], new Set(['tcp']))
  expect(apps[0]?.services[0]?.endpoints.map((item) => [item.protocol, item.conflict])).toEqual([
    ['tcp', true],
    ['udp', false],
  ])
})
