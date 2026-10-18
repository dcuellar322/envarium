import assert from 'node:assert/strict'
import { readFile, readdir } from 'node:fs/promises'
import { createHash } from 'node:crypto'
import { resolve } from 'node:path'
import { classifyReleaseAsset, missingStableArtifacts } from '../src/lib/releases/artifactContract'
import { normalizeManifest } from '../src/lib/releases/normalize'

// Run against real GoReleaser snapshot output, after prepare-release-files.
const directory = resolve(process.argv[2] ?? '../dist')
const names = (await readdir(directory)).filter((name) => classifyReleaseAsset(name))
assert.equal(names.length, 6, 'all six CLI platform/architecture archives must be recognized')
const checksums = await readFile(resolve(directory, 'checksums.txt'), 'utf8')
for (const name of names) {
  const digest = createHash('sha256')
    .update(await readFile(resolve(directory, name)))
    .digest('hex')
  assert.ok(checksums.includes(`${digest}  ${name}`), `checksum mismatch for ${name}`)
}
const manifest = normalizeManifest(
  [
    {
      tag_name: 'v1.0.0-alpha.1',
      html_url: 'https://example.test/release',
      published_at: null,
      created_at: '2026-10-18T00:00:00Z',
      prerelease: true,
      draft: false,
      assets: names.map((name) => ({
        name,
        size: 1,
        browser_download_url: `https://example.test/${name}`,
      })),
    },
  ],
  new Map([['v1.0.0-alpha.1', checksums]]),
)
assert.equal(manifest.channels.alpha?.assets.length, 6)
assert.ok(manifest.channels.alpha?.assets.every((asset) => asset.checksum))
const missing = missingStableArtifacts(manifest.channels.alpha?.assets ?? [])
assert.equal(
  missing.length,
  6,
  'CLI snapshot must still require all six desktop packages for stable release',
)
assert.ok(missing.every((label) => !label.includes('CLI archive')))
console.log(
  'Real CLI snapshot, aggregate checksums, alpha normalization and stable desktop gate passed.',
)
