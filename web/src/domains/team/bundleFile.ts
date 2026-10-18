import type { TeamBundle, TeamBundleKind } from '../../api/generated/types.gen'

const bundleKinds: ReadonlyArray<TeamBundleKind> = [
  'project-template',
  'policy-pack',
  'plugin-registry',
  'enterprise-config',
]

function record(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function reviewablePayload(value: Record<string, unknown>): void {
  const pending: Array<{ value: unknown; depth: number }> = [{ value, depth: 0 }]
  while (pending.length) {
    const item = pending.pop()!
    if (item.depth > 32)
      throw new Error('The bundle payload is too deeply nested to review safely.')
    if (typeof item.value === 'object' && item.value !== null)
      for (const child of Object.values(item.value))
        pending.push({ value: child, depth: item.depth + 1 })
  }
}

// Check the display envelope, not trust or policy semantics. The application
// still verifies the complete signed payload before any installation.
export function parseBundleFile(text: string): TeamBundle {
  const value: unknown = JSON.parse(text)
  if (
    !record(value) ||
    value.schemaVersion !== 'envarium.bundle/v1' ||
    !bundleKinds.includes(value.kind as TeamBundleKind) ||
    !record(value.metadata) ||
    !['id', 'name', 'version', 'publisherId', 'createdAt'].every(
      (key) => typeof (value.metadata as Record<string, unknown>)[key] === 'string',
    ) ||
    !record(value.payload) ||
    !record(value.signature) ||
    value.signature.algorithm !== 'Ed25519' ||
    typeof value.signature.keyId !== 'string' ||
    typeof value.signature.value !== 'string'
  ) {
    throw new Error(
      'Choose a signed Envarium bundle with metadata, payload, and an Ed25519 signature.',
    )
  }
  reviewablePayload(value.payload)
  return value as TeamBundle
}
