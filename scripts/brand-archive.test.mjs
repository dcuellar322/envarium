import assert from 'node:assert/strict'
import { mkdtemp, mkdir, readFile, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { createRequire } from 'node:module'
import { spawnSync } from 'node:child_process'
import { test } from 'node:test'
import { createPublicBrandArchive } from './brand-archive.mjs'

const { unzipSync, strFromU8 } = createRequire(new URL('../site/package.json', import.meta.url))('fflate')

async function fixture(run) {
  const kit = await mkdtemp(join(tmpdir(), 'envarium-brand-archive-'))
  try {
    for (const directory of ['app', 'interface', 'logos', 'marks', 'tray', 'web', 'references']) await mkdir(join(kit, directory))
    for (const name of ['README.md', 'icons.ts', 'palette.json', 'preview.png', 'preview.svg', 'social.svg']) await writeFile(join(kit, name), name === 'README.md' ? 'Public asset usage guide' : 'asset')
    await writeFile(join(kit, 'logos/logo.svg'), '<svg/>')
    await writeFile(join(kit, 'references/prompt.txt'), 'internal prompt')
    await writeFile(join(kit, 'logos/maintainer-notes.md'), 'internal notes')
    await run(kit)
  } finally { await rm(kit, { recursive: true, force: true }) }
}

test('public kit includes usable assets and licenses while retaining internal references only in source', async () => fixture(async (kit) => {
  const bytes = await createPublicBrandArchive(kit, Buffer.from('license'), Buffer.from('notice'))
  const files = unzipSync(bytes)
  assert.equal(strFromU8(files['envarium-v1/logos/logo.svg']), '<svg/>')
  assert.equal(strFromU8(files['envarium-v1/LICENSE']), 'license')
  assert.equal(strFromU8(files['envarium-v1/NOTICE']), 'notice')
  assert.ok(!Object.keys(files).some((name) => /references|maintainer-notes/.test(name)))
  assert.equal(await readFile(join(kit, 'references/prompt.txt'), 'utf8'), 'internal prompt')
  await writeFile(join(kit, 'references/prompt.txt'), 'changed internal notes')
  assert.deepEqual(await createPublicBrandArchive(kit, Buffer.from('license'), Buffer.from('notice')), bytes)
}))

test('ZIP bytes match across time zones', async () => fixture(async (kit) => {
  const moduleURL = new URL('./brand-archive.mjs', import.meta.url).href
  const code = `import { createPublicBrandArchive } from ${JSON.stringify(moduleURL)}; process.stdout.write(await createPublicBrandArchive(process.argv[1], Buffer.from('license'), Buffer.from('notice')))`
  const results = ['UTC', 'America/Chicago'].map((tz) => spawnSync(process.execPath, ['--input-type=module', '-e', code, kit], { env: { ...process.env, TZ: tz } }))
  for (const result of results) assert.equal(result.status, 0, result.stderr.toString())
  assert.deepEqual(results[0].stdout, results[1].stdout)
}))

test('public README cannot reintroduce internal milestones', async () => fixture(async (kit) => {
  await writeFile(join(kit, 'README.md'), 'Approved in Phase 20')
  await assert.rejects(createPublicBrandArchive(kit, Buffer.from('license'), Buffer.from('notice')), /internal planning/)
}))
