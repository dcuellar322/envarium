import assert from 'node:assert/strict'
import { test } from 'node:test'
import { verifyReleaseQuality } from './release-quality.mjs'

const sha = 'a'.repeat(40)
const input = { repository: 'owner/repo', sha, token: 'fixture' }
const run = {
  id: 1,
  head_sha: sha,
  head_branch: 'master',
  status: 'completed',
  conclusion: 'success',
}
test('requires successful latest exact-commit quality runs', async () => {
  const fetcher = async () => ({ ok: true, json: async () => ({ workflow_runs: [run] }) })
  await verifyReleaseQuality({ ...input, fetcher })
  for (const runs of [
    [],
    [{ ...run, head_sha: 'b'.repeat(40) }],
    [{ ...run, head_branch: 'feature' }],
    [{ ...run, conclusion: 'failure' }],
    [{ ...run, status: 'in_progress' }],
    [run, { ...run, id: 2, conclusion: 'cancelled' }],
  ]) {
    await assert.rejects(
      verifyReleaseQuality({
        ...input,
        fetcher: async () => ({ ok: true, json: async () => ({ workflow_runs: runs }) }),
      }),
    )
  }
})
