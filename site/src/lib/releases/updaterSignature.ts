import { createHash, createPublicKey, verify } from 'node:crypto'

function decode(value: string): Buffer {
  if (!value || value.length > 8192 || !/^[A-Za-z0-9+/=\s]+$/.test(value))
    throw new Error('Invalid updater signature encoding')
  return Buffer.from(value.trim(), 'base64')
}

/** Verify Minisign's authenticated signature metadata; payload verification is separate. */
export function authenticateUpdaterSignature(encoded: string, publicKey: string): Buffer {
  const keyLines = decode(publicKey).toString('utf8').trim().split('\n')
  const lines = decode(encoded).toString('utf8').trim().split('\n')
  const key = Buffer.from(keyLines[1] ?? '', 'base64')
  const signature = Buffer.from(lines[1] ?? '', 'base64')
  const globalSignature = Buffer.from(lines[3] ?? '', 'base64')
  if (keyLines.length !== 2 || lines.length !== 4 || key.length !== 42 ||
      key.subarray(0, 2).toString() !== 'Ed' || signature.length !== 74 ||
      signature.subarray(0, 2).toString() !== 'ED' || globalSignature.length !== 64 ||
      !signature.subarray(2, 10).equals(key.subarray(2, 10)) ||
      !lines[2]?.startsWith('trusted comment: '))
    throw new Error('Invalid updater signature or signing key')
  const ed25519 = createPublicKey({
    key: Buffer.concat([Buffer.from('302a300506032b6570032100', 'hex'), key.subarray(10)]),
    format: 'der', type: 'spki',
  })
  const metadata = Buffer.concat([signature.subarray(10), Buffer.from(lines[2].slice(17))])
  if (!verify(null, metadata, ed25519, globalSignature))
    throw new Error('Updater signature metadata authentication failed')
  return signature.subarray(10)
}

/** Match the prehashed Ed25519 format emitted by the pinned Tauri signer. */
export function verifyUpdaterPayload(payload: Uint8Array, signature: string, publicKey: string): void {
  const messageSignature = authenticateUpdaterSignature(signature, publicKey)
  const key = Buffer.from(decode(publicKey).toString('utf8').trim().split('\n')[1] ?? '', 'base64')
  const ed25519 = createPublicKey({
    key: Buffer.concat([Buffer.from('302a300506032b6570032100', 'hex'), key.subarray(10)]),
    format: 'der', type: 'spki',
  })
  if (!verify(null, createHash('blake2b512').update(payload).digest(), ed25519, messageSignature))
    throw new Error('Updater payload signature verification failed')
}
