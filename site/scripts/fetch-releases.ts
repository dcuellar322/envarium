import { updaterPublicKey } from '../src/lib/releases/updaterConfig'
import { readFile, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { normalizeManifest, validatePublishedStable } from '../src/lib/releases/normalize'
import type { GitHubRelease } from '../src/lib/releases/types'
import { repositoryRoot, siteRoot } from './paths'
import { buildUpdaterFeed, publishedUpdaterRelease, trustedReleaseAssetURL, updaterAssetTarget } from '../src/lib/releases/updater'

const repository = 'dcuellar322/envarium'
const headers = {
  Accept: 'application/vnd.github+json',
  'X-GitHub-Api-Version': '2022-11-28',
  'User-Agent': 'envarium-site-build',
  ...(process.env.GITHUB_TOKEN ? { Authorization: `Bearer ${process.env.GITHUB_TOKEN}` } : {}),
}

async function fetchRequired(url: string): Promise<Response> {
  const response = await fetch(url, { headers, signal: AbortSignal.timeout(30_000) })
  if (!response.ok) throw new Error(`GitHub request failed with ${response.status} for ${url}`)
  return response
}

async function main(): Promise<void> {
  const response = await fetchRequired(`https://api.github.com/repos/${repository}/releases?per_page=30`)
  const releases = (await response.json()) as GitHubRelease[]
  const checksumBodies = new Map<string, string>()
  for (const release of releases) {
    if (release.draft) continue
    const checksumAsset = release.assets.find((asset) => asset.name === 'checksums.txt')
    if (!checksumAsset) continue
    const checksumResponse = await fetchRequired(checksumAsset.browser_download_url)
    checksumBodies.set(release.tag_name, await checksumResponse.text())
  }
  const manifest = normalizeManifest(releases, checksumBodies)
  validatePublishedStable(manifest)
  const stable = publishedUpdaterRelease(releases)
  const signatures = new Map<string, string>()
  for (const asset of stable?.assets ?? []) {
    if (!updaterAssetTarget(asset.name)) continue
    const signature = stable?.assets.find((candidate) => candidate.name === `${asset.name}.sig`)
    if (!stable || !signature || signature.size > 8192) throw new Error('Missing or oversized updater signature')
    const response = await fetchRequired(trustedReleaseAssetURL(signature, stable.tag_name))
    const body = await response.text()
    if (body.length > 8192) throw new Error('Oversized updater signature response')
    signatures.set(asset.name, body)
  }
  const publicKey = updaterPublicKey
  const updater = buildUpdaterFeed(stable, signatures, publicKey)
  await writeFile(resolve(siteRoot, 'src/data/updater-manifest.json'), `${JSON.stringify(updater, null, 2)}\n`)
  await writeFile(
    resolve(siteRoot, 'src/data/release-manifest.json'),
    `${JSON.stringify(manifest, null, 2)}\n`,
    'utf8',
  )
}

if (process.argv.includes('--fixture')) {
  const fixturePath = process.argv[process.argv.indexOf('--fixture') + 1]
  if (!fixturePath) throw new Error('--fixture requires a path')
  const fixture = JSON.parse(await readFile(resolve(repositoryRoot, fixturePath), 'utf8')) as GitHubRelease[] | { releases: GitHubRelease[]; checksums?: Record<string, string> }
  const releases = Array.isArray(fixture) ? fixture : fixture.releases
  const checksums = new Map(Object.entries(Array.isArray(fixture) ? {} : fixture.checksums ?? {}))
  const manifest = normalizeManifest(releases, checksums)
  validatePublishedStable(manifest)
  // Download-page fixtures must never leave an authoritative update catalog.
  await writeFile(resolve(siteRoot, 'src/data/updater-manifest.json'), 'null\n')
  await writeFile(resolve(siteRoot, 'src/data/release-manifest.json'), `${JSON.stringify(manifest, null, 2)}\n`)
} else {
  await main()
}
