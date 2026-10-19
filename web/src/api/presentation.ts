import { client } from './generated/client.gen'

const pages = new Set([
  'projects',
  'ports',
  'resources',
  'logs',
  'discovery',
  'workspaces',
  'fleet',
  'team',
  'agents',
  'plugins',
  'settings',
  'diagnostics',
])

// Local diagnostic metadata only. Never include query strings, project IDs or
// repository paths, and never use this client-supplied label for permissions.
export function presentationLabel(path: string, native: boolean): string {
  const segment = path.split('/')[1] ?? ''
  const page = !segment ? 'dashboard' : pages.has(segment) ? segment : 'other'
  return `${native ? 'native' : 'browser'}:${page}`
}

export function observePresentationRequests() {
  const id = client.interceptors.request.use((request) => {
    const native =
      (window as Window & { __envariumNativeDesktop?: boolean }).__envariumNativeDesktop === true
    request.headers.set(
      'X-Envarium-Presentation',
      presentationLabel(window.location.pathname, native),
    )
    return request
  })
  return () => client.interceptors.request.eject(id)
}
