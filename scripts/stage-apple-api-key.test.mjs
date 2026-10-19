import assert from 'node:assert/strict'
import { generateKeyPairSync } from 'node:crypto'
import {
  existsSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  statSync,
  symlinkSync,
  writeFileSync,
} from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import { stageAppleApiKey } from './stage-apple-api-key.mjs'

const { privateKey } = generateKeyPairSync('ec', { namedCurve: 'prime256v1' })
const pem = privateKey.export({ type: 'pkcs8', format: 'pem' })

function fixture(t) {
  const directory = mkdtempSync(join(tmpdir(), 'envarium-apple-key-'))
  t.after(() => rmSync(directory, { recursive: true, force: true }))
  return {
    APPLE_API_KEY: 'EXAMPLE123',
    APPLE_API_ISSUER: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee',
    APPLE_API_KEY_BASE64: Buffer.from(pem).toString('base64'),
    RUNNER_TEMP: directory,
    GITHUB_ENV: join(directory, 'environment'),
  }
}

test('stages a private key with owner-only permissions and exports only its path', (t) => {
  const env = fixture(t)
  const target = stageAppleApiKey(env)
  assert.equal(readFileSync(target, 'utf8'), pem)
  assert.equal(statSync(target).mode & 0o777, 0o600)
  assert.equal(readFileSync(env.GITHUB_ENV, 'utf8'), `APPLE_API_KEY_PATH=${target}\n`)
  assert.throws(() => stageAppleApiKey(env), { code: 'EEXIST' })
})

test('rejects absent or malformed credentials without writing files', (t) => {
  for (const patch of [
    { APPLE_API_KEY: '' },
    { APPLE_API_ISSUER: '' },
    { APPLE_API_KEY_BASE64: '' },
    { APPLE_API_KEY: 'EXAMPLE123\n' },
    { APPLE_API_KEY_BASE64: 'bad!' },
    { APPLE_API_KEY_BASE64: 'A'.repeat(20_000) },
    { APPLE_API_KEY_BASE64: Buffer.from('invalid private fixture').toString('base64') },
    { RUNNER_TEMP: 'relative' },
    { GITHUB_ENV: '' },
  ]) {
    const env = { ...fixture(t), ...patch }
    assert.throws(() => stageAppleApiKey(env))
    assert.equal(existsSync(env.GITHUB_ENV), false)
  }
})

test('does not overwrite symlinks and removes its key if environment setup fails', (t) => {
  const env = fixture(t)
  const unrelated = join(env.RUNNER_TEMP, 'unrelated')
  writeFileSync(unrelated, 'unchanged')
  const target = join(env.RUNNER_TEMP, 'envarium-notarization.p8')
  symlinkSync(unrelated, target)
  assert.throws(() => stageAppleApiKey(env), { code: 'EEXIST' })
  assert.equal(readFileSync(unrelated, 'utf8'), 'unchanged')
  rmSync(target)
  assert.throws(() => stageAppleApiKey({ ...env, GITHUB_ENV: env.RUNNER_TEMP }))
  assert.equal(existsSync(target), false)
})
