import { updaterPublicKey } from '../src/lib/releases/updaterConfig'
import { execFileSync } from 'node:child_process'
import { copyFile, mkdir, readFile, readdir } from 'node:fs/promises'
import { resolve } from 'node:path'
import { desktopArtifactName } from '../src/lib/releases/desktopArtifacts'
import { verifyUpdaterPayload } from '../src/lib/releases/updaterSignature'
import { repositoryRoot } from './paths'

const platform = process.argv[2]
if (!platform) throw new Error('Specify the desktop matrix artifact platform')
const bundle = resolve(repositoryRoot, 'desktop/src-tauri/target/release/bundle')
const output = resolve(repositoryRoot, 'desktop-artifacts')
const version = process.env.ENVARIUM_VERSION || (JSON.parse(await readFile(resolve(repositoryRoot, 'desktop/src-tauri/tauri.conf.json'), 'utf8')) as { version: string }).version
const publicKey = updaterPublicKey
await mkdir(output, { recursive: false })
let count = 0
let updaterCount = 0
for (const directory of await readdir(bundle, { withFileTypes: true })) {
  if (!directory.isDirectory() || !['macos', 'dmg', 'appimage', 'deb', 'rpm', 'msi', 'nsis'].includes(directory.name)) continue
  const directoryPath = resolve(bundle, directory.name)
  for (const entry of await readdir(directoryPath, { withFileTypes: true })) {
    if (!entry.isFile()) continue
    const name = desktopArtifactName(entry.name, platform, version)
    if (!name) continue
    const path = resolve(directoryPath, entry.name)
    if (/\.(app\.tar\.gz|AppImage|msi)$/.test(entry.name) || entry.name.endsWith('-setup.exe')) {
      verifyUpdaterPayload(await readFile(path), await readFile(`${path}.sig`, 'utf8'), publicKey)
      updaterCount++
    }
    await copyFile(path, resolve(output, name))
    count++
    if (platform === 'linux-amd64' && !name.endsWith('.sig')) {
      const fingerprint = (await readFile(resolve(repositoryRoot, 'site/public/signing/envarium-linux-fingerprint.txt'), 'utf8')).trim()
      if (process.env.SIGN_KEY?.trim() !== fingerprint || !process.env.APPIMAGETOOL_SIGN_PASSPHRASE)
        throw new Error('Linux release signing configuration is missing or mismatched')
      const artifact = resolve(output, name)
      execFileSync('gpg', ['--batch', '--yes', '--pinentry-mode', 'loopback', '--passphrase-fd', '0', '--local-user', fingerprint, '--armor', '--detach-sign', artifact], {
        input: process.env.APPIMAGETOOL_SIGN_PASSPHRASE, stdio: ['pipe', 'pipe', 'pipe'],
      })
      const result = execFileSync('gpg', ['--batch', '--status-fd', '1', '--verify', `${artifact}.asc`, artifact], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] })
      if (!result.includes(`[GNUPG:] VALIDSIG ${fingerprint} `)) throw new Error('Linux artifact signature verification failed')
    }
  }
}
if (!count || updaterCount !== (platform === 'windows-amd64' ? 2 : 1))
  throw new Error('Desktop updater artifact set is incomplete')
console.log(`Collected ${count} installer/signature files; verified ${updaterCount} updater payloads for ${platform}.`)
