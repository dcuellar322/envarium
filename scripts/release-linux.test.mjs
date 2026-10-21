import assert from 'node:assert/strict'
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import test from 'node:test'
import { buildLinuxRelease, rpmMetadata } from './release-linux.mjs'

const version = '1.0.0-beta.rc.5'
function fixture() {
  const root = mkdtempSync(join(tmpdir(), 'envarium-rpm-package-test-'))
  const project = join(root, 'desktop/src-tauri')
  const binary = join(project, 'target/release/envarium-desktop')
  const sidecar = join(project, 'binaries/envarium-x86_64-unknown-linux-gnu')
  const directory = join(project, 'target/release/bundle/rpm')
  const calls = []
  function run(args) {
    calls.push(args)
    if (args[0] === 'build') {
      mkdirSync(join(project, 'target/release'), { recursive: true })
      mkdirSync(join(project, 'binaries'))
      writeFileSync(binary, `compiled product ${version} __TAURI_BUNDLE_TYPE_VAR_UNK`)
      writeFileSync(sidecar, `compiled Go sidecar ${version}`)
    } else if (args[0] === 'bundle') {
      const config = JSON.parse(args[4])
      mkdirSync(directory, { recursive: true })
      writeFileSync(join(directory, `Envarium-${config.version}-${config.bundle.linux.rpm.release}.x86_64.rpm`), 'disposable package fixture')
    }
  }
  return { root, binary, sidecar, directory, calls, run, cleanup() { rmSync(root, { recursive: true, force: true }) } }
}

test('maps supported product versions to valid RPM metadata without labeling a candidate stable', () => {
  assert.deepEqual(rpmMetadata('1.0.0'), { version: '1.0.0', release: '1' })
  assert.deepEqual(rpmMetadata(version), { version: '1.0.0', release: '0.beta.rc.5' })
  assert.deepEqual(rpmMetadata('1.2.3-alpha.2'), { version: '1.2.3', release: '0.alpha.2' })
  assert.deepEqual(rpmMetadata('1.2.3-beta.rc-2'), { version: '1.2.3', release: '0.beta.rc_2' })
  assert.deepEqual(rpmMetadata('1.2.3-nightly.20261021'), { version: '1.2.3', release: '0.nightly.20261021' })
  for (const value of ['', undefined, 'v1.0.0', '01.0.0', '1.0.0+build', '1.0.0-beta.', '1.0.0-other.1', '1.0.0-beta.1\n'])
    assert.throws(() => rpmMetadata(value), /Unsupported/)
})

test('builds once, bundles RPM metadata separately, then signs the full product version', () => {
  const f = fixture()
  try {
    buildLinuxRelease({ root: f.root, version, run: f.run })
    assert.deepEqual(f.calls.map((args) => args[0]), ['build', 'bundle', 'signer'])
    assert.deepEqual(JSON.parse(f.calls[0][4]), { version })
    assert.equal(f.calls[0][6], 'appimage,deb')
    assert.deepEqual(JSON.parse(f.calls[1][4]), { version: '1.0.0', bundle: { createUpdaterArtifacts: false, linux: { rpm: { release: '0.beta.rc.5' } } } })
    assert.equal(f.calls[1][6], 'rpm')
    assert.deepEqual(f.calls[2].slice(0, 4), ['signer', 'sign', '--app-version', version])
    assert.equal(f.calls[2][4], join(f.directory, 'Envarium-1.0.0-0.beta.rc.5.x86_64.rpm'))
    assert.match(readFileSync(f.binary, 'utf8'), /compiled product 1\.0\.0-beta\.rc\.5.*_VAR_UNK/)
    assert.equal(readFileSync(f.sidecar, 'utf8'), `compiled Go sidecar ${version}`)
  } finally { f.cleanup() }
})

test('refuses unexpected native or sidecar changes and never signs them', () => {
  for (const mutation of ['native', 'sidecar', 'marker', 'extra-package']) {
    const f = fixture()
    try {
      assert.throws(() => buildLinuxRelease({ root: f.root, version, run(args) {
        f.run(args)
        if (args[0] !== 'bundle') return
        if (mutation === 'native') writeFileSync(f.binary, readFileSync(f.binary).toString().replace('compiled product', 'different product'))
        if (mutation === 'sidecar') writeFileSync(f.sidecar, 'rebuilt unexpected sidecar')
        if (mutation === 'marker') writeFileSync(f.binary, readFileSync(f.binary).toString().replace('_VAR_UNK', '_VAR_DEB'))
        if (mutation === 'extra-package') writeFileSync(join(f.directory, 'unexpected.rpm'), 'unexpected')
      } }), /changed product|wrong updater|ambiguous/)
      assert.equal(f.calls.some((args) => args[0] === 'signer'), false)
    } finally { f.cleanup() }
  }
})

test('stops on command failure or cancellation before any later packaging/signing command', () => {
  for (const phase of ['build', 'bundle', 'signer']) {
    const f = fixture()
    try {
      assert.throws(() => buildLinuxRelease({ root: f.root, version, run(args) {
        f.run(args)
        if (args[0] === phase) throw new Error('command failed or cancelled')
      } }), /failed or cancelled/)
      assert.equal(f.calls.at(-1)[0], phase)
    } finally { f.cleanup() }
  }
})
