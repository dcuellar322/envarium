import type { PortFact } from '../../api/generated/types.gen'

export type PortEndpoint = {
  key: string
  port: number
  protocol: string
  hosts: string[]
  targets: number[]
  active: boolean
  conflict: boolean
}
export type PortApplication = {
  id: string
  name: string
  services: Array<{ id: string; name: string; endpoints: PortEndpoint[] }>
}

/** Presentation grouping only. Safety decisions retain the complete registry. */
export function portTopology(
  facts: PortFact[],
  conflictIds: Set<string>,
  options = { configured: false, host: false, search: '' },
): PortApplication[] {
  const apps = new Map<string, PortApplication>()
  const search = options.search.trim().toLocaleLowerCase()
  for (const fact of facts) {
    if (!fact.projectId && !options.host) continue
    if (fact.kind !== 'binding' && !options.configured) continue
    const name = fact.projectName || fact.projectId || 'Other host listeners'
    const serviceName =
      fact.serviceId || (fact.processId ? `Process ${fact.processId}` : 'Unattributed service')
    if (
      search &&
      !`${name} ${serviceName} ${fact.port} ${fact.protocol}`.toLocaleLowerCase().includes(search)
    )
      continue
    const appId = fact.projectId || 'host-listeners'
    let app = apps.get(appId)
    if (!app) {
      app = { id: appId, name, services: [] }
      apps.set(appId, app)
    }
    const serviceId = fact.serviceId || `${fact.source}:${fact.processId || 'unknown'}`
    let service = app.services.find((item) => item.id === serviceId)
    if (!service) {
      service = { id: serviceId, name: serviceName, endpoints: [] }
      app.services.push(service)
    }
    const key = `${fact.port}/${fact.protocol}`
    let endpoint = service.endpoints.find((item) => item.key === key)
    if (!endpoint) {
      endpoint = {
        key,
        port: fact.port,
        protocol: fact.protocol,
        hosts: [],
        targets: [],
        active: false,
        conflict: false,
      }
      service.endpoints.push(endpoint)
    }
    endpoint.active ||= fact.kind === 'binding'
    endpoint.conflict ||= conflictIds.has(fact.id)
    if (!endpoint.hosts.includes(fact.host)) endpoint.hosts.push(fact.host)
    if (fact.target && !endpoint.targets.includes(fact.target)) endpoint.targets.push(fact.target)
  }
  const result = [...apps.values()].sort((a, b) => a.name.localeCompare(b.name))
  for (const app of result) {
    app.services.sort((a, b) => a.name.localeCompare(b.name))
    for (const service of app.services)
      service.endpoints.sort((a, b) => a.port - b.port || a.protocol.localeCompare(b.protocol))
  }
  return result
}
