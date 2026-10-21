import { createHash, generateKeyPairSync, randomBytes, sign } from 'node:crypto'
import { describe, expect, it } from 'vitest'
import type { GitHubRelease } from '../../src/lib/releases/types'
import { buildUpdaterFeed, publishedUpdaterRelease, updaterAssetTarget } from '../../src/lib/releases/updater'
import { authenticateUpdaterSignature, verifyUpdaterPayload } from '../../src/lib/releases/updaterSignature'
import { pagesWorkerSource } from '../../src/lib/hosting/pagesWorker'
import { desktopArtifactName } from '../../src/lib/releases/desktopArtifacts'

// Ephemeral keys exercise the actual Minisign wire format without a private fixture.
function signer() {
  const { privateKey, publicKey } = generateKeyPairSync('ed25519')
  const keyID = randomBytes(8)
  const key = Buffer.concat([Buffer.from('Ed'), keyID, publicKey.export({ type: 'spki', format: 'der' }).subarray(-32)])
  const encodedKey = Buffer.from(`untrusted comment: test public key\n${key.toString('base64')}\n`).toString('base64')
  return {
    publicKey: encodedKey,
    sign(payload: Uint8Array, comment = 'timestamp:123\tfile:fixture\tversion:1.0.0') {
      const signature = sign(null, createHash('blake2b512').update(payload).digest(), privateKey)
      const globalSignature = sign(null, Buffer.concat([signature, Buffer.from(comment)]), privateKey)
      return Buffer.from(`untrusted comment: test signature\n${Buffer.concat([Buffer.from('ED'), keyID, signature]).toString('base64')}\ntrusted comment: ${comment}\n${globalSignature.toString('base64')}\n`).toString('base64')
    },
  }
}

const names = [
  'Envarium_1.0.0_aarch64.app.tar.gz', 'Envarium_1.0.0_x64.app.tar.gz',
  'Envarium_1.0.0_amd64.AppImage', 'Envarium_1.0.0_amd64.deb',
  'Envarium-1.0.0-1.x86_64.rpm',
]
function release(): GitHubRelease {
  return {
    tag_name: 'v1.0.0', body: 'Reviewed release', draft: false, prerelease: false,
    published_at: '2026-10-19T00:00:00Z', created_at: '2026-10-19T00:00:00Z',
    html_url: 'https://github.com/dcuellar322/envarium/releases/tag/v1.0.0',
    assets: names.flatMap((name) => [name, `${name}.sig`]).map((name) => ({
      name, size: 100, browser_download_url: `https://github.com/dcuellar322/envarium/releases/download/v1.0.0/${name}`,
    })),
  }
}

describe('updater signature trust', () => {
  it('verifies payload and authenticated metadata, rejecting changes and foreign keys', () => {
    const key = signer()
    const payload = Buffer.from('artifact')
    const signature = key.sign(payload)
    expect(() => verifyUpdaterPayload(payload, signature, key.publicKey, '1.0.0')).not.toThrow()
    expect(() => verifyUpdaterPayload(Buffer.from('tampered'), signature, key.publicKey, '1.0.0')).toThrow(/payload signature/)
    expect(() => authenticateUpdaterSignature(signature, signer().publicKey, '1.0.0')).toThrow()
    const changed = Buffer.from(Buffer.from(signature, 'base64').toString().replace('timestamp:123', 'timestamp:456')).toString('base64')
    expect(() => authenticateUpdaterSignature(changed, key.publicKey, '1.0.0')).toThrow(/metadata/)
    for (const malformed of ['', '!', 'a'.repeat(8193), Buffer.from('invalid').toString('base64')])
      expect(() => authenticateUpdaterSignature(malformed, key.publicKey, '1.0.0')).toThrow()
  })
})

describe('signed release versions', () => {
  it('refuses legacy, missing, duplicate and replayed signed versions', () => {
    const key = signer()
    const payload = Buffer.from('genuine older artifact')
    for (const comment of [
      'timestamp:123\tfile:fixture',
      'timestamp:123\tfile:fixture\tversion:',
      'timestamp:123\tfile:fixture\tversion:0.9.0',
      'timestamp:123\tfile:fixture\tversion:1.0.0\tversion:0.9.0',
    ]) {
      const signature = key.sign(payload, comment)
      expect(() => verifyUpdaterPayload(payload, signature, key.publicKey, '1.0.0')).toThrow(/signed version/)
      const signatures = new Map(names.map((name) => [name, signature]))
      expect(() => buildUpdaterFeed(release(), signatures, key.publicKey)).toThrow(/signed version/)
    }
  })

  it('accepts the exact signed prerelease and rejects an inflated advertised version', () => {
    const key = signer()
    const payload = Buffer.from('candidate')
    const signature = key.sign(payload, 'timestamp:123\tfile:fixture\tversion:1.0.0-beta.rc.5')
    expect(() => verifyUpdaterPayload(payload, signature, key.publicKey, '1.0.0-beta.rc.5')).not.toThrow()
    expect(() => verifyUpdaterPayload(payload, signature, key.publicKey, '1.0.0')).toThrow(/signed version/)
    expect(() => verifyUpdaterPayload(payload, signature, key.publicKey, '1.0.0-beta.rc.6')).toThrow(/signed version/)
  })
})

describe('published stable feed', () => {
  it('returns no catalog for drafts, prereleases and an empty release list', () => {
    expect(publishedUpdaterRelease([])).toBeUndefined()
    const draft = { ...release(), draft: true }
    const beta = { ...release(), tag_name: 'v1.1.0-beta.1', prerelease: true }
    expect(publishedUpdaterRelease([draft, beta])).toBeUndefined()
    expect(buildUpdaterFeed(undefined, new Map(), '')).toBeNull()
    for (const invalid of [draft, beta, { ...release(), published_at: null }, { ...release(), published_at: 'invalid' }])
      expect(() => buildUpdaterFeed(invalid, new Map(), '')).toThrow(/published stable/)
  })

  it('requires a complete, authenticated platform matrix from exact repository asset URLs', () => {
    const key = signer()
    const signatures = new Map(names.map((name) => [name, key.sign(Buffer.from(name))]))
    const candidate = release()
    const feed = buildUpdaterFeed(candidate, signatures, key.publicKey)!
    expect(feed.version).toBe('1.0.0')
    expect(Object.keys(feed.platforms).sort()).toEqual([
      'darwin-aarch64', 'darwin-x86_64', 'linux-x86_64',
      'linux-x86_64-deb', 'linux-x86_64-rpm',
    ])
    for (const missing of names) {
      expect(() => buildUpdaterFeed({ ...candidate, assets: candidate.assets.filter((asset) => asset.name !== missing) }, signatures, key.publicKey)).toThrow(/Missing updater target/)
      const incomplete = new Map(signatures)
      incomplete.delete(missing)
      expect(() => buildUpdaterFeed(candidate, incomplete, key.publicKey)).toThrow(/Missing updater signature/)
    }
    expect(() => buildUpdaterFeed(candidate, new Map(), key.publicKey)).toThrow(/Missing updater signature/)
    expect(() => buildUpdaterFeed({ ...candidate, assets: candidate.assets.slice(2) }, signatures, key.publicKey)).toThrow(/Missing updater target/)
    expect(() => buildUpdaterFeed({ ...candidate, assets: [...candidate.assets, candidate.assets[0]!] }, signatures, key.publicKey)).toThrow(/Duplicate/)
    const foreign = structuredClone(candidate)
    foreign.assets[0]!.browser_download_url = 'https://attacker.example.com/payload'
    expect(() => buildUpdaterFeed(foreign, signatures, key.publicKey)).toThrow(/Untrusted/)
    const empty = structuredClone(candidate)
    empty.assets[0]!.size = 0
    expect(() => buildUpdaterFeed(empty, signatures, key.publicKey)).toThrow(/empty/)
    expect(() => buildUpdaterFeed(candidate, signatures, signer().publicKey)).toThrow()
  })

  it('offers the matching signed installer to native Linux package clients', () => {
    const key = signer()
    const signatures = new Map(names.map((name) => [name, key.sign(Buffer.from(name))]))
    const feed = buildUpdaterFeed(release(), signatures, key.publicKey)!
    // Tauri tries the installed format first, then the generic OS/architecture.
    // deb/rpm clients must not fall back to the generic AppImage payload.
    for (const [format, name] of [
      ['appimage', names[2]!], ['deb', names[3]!], ['rpm', names[4]!],
    ] as const) {
      const target = `linux-x86_64-${format}` as keyof typeof feed.platforms
      const entry = feed.platforms[target] ?? feed.platforms['linux-x86_64']
      expect(entry.url).toBe(`https://github.com/dcuellar322/envarium/releases/download/v1.0.0/${name}`)
      expect(entry.signature).toBe(signatures.get(name))
      expect(() => verifyUpdaterPayload(Buffer.from(name), entry.signature, key.publicKey, '1.0.0')).not.toThrow()
    }
  })

  it('ignores unsupported targets and CLI downloads', () => {
    for (const name of ['README.txt', 'envarium_Darwin_arm64.tar.gz', 'Envarium_universal.app.tar.gz', 'Envarium_arm64.AppImage', 'Envarium_arm64.deb', 'Envarium_arm64.rpm', 'Envarium_x64.msi', 'Envarium_1.0.0_x64-setup.exe'])
      expect(updaterAssetTarget(name)).toBeUndefined()
    const key = signer()
    const signatures = new Map(names.map((name) => [name, key.sign(Buffer.from(name))]))
    const candidate = release()
    const windows = { name: 'Envarium_1.0.0_x64-setup.exe', size: 100, browser_download_url: 'https://example.test/windows.exe' }
    expect(buildUpdaterFeed({ ...candidate, assets: [...candidate.assets, windows] }, signatures, key.publicKey)).toEqual(buildUpdaterFeed(candidate, signatures, key.publicKey))
  })
})

describe('Pages updater response', () => {
  async function worker(feed: ReturnType<typeof buildUpdaterFeed>) {
    const source = pagesWorkerSource(feed)
    return (await import(/* @vite-ignore */ `data:text/javascript;base64,${Buffer.from(source).toString('base64')}`)).default
  }
  it('returns HTTP 204 with no synthetic version before publication; limits methods', async () => {
    const handler = await worker(null)
    const request = new Request('https://envarium.davidcuellar.tech/updates/stable.json')
    const response = await handler.fetch(request, {})
    expect(response.status).toBe(204)
    expect(await response.text()).toBe('')
    expect(response.headers.get('Cache-Control')).toBe('no-store')
    expect((await handler.fetch(new Request(request, { method: 'POST' }), {})).status).toBe(405)
    expect((await handler.fetch(new Request(request, { method: 'HEAD' }), {})).status).toBe(204)
    const assets = { fetch: () => new Response('static') }
    expect(await (await handler.fetch(new Request('https://example.com/docs/'), { ASSETS: assets })).text()).toBe('static')
  })
  it('serves the generated catalog and bodyless HEAD after publication', async () => {
    const key = signer()
    const feed = buildUpdaterFeed(release(), new Map(names.map((name) => [name, key.sign(Buffer.from(name))])), key.publicKey)!
    const handler = await worker(feed)
    const request = new Request('https://envarium.davidcuellar.tech/updates/stable.json')
    expect(await (await handler.fetch(request, {})).json()).toEqual(feed)
    const head = await handler.fetch(new Request(request, { method: 'HEAD' }), {})
    expect(head.status).toBe(200)
    expect(await head.text()).toBe('')
  })
})

describe('desktop artifact collection names', () => {
  it('separates Mac architectures and their signatures without uploading raw bundles', () => {
    expect(desktopArtifactName('Envarium.app.tar.gz', 'macos-arm64', '1.0.0')).toBe(names[0])
    expect(desktopArtifactName('Envarium.app.tar.gz.sig', 'macos-amd64', '1.0.0')).toBe(`${names[1]}.sig`)
    expect(desktopArtifactName('Envarium.app.tar.gz', 'macos-arm64', '1.0.0-beta.rc.1')).toBe('Envarium_1.0.0-beta.rc.1_aarch64.app.tar.gz')
    expect(desktopArtifactName('Envarium_1.0.0_amd64.AppImage', 'linux-amd64', '1.0.0')).toBe(names[2])
    expect(desktopArtifactName('Envarium.app', 'macos-amd64', '1.0.0')).toBeUndefined()
    expect(() => desktopArtifactName('Envarium.app.tar.gz', 'linux-amd64', '1.0.0')).toThrow()
    expect(() => desktopArtifactName(names[2]!, 'windows-amd64', '1.0.0')).toThrow()
    expect(() => desktopArtifactName(names[2]!, 'unknown', '1.0.0')).toThrow()
    expect(desktopArtifactName('envarium_Linux_x86_64.tar.gz', 'linux-amd64', '1.0.0')).toBeUndefined()
  })
})
