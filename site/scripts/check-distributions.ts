import { readFile, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import releaseManifest from '../src/data/release-manifest.json' with { type: 'json' }
import { packageRecordMatches } from '../src/lib/distribution/status'
import type { DistributionChannel, DistributionStatus } from '../src/lib/distribution/types'
import type { ReleaseManifest } from '../src/lib/releases/types'
import { siteRoot } from './paths'

const statusPath = resolve(siteRoot, 'src/data/distribution-status.json')
const current = JSON.parse(await readFile(statusPath, 'utf8')) as DistributionStatus
const stable = (releaseManifest as ReleaseManifest).channels.stable

async function checked(channel: DistributionChannel, url: string, platform: 'macos' | 'windows'): Promise<DistributionChannel> {
  if (!stable) return { ...channel, available: false, verifiedVersion: undefined, lastChecked: undefined }
  const response = await fetch(url, { headers: { 'User-Agent': 'envarium-distribution-monitor' } })
  const available = response.ok && packageRecordMatches(await response.text(), stable, platform)
  const same = channel.available === available && channel.verifiedVersion === (available ? stable.version : undefined)
  return {
    ...channel,
    available,
    verifiedVersion: available ? stable.version : undefined,
    lastChecked: same ? channel.lastChecked : new Date().toISOString(),
  }
}

const homebrewUrl = 'https://raw.githubusercontent.com/dcuellar322/homebrew-tap/master/Casks/envarium.rb'
const next: DistributionStatus = {
  schemaVersion: 'envarium.distribution-status/v1',
  homebrew: await checked(current.homebrew, homebrewUrl, 'macos'),
  // Windows/WSL2 remain coming soon; do not query or activate WinGet yet.
  winget: { ...current.winget, available: false, verifiedVersion: undefined, lastChecked: undefined },
}
await writeFile(statusPath, `${JSON.stringify(next, null, 2)}\n`)
console.log(`Homebrew: ${next.homebrew.available ? next.homebrew.verifiedVersion : 'unavailable'}; WinGet: ${next.winget.available ? next.winget.verifiedVersion : 'unavailable'}`)
