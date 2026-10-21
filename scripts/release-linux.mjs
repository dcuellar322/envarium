import { createHash } from 'node:crypto'
import { readFileSync, readdirSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { spawnSync } from 'node:child_process'
import { pathToFileURL } from 'node:url'

export function rpmMetadata(version) {
  const match = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:alpha|beta|nightly)\.[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*))?$/.exec(version ?? '')
  if (!match) throw new Error('Unsupported Envarium RPM release version')
  return { version: `${match[1]}.${match[2]}.${match[3]}`, release: match[4] ? `0.${match[4].replaceAll('-', '_')}` : '1' }
}

export function buildLinuxRelease({ root, version, run }) {
  const rpm = rpmMetadata(version)
  const project = join(root, 'desktop/src-tauri')
  const config = join(project, 'tauri.release.conf.json')
  run(['build', '--config', config, '--config', JSON.stringify({ version }), '--bundles', 'appimage,deb', '--ci'])
  const executable = join(project, 'target/release/envarium-desktop')
  const before = readFileSync(executable)
  const marker = Buffer.from('__TAURI_BUNDLE_TYPE_VAR_')
  const offset = before.indexOf(marker)
  if (offset < 0 || before.indexOf(marker, offset + 1) !== -1)
    throw new Error('Native Linux bundle marker is missing or ambiguous')
  const sidecarDirectory = join(project, 'binaries')
  const sidecars = readdirSync(sidecarDirectory).filter((name) => name.startsWith('envarium-'))
  if (!sidecars.length) throw new Error('Bundled Linux sidecar is missing')
  const hash = (name) => createHash('sha256').update(readFileSync(join(sidecarDirectory, name))).digest('hex')
  const hashes = new Map(sidecars.map((name) => [name, hash(name)]))
  run(['bundle', '--config', config, '--config', JSON.stringify({ version: rpm.version,
    bundle: { createUpdaterArtifacts: false, linux: { rpm: { release: rpm.release } } } }), '--bundles', 'rpm', '--ci'])
  // Tauri restores the build output after patching the copy inside each bundle.
  const after = readFileSync(executable)
  const currentSidecars = readdirSync(sidecarDirectory).filter((name) => name.startsWith('envarium-'))
  if (!after.equals(before) || currentSidecars.length !== sidecars.length ||
      sidecars.some((name) => !currentSidecars.includes(name) || hashes.get(name) !== hash(name)))
    throw new Error('RPM bundle-only packaging changed product binaries')
  const directory = join(project, 'target/release/bundle/rpm')
  const names = readdirSync(directory).filter((name) => name.endsWith('.rpm'))
  const expected = `Envarium-${rpm.version}-${rpm.release}.x86_64.rpm`
  if (names.length !== 1 || names[0] !== expected)
    throw new Error('RPM package identity is missing or ambiguous')
  run(['signer', 'sign', '--app-version', version, join(directory, expected)])
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  const root = process.env.GITHUB_WORKSPACE
  if (!root) throw new Error('Linux release packaging requires GITHUB_WORKSPACE')
  const project = join(root, 'desktop/src-tauri')
  const cli = join(root, 'node_modules/@tauri-apps/cli/tauri.js')
  buildLinuxRelease({ root, version: process.env.ENVARIUM_VERSION, run(args) {
    const result = spawnSync(process.execPath, [cli, ...args], { cwd: project, stdio: 'inherit' })
    if (result.error || result.status !== 0) throw new Error('Official Linux build/bundle/sign command failed')
  } })
}
