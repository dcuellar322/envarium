import { createPrivateKey } from 'node:crypto'
import { appendFileSync, rmSync, writeFileSync } from 'node:fs'
import { isAbsolute, join, resolve } from 'node:path'
import { pathToFileURL } from 'node:url'

// GitHub stores the key encrypted. Tauri/notarytool require a private file on
// the ephemeral macOS runner; only its path goes into the runner environment.
export function stageAppleApiKey(env) {
  if (
    env.APPLE_API_KEY?.length !== 10 ||
    !/^[A-Z0-9]{10}$/.test(env.APPLE_API_KEY) ||
    env.APPLE_API_ISSUER?.length !== 36 ||
    !/^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/i.test(env.APPLE_API_ISSUER)
  ) {
    throw new Error('Apple notarization identifiers are missing or invalid')
  }
  const encoded = env.APPLE_API_KEY_BASE64 ?? ''
  if (!encoded || encoded.length > 16_384 || !/^[A-Za-z0-9+/=\s]+$/.test(encoded)) {
    throw new Error('Apple notarization private key is missing or invalid')
  }
  const bytes = Buffer.from(encoded, 'base64')
  try {
    const key = createPrivateKey(bytes)
    if (key.asymmetricKeyType !== 'ec' || key.asymmetricKeyDetails?.namedCurve !== 'prime256v1') {
      throw new Error('Unsupported key')
    }
  } catch {
    throw new Error('Apple notarization private key must be a valid P-256 key')
  }
  if (
    !env.RUNNER_TEMP ||
    !isAbsolute(env.RUNNER_TEMP) ||
    /[\r\n]/.test(env.RUNNER_TEMP) ||
    !env.GITHUB_ENV
  ) {
    throw new Error('GitHub runner paths are missing or invalid')
  }
  const target = join(env.RUNNER_TEMP, 'envarium-notarization.p8')
  // Exclusive creation refuses a pre-existing file or symlink.
  writeFileSync(target, bytes, { flag: 'wx', mode: 0o600 })
  try {
    appendFileSync(env.GITHUB_ENV, `APPLE_API_KEY_PATH=${target}\n`)
  } catch {
    rmSync(target, { force: true })
    throw new Error('Could not configure the Apple notarization key path')
  }
  return target
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    stageAppleApiKey(process.env)
    console.log('Apple notarization key staged for this runner.')
  } catch {
    // Never print parser errors, environment values, or private key material.
    console.error(
      'Apple notarization setup failed. Check the desktop-release variables and key secret.',
    )
    process.exitCode = 1
  }
}
