import { mkdir, rm, writeFile } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'
import releaseManifest from '../src/data/release-manifest.json' with { type: 'json' }
import { generatePackageDrafts } from '../src/lib/distribution/generate'
import type { ReleaseManifest } from '../src/lib/releases/types'
import { siteRoot } from './paths'

const manifest = releaseManifest as ReleaseManifest
const stable = manifest.channels.stable
const expectedTagIndex = process.argv.indexOf('--expect-tag')
const expectedTag = expectedTagIndex >= 0 ? process.argv[expectedTagIndex + 1] : undefined
if (!stable) throw new Error('Package drafts require a published stable release')
if (expectedTag && stable.tag !== expectedTag) throw new Error(`Refusing package generation: expected ${expectedTag}, release manifest contains ${stable.tag}`)

const output = resolve(siteRoot, '.generated/package-drafts')
const drafts = generatePackageDrafts(stable)
await rm(output, { force: true, recursive: true })
for (const [name, body] of Object.entries(drafts)) {
  const path = resolve(output, name)
  await mkdir(dirname(path), { recursive: true })
  await writeFile(path, body)
}
console.log(`Generated reviewed package drafts for ${stable.tag} in ${output}`)
