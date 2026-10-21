import { classifyReleaseAsset } from './artifactContract'
import { authenticateUpdaterSignature } from './updaterSignature'
import type { GitHubRelease, GitHubReleaseAsset } from './types'

const repository = 'dcuellar322/envarium'
export const updaterTargets = [
  'darwin-aarch64', 'darwin-x86_64', 'linux-x86_64',
  'linux-x86_64-deb', 'linux-x86_64-rpm',
] as const
type UpdaterTarget = typeof updaterTargets[number]
export interface UpdaterFeed {
  version: string
  notes?: string
  pub_date: string
  platforms: Record<UpdaterTarget, { url: string; signature: string }>
}

export function updaterAssetTarget(name: string): UpdaterTarget | undefined {
  const asset = classifyReleaseAsset(name)
  if (!asset) return undefined
  if (asset.platform === 'macos' && asset.packageType === 'app') {
    if (asset.architecture === 'arm64') return 'darwin-aarch64'
    if (asset.architecture === 'amd64') return 'darwin-x86_64'
  }
  if (asset.architecture !== 'amd64') return undefined
  if (asset.platform === 'linux') {
    if (asset.packageType === 'appimage') return 'linux-x86_64'
    if (asset.packageType === 'deb') return 'linux-x86_64-deb'
    if (asset.packageType === 'rpm') return 'linux-x86_64-rpm'
  }
  return undefined
}

export function trustedReleaseAssetURL(asset: GitHubReleaseAsset, tag: string): string {
  const expected = `https://github.com/${repository}/releases/download/${tag}/${encodeURIComponent(asset.name)}`
  if (asset.browser_download_url !== expected || asset.size <= 0)
    throw new Error(`Untrusted or empty release asset: ${asset.name}`)
  return expected
}

export function publishedUpdaterRelease(releases: GitHubRelease[]): GitHubRelease | undefined {
  return releases.find((release) => !release.draft && !release.prerelease &&
    /^v(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)$/.test(release.tag_name) && release.published_at)
}

export function buildUpdaterFeed(
  release: GitHubRelease | undefined, signatures: ReadonlyMap<string, string>, publicKey: string,
): UpdaterFeed | null {
  if (!release) return null
  if (publishedUpdaterRelease([release]) !== release ||
      !release.published_at || !Number.isFinite(Date.parse(release.published_at)))
    throw new Error('Updater feed requires a published stable release')
  const platforms: Partial<UpdaterFeed['platforms']> = {}
  for (const asset of release.assets) {
    const target = updaterAssetTarget(asset.name)
    if (!target) continue
    if (platforms[target]) throw new Error(`Duplicate updater target: ${target}`)
    const signatureAsset = release.assets.find((candidate) => candidate.name === `${asset.name}.sig`)
    const signature = signatures.get(asset.name)?.trim()
    if (!signatureAsset || !signature) throw new Error(`Missing updater signature: ${asset.name}`)
    trustedReleaseAssetURL(signatureAsset, release.tag_name)
    authenticateUpdaterSignature(signature, publicKey, release.tag_name.slice(1))
    platforms[target] = { url: trustedReleaseAssetURL(asset, release.tag_name), signature }
  }
  for (const target of updaterTargets) {
    if (!platforms[target]) throw new Error(`Missing updater target: ${target}`)
  }
  return {
    version: release.tag_name.slice(1), notes: release.body?.slice(0, 16_384),
    pub_date: release.published_at, platforms: platforms as UpdaterFeed['platforms'],
  }
}
