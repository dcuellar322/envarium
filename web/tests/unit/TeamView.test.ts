import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/vue'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

const api = vi.hoisted(() => ({
  installBundle: vi.fn(),
  loadCuratedPlugins: vi.fn(),
  loadEffectivePolicy: vi.fn(),
  loadTeamBundles: vi.fn(),
  loadTeamPublishers: vi.fn(),
  trustPublisher: vi.fn(),
}))
vi.mock('../../src/domains/team/api', () => api)
import TeamView from '../../src/domains/team/views/TeamView.vue'

const publisher = {
  id: `publisher-${'a'.repeat(32)}`,
  name: 'Envarium maintainers',
  publicKey: 'cHVibGlj',
  trustedAt: '2026-07-16T12:00:00Z',
}

afterEach(cleanup)
beforeEach(() => {
  vi.clearAllMocks()
  api.loadTeamPublishers.mockResolvedValue([publisher])
  api.loadTeamBundles.mockResolvedValue([])
  api.loadEffectivePolicy.mockResolvedValue({
    sourceBundleIds: ['policy.official'],
    allowedRemoteCapabilities: ['inventory.read'],
    allowedRemoteActions: ['start'],
    allowedRegistryPublishers: [publisher.id],
    telemetryAllowed: false,
  })
  api.loadCuratedPlugins.mockResolvedValue([
    {
      id: 'reviewed-plugin',
      name: 'Reviewed plugin',
      version: '1.0.0',
      summary: 'Signed metadata.',
      publisher: publisher.id,
      downloadUrl: 'https://plugins.example.test/plugin.tar.gz',
      sha256: 'b'.repeat(64),
      platforms: ['linux/amd64'],
      capabilities: ['project.inspect'],
    },
  ])
  api.trustPublisher.mockResolvedValue(publisher)
})

function renderView() {
  return render(TeamView, {
    global: {
      plugins: [
        [
          VueQueryPlugin,
          {
            queryClient: new QueryClient({
              defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
            }),
          },
        ],
      ],
    },
  })
}

test('shows restrictive signed policy and keeps publisher trust behind review', async () => {
  renderView()
  expect(await screen.findByText('Envarium maintainers')).toBeInTheDocument()
  expect(await screen.findByText('Reviewed plugin')).toBeInTheDocument()
  expect(screen.getByText('denied')).toBeInTheDocument()
  const button = screen.getByRole('button', { name: 'Trust exact key' })
  expect(button).toBeDisabled()
  await fireEvent.update(screen.getByLabelText('Name'), 'Maintainers')
  await fireEvent.update(screen.getByLabelText('Base64 Ed25519 public key'), 'cHVibGlj')
  await fireEvent.click(screen.getByRole('checkbox', { name: /verified this exact key/i }))
  await fireEvent.click(button)
  await waitFor(() =>
    expect(api.trustPublisher).toHaveBeenCalledWith('Maintainers', 'cHVibGlj', true),
  )
})

test('explains that encrypted sync excludes sensitive local state', async () => {
  renderView()
  expect(await screen.findByRole('heading', { name: 'Encrypted sync' })).toBeInTheDocument()
  expect(
    screen.getByText(
      /Fleet credentials, projects, paths, logs, operations, and secrets are never included/,
    ),
  ).toBeInTheDocument()
})

test('explains the local configuration scope when no team data is installed', async () => {
  api.loadTeamPublishers.mockResolvedValue([])
  api.loadTeamBundles.mockResolvedValue([])
  api.loadEffectivePolicy.mockResolvedValue({
    sourceBundleIds: [],
    allowedRemoteCapabilities: [],
    allowedRemoteActions: [],
    allowedRegistryPublishers: [],
    telemetryAllowed: false,
  })
  api.loadCuratedPlugins.mockResolvedValue([])

  renderView()

  expect(
    await screen.findByRole('heading', { name: 'No shared team configuration is installed' }),
  ).toBeInTheDocument()
  expect(screen.getByText(/not a live member directory/)).toBeInTheDocument()
})

const signedBundle = {
  schemaVersion: 'envarium.bundle/v1',
  kind: 'policy-pack',
  metadata: {
    id: 'policy.reviewed',
    name: 'Local development policy',
    version: '1.0.0',
    publisherId: publisher.id,
    createdAt: '2026-07-16T12:00:00Z',
  },
  payload: { allowedRemoteActions: ['start'], telemetryAllowed: false },
  signature: { keyId: publisher.id, algorithm: 'Ed25519', value: 'fixture-signature' },
}

async function selectFile(text: () => Promise<string>, size = 1000) {
  await fireEvent.change(screen.getByLabelText('Signed JSON bundle'), {
    target: { files: [{ size, text }] },
  })
}

test('changing the publisher key revokes its previous confirmation', async () => {
  renderView()
  await screen.findByRole('heading', { name: 'Publisher identities' })
  await fireEvent.update(screen.getByLabelText('Name'), 'Owner')
  const input = screen.getByLabelText('Base64 Ed25519 public key')
  await fireEvent.update(input, 'first-key')
  const confirmation = screen.getByRole('checkbox', { name: /verified this exact key/i })
  await fireEvent.click(confirmation)
  expect(screen.getByRole('button', { name: 'Trust exact key' })).toBeEnabled()
  await fireEvent.update(input, 'replacement-key')
  expect(confirmation).not.toBeChecked()
  expect(screen.getByRole('button', { name: 'Trust exact key' })).toBeDisabled()
  expect(api.trustPublisher).not.toHaveBeenCalled()
})

test('shows the exact payload and requires fresh review when the file changes', async () => {
  renderView()
  await screen.findByRole('heading', { name: 'Signed bundles' })
  await selectFile(async () => JSON.stringify(signedBundle))
  expect(
    await screen.findByRole('heading', { name: signedBundle.metadata.name }),
  ).toBeInTheDocument()
  expect(
    JSON.parse((screen.getByLabelText('Signed payload JSON') as HTMLTextAreaElement).value),
  ).toEqual(signedBundle.payload)
  const confirmation = screen.getByRole('checkbox', { name: /reviewed the kind/i })
  await fireEvent.click(confirmation)
  const install = screen.getByRole('button', { name: 'Verify signature and install' })
  expect(install).toBeEnabled()
  const changed = { ...signedBundle, metadata: { ...signedBundle.metadata, version: '2.0.0' } }
  await selectFile(async () => JSON.stringify(changed))
  await waitFor(() => expect(screen.getByText('2.0.0')).toBeInTheDocument())
  expect(confirmation).not.toBeChecked()
  expect(install).toBeDisabled()
  await fireEvent.click(confirmation)
  api.installBundle.mockResolvedValue(changed)
  await fireEvent.click(install)
  await waitFor(() => expect(api.installBundle).toHaveBeenCalledWith(changed, true))
})

test.each(['null', '{}', '{"metadata":null}', 'invalid-json'])(
  'rejects malformed bundle files without crashing or retaining consent: %s',
  async (text) => {
    renderView()
    await screen.findByRole('heading', { name: 'Signed bundles' })
    await selectFile(async () => text)
    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Verify signature and install' })).toBeDisabled()
    expect(api.installBundle).not.toHaveBeenCalled()
  },
)

test('late reads cannot replace a more recently selected file', async () => {
  renderView()
  await screen.findByRole('heading', { name: 'Signed bundles' })
  let finish!: (text: string) => void
  const pending = new Promise<string>((resolve) => {
    finish = resolve
  })
  await selectFile(() => pending)
  const latest = {
    ...signedBundle,
    metadata: { ...signedBundle.metadata, name: 'Latest selection' },
  }
  await selectFile(async () => JSON.stringify(latest))
  expect(await screen.findByRole('heading', { name: 'Latest selection' })).toBeInTheDocument()
  finish(JSON.stringify(signedBundle))
  await new Promise((resolve) => setTimeout(resolve, 0))
  expect(screen.getByRole('heading', { name: 'Latest selection' })).toBeInTheDocument()
  expect(
    screen.queryByRole('heading', { name: signedBundle.metadata.name }),
  ).not.toBeInTheDocument()
})

test('oversized and deeply nested bundle files fail before any installation', async () => {
  renderView()
  await screen.findByRole('heading', { name: 'Signed bundles' })
  const text = vi.fn(async () => JSON.stringify(signedBundle))
  await selectFile(text, 2 * 1024 * 1024 + 1)
  expect(await screen.findByRole('alert')).toHaveTextContent('2 MiB')
  expect(text).not.toHaveBeenCalled()
  let payload: Record<string, unknown> = {}
  for (let depth = 0; depth < 40; depth++) payload = { child: payload }
  await selectFile(async () => JSON.stringify({ ...signedBundle, payload }))
  expect(await screen.findByRole('alert')).toHaveTextContent('too deeply nested')
  expect(api.installBundle).not.toHaveBeenCalled()
})

test('configuration errors have a retry instead of claiming that no data is installed', async () => {
  api.loadEffectivePolicy.mockRejectedValueOnce(new Error('Policy unavailable'))
  renderView()
  expect(await screen.findByRole('alert')).toHaveTextContent('Policy unavailable')
  expect(
    screen.queryByRole('heading', { name: 'No shared team configuration is installed' }),
  ).not.toBeInTheDocument()
  await fireEvent.click(screen.getByRole('button', { name: 'Retry configuration' }))
  await waitFor(() => expect(screen.queryByRole('alert')).not.toBeInTheDocument())
})
