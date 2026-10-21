import { updaterPublicKey } from '../src/lib/releases/updaterConfig'
import { createHash } from 'node:crypto'
import { readdir, readFile, stat, writeFile } from 'node:fs/promises'
import { basename, resolve } from 'node:path'
import { validateReleaseFileSet } from '../src/lib/releases/releaseFiles'
import { updaterAssetTarget, updaterTargets } from '../src/lib/releases/updater'
import { verifyUpdaterPayload } from '../src/lib/releases/updaterSignature'
import type { ReleaseChannelName } from '../src/lib/releases/types'
import { repositoryRoot } from './paths'

async function filesUnder(directory: string): Promise<string[]> {
  const entries = await readdir(directory, { withFileTypes: true })
  return (
    await Promise.all(
      entries.map((entry) => {
        const path = resolve(directory, entry.name)
        return entry.isDirectory() ? filesUnder(path) : [path]
      }),
    )
  ).flat()
}

async function sha256(path: string): Promise<string> {
  return createHash('sha256')
    .update(await readFile(path))
    .digest('hex')
}

const directoryArgument = process.argv[2]
const channelArgument = process.argv[3]
if (!directoryArgument || !channelArgument) {
  throw new Error('Usage: prepare-release-files <directory> <stable|alpha|beta|nightly>')
}
if (!['stable', 'alpha', 'beta', 'nightly'].includes(channelArgument)) {
  throw new Error(`Unknown release channel: ${channelArgument}`)
}

const releaseDirectory = resolve(repositoryRoot, directoryArgument)
if (!(await stat(releaseDirectory)).isDirectory())
  throw new Error(`${releaseDirectory} is not a directory`)

const metadata = JSON.parse(await readFile(resolve(releaseDirectory, 'metadata.json'), 'utf8')) as { version?: string; tag?: string }
if (typeof metadata.version !== 'string' || !metadata.version)
  throw new Error('Release metadata must provide a version')

const checksumPath = resolve(releaseDirectory, 'checksums.txt')
const files = (await filesUnder(releaseDirectory)).filter((path) => path !== checksumPath)
const fileNames = files.map((path) => basename(path))
const checksumEntries = await Promise.all(
  files.map(async (path) => `${await sha256(path)}  ${basename(path)}`),
)
const checksummedNames = new Set(fileNames)

// Recheck signed payloads after matrix collection, before attestation or a draft.
const publicKey = updaterPublicKey
const targets = new Set<string>()
for (const path of files) {
  const target = updaterAssetTarget(basename(path))
  if (!target) continue
  // CLI-only GoReleaser snapshots intentionally have a version unlike their tag.
  // Native updater payloads must always bind the exact tagged version.
  if (metadata.tag !== `v${metadata.version}`)
    throw new Error('Updater release metadata must provide the exact tagged version')
  if (targets.has(target)) throw new Error(`Duplicate updater payload for ${target}`)
  targets.add(target)
  verifyUpdaterPayload(await readFile(path), await readFile(`${path}.sig`, 'utf8'), publicKey, metadata.version)
}
if (channelArgument === 'stable') {
  for (const target of updaterTargets)
    if (!targets.has(target)) throw new Error(`Stable release is missing updater payload: ${target}`)
}

validateReleaseFileSet({
  channel: channelArgument as ReleaseChannelName,
  fileNames,
  checksummedNames,
})

await writeFile(checksumPath, `${checksumEntries.sort().join('\n')}\n`, 'utf8')
console.log(
  `Validated ${fileNames.length} release files for ${channelArgument} and wrote aggregate checksums.`,
)
