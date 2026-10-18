import { createBrowserSession } from '../../api/generated/sdk.gen'
import { client } from '../../api/generated/client.gen'

const csrfStorageKey = 'envarium.csrf-token'

export async function bootstrapBrowserSession(): Promise<void> {
  configureClient()
  const url = new URL(window.location.href)
  const bootstrapToken = url.searchParams.get('bootstrap')
  if (!bootstrapToken) return

  const result = await createBrowserSession({
    body: { bootstrapToken },
    credentials: 'omit',
  })
  url.searchParams.delete('bootstrap')
  window.history.replaceState({}, '', `${url.pathname}${url.search}${url.hash}`)
  if (result.error || !result.data) {
    throw new Error('The browser bootstrap token is invalid or expired.')
  }
  window.sessionStorage.setItem(csrfStorageKey, result.data.csrfToken)
  configureClient()
}

export function mutationHeaders(idempotencyKey: string): Record<string, string> {
  const csrfToken = window.sessionStorage.getItem(csrfStorageKey)
  return {
    'Idempotency-Key': idempotencyKey,
    ...(csrfToken ? { 'X-CSRF-Token': csrfToken } : {}),
  }
}

function configureClient() {
  client.setConfig({
    credentials: 'omit',
    headers: { 'X-CSRF-Token': window.sessionStorage.getItem(csrfStorageKey) ?? '' },
  })
}

export function browserSocketProtocols(): string[] {
  const proof = window.sessionStorage.getItem(csrfStorageKey)
  return proof ? ['envarium.v1', `envarium.session.${proof}`] : []
}
