import { spawnSync } from 'node:child_process'
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { expect, it } from 'vitest'

it('accepts CLI-only snapshots but fails closed on mismatched native updater metadata', () => {
  const directory = mkdtempSync(join(tmpdir(), 'envarium-release-preparation-'))
  const site = resolve(import.meta.dirname, '../..')
  function prepare() {
    return spawnSync(process.execPath, ['--import', 'tsx', 'scripts/prepare-release-files.ts', directory, 'alpha'], {
      cwd: site, encoding: 'utf8', timeout: 15_000,
    })
  }
  try {
    writeFileSync(join(directory, 'metadata.json'), JSON.stringify({ tag: 'v1.0.0', version: '1.0.1-next' }))
    writeFileSync(join(directory, 'envarium_Linux_x86_64.tar.gz'), 'disposable CLI archive fixture')
    expect(prepare().status).toBe(0)
    const checksumPath = join(directory, 'checksums.txt')
    const before = readFileSync(checksumPath, 'utf8')
    expect(before).toContain('envarium_Linux_x86_64.tar.gz')
    writeFileSync(join(directory, 'Envarium_1.0.1-next_amd64.AppImage'), 'disposable native artifact fixture')
    const mismatch = prepare()
    expect(mismatch.status).not.toBe(0)
    expect(mismatch.stderr).toContain('exact tagged version')
    expect(readFileSync(checksumPath, 'utf8')).toBe(before)
    writeFileSync(join(directory, 'metadata.json'), JSON.stringify({ tag: 'v1.0.1-next', version: '1.0.1-next' }))
    const unsigned = prepare()
    expect(unsigned.status).not.toBe(0)
    expect(unsigned.stderr).toContain('.AppImage.sig')
    expect(readFileSync(checksumPath, 'utf8')).toBe(before)
  } finally {
    rmSync(directory, { recursive: true, force: true })
  }
}, 30_000)
