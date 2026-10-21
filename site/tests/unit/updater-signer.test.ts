import { spawnSync } from 'node:child_process'
import { mkdtempSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import { verifyUpdaterPayload } from '../../src/lib/releases/updaterSignature'

// Exercise the installed official CLI with an ephemeral key, never a release key.
describe('official version-aware updater signer', () => {
  it('binds the signed candidate version and refuses unversioned or replayed metadata', () => {
    const directory = mkdtempSync(join(tmpdir(), 'envarium-signer-test-'))
    const root = resolve(import.meta.dirname, '../../..')
    const key = join(directory, 'test.key')
    const artifact = join(directory, 'fixture.AppImage')
    const version = '1.0.0-beta.rc.5'
    const payload = Buffer.from('disposable updater signer qualification payload')
    function cli(args: string[]) {
      const result = spawnSync(process.execPath, [join(root, 'node_modules/@tauri-apps/cli/tauri.js'), 'signer', ...args], {
        cwd: root, stdio: 'ignore', timeout: 15_000,
        env: { PATH: process.env.PATH, SYSTEMROOT: process.env.SYSTEMROOT, CI: 'true',
          TAURI_SIGNING_PRIVATE_KEY_PATH: key, TAURI_SIGNING_PRIVATE_KEY_PASSWORD: '' },
      })
      // Key generation may print key material; never capture or persist its output.
      expect(result.status, 'official signer command failed').toBe(0)
    }
    try {
      cli(['generate', '--ci', '--password', '', '--write-keys', key])
      if (process.platform !== 'win32') expect(statSync(key).mode & 0o777).toBe(0o600)
      const publicKey = readFileSync(`${key}.pub`, 'utf8')
      writeFileSync(artifact, payload)
      cli(['sign', '--app-version', version, artifact])
      const signature = readFileSync(`${artifact}.sig`, 'utf8')
      expect(() => verifyUpdaterPayload(payload, signature, publicKey, version)).not.toThrow()
      expect(() => verifyUpdaterPayload(payload, signature, publicKey, '1.0.0-beta.rc.6')).toThrow(/signed version/)
      expect(() => verifyUpdaterPayload(Buffer.from('tampered'), signature, publicKey, version)).toThrow(/payload signature/)
      cli(['sign', artifact])
      const legacy = readFileSync(`${artifact}.sig`, 'utf8')
      expect(() => verifyUpdaterPayload(payload, legacy, publicKey, version)).toThrow(/signed version/)
    } finally {
      rmSync(directory, { recursive: true, force: true })
    }
  }, 30_000)
})
